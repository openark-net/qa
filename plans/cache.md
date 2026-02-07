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

The git tree hash only reflects committed state. Uncommitted changes won't invalidate the cache on their own.

```bash
git diff --quiet <path>
# Exit code 0 = clean, 1 = dirty
```

A directory is considered cache-valid only if:
1. Tree hash matches cached hash, AND
2. Working directory is clean (`git diff --quiet <path>` returns 0)

> **Note**: This approach needs investigation. Consider edge cases like untracked files, staged but uncommitted changes, etc.

## Cache Location

Flat file structure in user's cache directory:

```
~/.cache/qa/
├── _Users_fargus_dv_oa_qa.yml
├── _Users_fargus_dv_oa_other-repo.yml
└── _Users_fargus_work_bigproject.yml
```

Filename is the repo's absolute path with `/` replaced by `_`.

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

- `Hit`: Compare `cmd.WorkingDir` hash to stored hash, verify `git diff --quiet` passes
- `RecordResult`: Track per-directory success (all commands must pass)
- `Flush`: Write cache file, only update entries where all commands succeeded

## Notes

- Cache only applies to `checks`, not `format` commands (formatting should always run)
- A directory is only cached after all its checks pass
- Failed checks do not update the cache
