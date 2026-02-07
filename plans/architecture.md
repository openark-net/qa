# Architecture

## Overview

CLI application using clean architecture pattern with Cobra.

## Directory Structure

```
pkg/
├── cli/
│   └── main.go           # Cobra CLI entry point
└── qa/
    ├── domain/           # Pure Go, no dependencies, core types and interfaces
    ├── application/      # Orchestrates execution, emits events
    ├── infrastructure/   # Config parsing, command runner
    └── interfaces/       # CLI rendering, event subscription
```

## Domain Types

### Command

```go
type Command struct {
    Cmd        string
    WorkingDir string
}

func (c Command) ID() string {
    return c.WorkingDir + ":" + c.Cmd
}
```

### CommandState

```go
type CommandState int

const (
    Pending CommandState = iota
    Running
    Completed
    Failed
)
```

### CommandResult

```go
type CommandResult struct {
    Command  Command
    State    CommandState
    Output   string // stdout + stderr, only shown on failure
    ExitCode int
}
```

### ConfigSet

```go
type ConfigSet struct {
    Format map[string][]Command // key is directory path, commands run sequentially
    Checks []Command            // all checks run in parallel
}
```

### Interfaces

```go
type ConfigLoader interface {
    Load(rootPath string) (ConfigSet, error)
}

type CommandRunner interface {
    Run(ctx context.Context, cmd Command) CommandResult
}
```

### Events

```go
type Event interface{ sealed() }

type CommandStarted struct{ Command Command }
type CommandFinished struct{ Result CommandResult }
type PhaseCompleted struct{ Phase string; Success bool }
```

## Execution Semantics

| Phase  | Parallelism                                      | On Failure                              |
|--------|--------------------------------------------------|-----------------------------------------|
| Format | Parallel across directories, sequential within   | Wait for all to finish, then exit       |
| Checks | Fully parallel                                   | Wait for all to finish, report all      |

### Format Phase Example

Given `/api` with `go fmt` and `/app` with `eslint`, `prettier`:

- Goroutine 1: `/api` runs `go fmt`
- Goroutine 2: `/app` runs `eslint` → waits → runs `prettier`

Both goroutines run concurrently. Within `/app`, commands are sequential.

## Layer Responsibilities

| Package                  | Responsibility                                           |
|--------------------------|----------------------------------------------------------|
| `domain`                 | `Command`, `CommandState`, `CommandResult`, `Event`      |
| `infrastructure/config`  | Find `.qa.yml` files, parse YAML, return commands        |
| `infrastructure/runner`  | Execute command, buffer output, return `CommandResult`   |
| `application`            | Orchestrate phases, manage parallelism, emit events      |
| `interfaces/cli`         | Subscribe to events, render spinners and status          |

## CLI Display

One line per command with status indicator:

```
✓ /app:eslint
⠋ /app:npm run test
⠙ /api:go test ./...
✗ /api:go vet
```

- Spinner (braille) for running commands
- Check emoji for completed
- Cross emoji for failed
- Output only shown on failure

## High-Level Flow

### Phase 1 (MVP)

1. Find and parse all config files
2. Run format commands (sequential per-directory, parallel across directories)
3. If any format failed, exit without running checks
4. Run all check commands in parallel
5. Wait for all checks, report all failures

### Phase 2 (Caching)

1. Find and parse all config files
2. Load cache, compute current hashes for each directory
3. Determine which directories need checks (cache miss or dirty)
4. Run format commands
5. Run check commands (only for uncached directories)
6. Update cache for directories where all checks passed
