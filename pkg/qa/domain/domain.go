package domain

import "context"

type Command struct {
	Cmd        string
	WorkingDir string
}

func (c Command) ID() string {
	return c.WorkingDir + ":" + c.Cmd
}

type CommandState int

const (
	Completed CommandState = iota
	Failed
)

type Phase int

const (
	PhaseFormat Phase = iota
	PhaseChecks
)

type CommandResult struct {
	Command  Command
	State    CommandState
	Output   string
	ExitCode int
}

type ConfigSet struct {
	Format map[string][]Command
	Checks []Command
}

type ConfigLoader interface {
	Load(rootPath string) (ConfigSet, error)
}

type CommandRunner interface {
	Run(ctx context.Context, cmd Command) CommandResult
}

type Cache interface {
	Hit(cmd Command) bool
	RecordResult(cmd Command, success bool)
	Flush() error
}

type Event interface {
	sealed()
}

type CommandStarted struct {
	Command Command
}

func (CommandStarted) sealed() {}

type CommandFinished struct {
	Result CommandResult
}

func (CommandFinished) sealed() {}

type CommandCached struct {
	Command Command
}

func (CommandCached) sealed() {}

type PhaseCompleted struct {
	Phase   Phase
	Success bool
}

func (PhaseCompleted) sealed() {}
