package runner

import (
	"bytes"
	"context"
	"os/exec"

	"github.com/openark-net/qa/pkg/qa/domain"
)

type Runner struct{}

func New() *Runner {
	return &Runner{}
}

// NOTE: Hardcoded to sh. Will break on Windows.
func (r *Runner) Run(ctx context.Context, cmd domain.Command) domain.CommandResult {
	shellCmd := exec.CommandContext(ctx, "sh", "-c", cmd.Cmd)
	shellCmd.Dir = cmd.WorkingDir

	var output bytes.Buffer
	shellCmd.Stdout = &output
	shellCmd.Stderr = &output

	err := shellCmd.Run()

	result := domain.CommandResult{
		Command: cmd,
		Output:  output.String(),
	}

	if err != nil {
		result.State = domain.Failed
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = 1
		}
		return result
	}

	result.State = domain.Completed
	result.ExitCode = 0
	return result
}
