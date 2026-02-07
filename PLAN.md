# QA Tool - YAML Schema Plan

## Overview

Replace copy-pasted bash QA scripts with a unified Go tool that reads `.qa.yml` configuration files. Supports mono-repos with multiple languages/checks.

## YAML Schema

### Top-Level `.qa.yml` (Orchestrator)

References child QA files and optionally runs its own format commands:

```yaml
# /.qa.yml
includes:
  - app/.qa.yml
  - api/.qa.yml
  - website/.qa.yml

format:
  - terraform fmt ./infra
```

### Child `.qa.yml` (Leaf)

Defines format commands and QA checks for a specific directory:

```yaml
# /app/.qa.yml
format:
  - npm run prettier_fix

checks:
  - npm run lint_fix
  - npm run test --ci
```

```yaml
# /api/.qa.yml
format:
  - go fmt ./...

checks:
  - go vet ./...
  - go test ./...
```

## Execution Order

1. **Format phase**: All `format` commands run first (sequentially, in order listed)
2. **Checks phase**: All `checks` run in parallel after formatting completes

**Important**: All commands in a `.qa.yml` file run with the working directory set to that file's parent directory. For example, commands in `/app/.qa.yml` run from `/app/`.

For files with `includes`:
1. Run own `format` commands first (in the current directory)
2. Then process each included file (each runs in its own directory)

## Schema Definition

```yaml
# Optional: list of relative paths to other .qa.yml files
includes:
  - path/to/.qa.yml

# Optional: format commands run sequentially before checks
format:
  - "command string"

# Optional: QA checks run in parallel
checks:
  - "command string"
```

## Key Design Decisions

1. **Simple string commands** - No complex command objects, just strings
2. **Format runs first** - Ensures code is formatted before linting/testing
3. **Checks run in parallel** - Faster execution for independent checks
4. **Relative paths** - All paths relative to the `.qa.yml` file location
5. **No fast mode** - Run all checks every time
6. **Commands run from file's directory** - Each `.qa.yml`'s commands execute with cwd set to that file's parent directory (e.g., `/api/.qa.yml` commands run from `/api/`)

## Example Mono-Repo Structure

```
/
├── .qa.yml              # includes: [app, api, website]
├── app/
│   └── .qa.yml          # npm checks
├── api/
│   └── .qa.yml          # go checks
└── website/
    └── .qa.yml          # next.js checks
```

## Output

- Show which commands are running
- Show OK/ERR status for each command
- Exit with non-zero if any check fails
- Aggregate all errors and show at end
