package application

import (
	"context"
	"sync"

	"github.com/openark-net/qa/pkg/qa/domain"
)

type Executor struct {
	runner   domain.CommandRunner
	eventsCh chan domain.Event
}

func New(runner domain.CommandRunner) *Executor {
	return &Executor{
		runner:   runner,
		eventsCh: make(chan domain.Event, 100),
	}
}

func (e *Executor) Events() <-chan domain.Event {
	return e.eventsCh
}

func (e *Executor) Run(ctx context.Context, cfg domain.ConfigSet) bool {
	formatSuccess := e.runFormat(ctx, cfg.Format)
	e.eventsCh <- domain.PhaseCompleted{Phase: domain.PhaseFormat, Success: formatSuccess}

	if !formatSuccess {
		close(e.eventsCh)
		return false
	}

	checksSuccess := e.runChecks(ctx, cfg.Checks)
	e.eventsCh <- domain.PhaseCompleted{Phase: domain.PhaseChecks, Success: checksSuccess}

	close(e.eventsCh)
	return checksSuccess
}

func (e *Executor) runFormat(ctx context.Context, formatCmds map[string][]domain.Command) bool {
	if len(formatCmds) == 0 {
		return true
	}

	var wg sync.WaitGroup
	results := make(chan bool, len(formatCmds))

	for _, cmds := range formatCmds {
		wg.Add(1)
		go func(commands []domain.Command) {
			defer wg.Done()
			results <- e.runSequential(ctx, commands)
		}(cmds)
	}

	wg.Wait()
	close(results)

	for success := range results {
		if !success {
			return false
		}
	}
	return true
}

func (e *Executor) runSequential(ctx context.Context, cmds []domain.Command) bool {
	for _, cmd := range cmds {
		e.eventsCh <- domain.CommandStarted{Command: cmd}
		result := e.runner.Run(ctx, cmd)
		e.eventsCh <- domain.CommandFinished{Result: result}

		if result.State == domain.Failed {
			return false
		}
	}
	return true
}

func (e *Executor) runChecks(ctx context.Context, checks []domain.Command) bool {
	if len(checks) == 0 {
		return true
	}

	var wg sync.WaitGroup
	results := make(chan bool, len(checks))

	for _, cmd := range checks {
		wg.Add(1)
		go func(c domain.Command) {
			defer wg.Done()
			e.eventsCh <- domain.CommandStarted{Command: c}
			result := e.runner.Run(ctx, c)
			e.eventsCh <- domain.CommandFinished{Result: result}
			results <- result.State == domain.Completed
		}(cmd)
	}

	wg.Wait()
	close(results)

	for success := range results {
		if !success {
			return false
		}
	}
	return true
}
