# Caching Design

## Overview

Skip running checks for directories that haven't changed since the last successful run.

## Git Tree Hashes

Git stores a tree object for every directory in every commit. Each tree has a SHA computed from the directory's contents.

```bash
git rev-parse HEAD:app
# => 4b825dc642cb6eb9a060e54bf8d69288fbee4904
```

Key property: **same content = same hash, regardless of branch**. This means:
- Switching branches automatically invalidates cache if content differs
- Two branches with identical directories share the same cache entry

## Dirty Working Directory Check

The git tree hash only reflects committed state. We must also verify the working directory has no uncommitted changes of any kind.

```bash
git status --porcelain <path>
# Empty output = clean, any output = dirty
```

A directory is considered cache-valid only if:
1. `git status --porcelain <path>` returns empty (no untracked, modified, staged, or deleted files)
2. Tree hash matches cached hash

This is strict: any uncommitted work invalidates the cache. The only way to skip checks is when a directory is fully committed and matches the cached state.

## Cache Location

Flat file structure in configurable cache directory (default: `~/.cache/qa/`):

```
~/.cache/qa/
├── _Users_fargus_dv_oa_qa.yml
├── _Users_fargus_dv_oa_other-repo.yml
└── _Users_fargus_work_bigproject.yml
```

Filename is the repo's absolute path with `/` replaced by `_`.

### Configuration

Cache directory can be configured via:
```
qa --cache-dir /custom/path    # Override cache directory
```

Default: `~/.cache/qa/`

## Cache File Format

Keys are paths relative to the git root:

```yaml
# ~/.cache/qa/_Users_fargus_dv_oa_qa.yml
app:
  hash: "abc123..."
  last_pass: 2024-02-06T15:30:00Z
api:
  hash: "def456..."
  last_pass: 2024-02-06T15:30:00Z
.:
  hash: "789xyz..."
  last_pass: 2024-02-06T15:30:00Z
```

Note: `.` represents the repo root directory.

## Interface

```go
// Defined in pkg/qa/domain/domain.go (already exists)
type Cache interface {
    Hit(cmd Command) bool
    RecordResult(cmd Command, success bool)
    Flush() error
}

type CommandCached struct {
    Command Command
}
```

- `Hit`: converts `cmd.WorkingDir` to relative path, returns true if hash matches cached AND working dir is clean
- `RecordResult`: tracks success/failure per directory (converts to relative path, groups internally)
- `Flush`: persists entries where ALL commands for that directory succeeded
- `CommandCached`: event emitted when a command is skipped due to cache hit (already defined in domain)

## Algorithm

### Cache Creation

1. Build cache path: `~/.cache/qa/` + repo path with `/` → `_` + `.yml`
2. Load cache file if exists
3. Prune entries where `last_pass` > 7 days old
4. Compute current git tree hashes for relevant directories

### Executor Integration

Update executor to require cache as a dependency:

```go
// pkg/qa/application/executor.go
func New(runner domain.CommandRunner, cache domain.Cache) *Executor
```

Cache is required (not nil). CLI wires either real Cache or NoOp based on `--no-cache` flag.

For each check command:
```
if cache.Hit(cmd) {
    emit CommandCached{Command: cmd}
    continue
}
emit CommandStarted{Command: cmd}
result := runner.Run(ctx, cmd)
emit CommandFinished{Result: result}
cache.RecordResult(cmd, result.State == Completed)
```

After checks phase completes:
```
cache.Flush()
```

### Cache Implementation

- `Hit`: Return true only if `git status --porcelain <path>` is empty AND tree hash matches stored hash
- `RecordResult`: Track per-directory success (all commands must pass)
- `Flush`: Write cache file, only update entries where all commands succeeded

## Notes

- Cache only applies to `checks`, not `format` commands (formatting should always run)
- A directory is only cached after all its checks pass
- Failed checks do not update the cache

## CLI

```
qa --no-cache    # Skip cache, run all checks
```

When `--no-cache` is set, inject a no-op cache implementation that always returns `Hit() = false` and discards results.

## Architecture

Flat structure within `pkg/qa/infrastructure/cache/`:

```
pkg/qa/infrastructure/cache/
├── cache.go      # Cache struct implementing domain.Cache
├── noop.go       # NoOp struct implementing domain.Cache
├── git.go        # git operations
└── storage.go    # file I/O
```

### cache.go

Main implementation of `domain.Cache`:

```go
type Cache struct {
    git      *GitClient
    storage  *Storage
    cacheDir string
    repoRoot string
    data     map[string]Entry  // keyed by relative path from git root
    results  map[string]bool   // tracks per-directory success this run
}

type Entry struct {
    Hash     string
    LastPass time.Time
}
```

- `New(cacheDir string) (*Cache, error)` - returns error if not in git repo
- `Hit(cmd Command) bool` - converts WorkingDir to relative path, checks dirty + hash match
- `RecordResult(cmd Command, success bool)` - tracks per-directory results (relative path)
- `Flush() error` - persists only directories where ALL commands passed
- 7-day TTL pruning on load

### noop.go

```go
type NoOp struct{}

func (NoOp) Hit(cmd Command) bool { return false }
func (NoOp) RecordResult(cmd Command, success bool) {}
func (NoOp) Flush() error { return nil }
```

### git.go

```go
type GitClient struct{}

func (g *GitClient) RepoRoot() (string, error)
func (g *GitClient) TreeHash(relativePath string) (string, error)  // path relative to repo root
func (g *GitClient) IsDirty(relativePath string) (bool, error)     // path relative to repo root
func (g *GitClient) ToRelative(absolutePath string) (string, error) // converts absolute to relative
```

**Critical**: All paths passed to git commands must be relative to the repo root. The cache is responsible for converting `cmd.WorkingDir` (absolute) to a relative path before any git operations or cache key lookups.

### storage.go

```go
type Storage struct{}

func (s *Storage) Load(cacheDir, repoRoot string) (map[string]Entry, error)
func (s *Storage) Save(cacheDir, repoRoot string, data map[string]Entry) error
```

Path conversion for filename: `/Users/foo/repo` → `_Users_foo_repo.yml`

## Files to Modify

- `pkg/qa/domain/domain.go` - Cache interface already defined, no changes needed
- `pkg/qa/application/executor.go` - add cache dependency (required, not nil), filter hits in runChecks
- `pkg/qa/interfaces/cli/cli.go` - add `--no-cache` and `--cache-dir` flags, wire up cache
- `pkg/qa/interfaces/presenter/presenter.go` - handle `CommandCached` event in the event loop
