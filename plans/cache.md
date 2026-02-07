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

## Algorithm

On startup:
1. Build cache path: `~/.cache/qa/` + repo path with `/` → `_` + `.yml`
2. Load cache file if exists
3. Prune entries where `last_pass` > 7 days old
4. For each `.qa.yml` directory:
   - Get current hash: `git rev-parse HEAD:<path>`
   - Compare to stored hash
   - If match → skip checks for this directory
   - If miss → run checks
5. On successful check, update hash + timestamp
6. Write cache file

## Notes

- Cache only applies to `checks`, not `format` commands (formatting should always run)
- A directory is only cached after all its checks pass
- Failed checks do not update the cache
