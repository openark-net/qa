package runner_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/openark-net/qa/pkg/qa/domain"
	"github.com/openark-net/qa/pkg/qa/infrastructure/runner"
)

func TestRunner_Run_Success(t *testing.T) {
	r := runner.New()
	cmd := domain.Command{
		Cmd:        "echo hello",
		WorkingDir: "/tmp",
	}

	result := r.Run(context.Background(), cmd)

	if result.State != domain.Completed {
		t.Errorf("State = %v, want Completed", result.State)
	}
	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", result.ExitCode)
	}
	if got := strings.TrimSpace(result.Output); got != "hello" {
		t.Errorf("Output = %q, want %q", got, "hello")
	}
	if result.Command != cmd {
		t.Errorf("Command not preserved in result")
	}
}

func TestRunner_Run_CapturesStderr(t *testing.T) {
	r := runner.New()
	cmd := domain.Command{
		Cmd:        "echo error >&2",
		WorkingDir: "/tmp",
	}

	result := r.Run(context.Background(), cmd)

	if result.State != domain.Completed {
		t.Errorf("State = %v, want Completed", result.State)
	}
	if !strings.Contains(result.Output, "error") {
		t.Errorf("Output should contain stderr, got %q", result.Output)
	}
}

func TestRunner_Run_RespectsWorkingDir(t *testing.T) {
	r := runner.New()
	cmd := domain.Command{
		Cmd:        "pwd",
		WorkingDir: "/tmp",
	}

	result := r.Run(context.Background(), cmd)

	if result.State != domain.Completed {
		t.Errorf("State = %v, want Completed", result.State)
	}
	if got := strings.TrimSpace(result.Output); got != "/tmp" && got != "/private/tmp" {
		t.Errorf("Output = %q, want /tmp or /private/tmp", got)
	}
}

func TestRunner_Run_Failure(t *testing.T) {
	r := runner.New()
	cmd := domain.Command{
		Cmd:        "exit 42",
		WorkingDir: "/tmp",
	}

	result := r.Run(context.Background(), cmd)

	if result.State != domain.Failed {
		t.Errorf("State = %v, want Failed", result.State)
	}
	if result.ExitCode != 42 {
		t.Errorf("ExitCode = %d, want 42", result.ExitCode)
	}
}

func TestRunner_Run_ContextCancellation(t *testing.T) {
	r := runner.New()
	cmd := domain.Command{
		Cmd:        "sleep 10",
		WorkingDir: "/tmp",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	result := r.Run(ctx, cmd)
	elapsed := time.Since(start)

	if result.State != domain.Failed {
		t.Errorf("State = %v, want Failed", result.State)
	}
	if elapsed > time.Second {
		t.Errorf("Command took %v, expected cancellation within 1s", elapsed)
	}
}
