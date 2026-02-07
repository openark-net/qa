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

```yaml
# ~/.cache/qa/_Users_fargus_dv_oa_qa.yml
app:
  hash: "abc123..."
  last_pass: 2024-02-06T15:30:00Z
api:
  hash: "def456..."
  last_pass: 2024-02-06T15:30:00Z
```

## Interface

```go
type Cache interface {
    Hit(cmd Command) bool
    RecordResult(cmd Command, success bool)
    Flush() error
}

type CommandCached struct {
    Command Command
}
```

- `Hit`: returns true if `cmd.WorkingDir` hash matches cached AND working dir is clean
- `RecordResult`: tracks success/failure per directory (groups by WorkingDir internally)
- `Flush`: persists entries where ALL commands for that directory succeeded
- `CommandCached`: event emitted when a command is skipped due to cache hit

## Algorithm

### Cache Creation

1. Build cache path: `~/.cache/qa/` + repo path with `/` → `_` + `.yml`
2. Load cache file if exists
3. Prune entries where `last_pass` > 7 days old
4. Compute current git tree hashes for relevant directories

### Executor Integration

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

Cache module follows layered architecture within `pkg/qa/infrastructure/cache/`:

```
pkg/qa/infrastructure/cache/
├── domain/           # cache-specific domain types
├── application/      # orchestration logic
├── infrastructure/   # git commands, file I/O
└── interfaces/       # exports Cache + NoOp structs
```

Git operations and file I/O are infrastructure concerns. The complex logic to determine cache validity needs proper separation.

### domain/types.go

Types internal to the cache subsystem:
- `Entry` struct: `{Hash string, LastPass time.Time}`
- `CacheData` map: `map[string]Entry` (dir path -> entry)

### infrastructure/git.go

Git operations:
- `GetRepoRoot() (string, error)` - `git rev-parse --show-toplevel`
- `GetTreeHash(path string) (string, error)` - `git rev-parse HEAD:<path>`
- `IsDirty(path string) (bool, error)` - `git status --porcelain <path>`

### infrastructure/storage.go

Cache file I/O:
- `Load(cacheDir, repoRoot string) (CacheData, error)` - read YAML
- `Save(cacheDir, repoRoot string, data CacheData) error` - write YAML
- Path conversion: `/Users/foo/repo` → `_Users_foo_repo.yml`

### application/cache.go

Orchestration logic:
- `Hit(cmd Command) bool` - combines dirty check + hash comparison
- `RecordResult(cmd Command, success bool)` - tracks per-directory results
- `Flush() error` - persists only directories where ALL commands passed
- 7-day TTL pruning on load

### interfaces/service.go

Exports structs implementing `domain.Cache`:
- `Service` struct - real implementation, wraps application layer
- `NoOp` struct - always returns `Hit()=false`, discards results

## Files to Modify

- `pkg/qa/application/executor.go` - add cache dependency, filter hits in runChecks
- `pkg/qa/interfaces/cli/cli.go` - add `--no-cache` and `--cache-dir` flags, wire up cache
