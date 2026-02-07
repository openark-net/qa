package presenter

import (
	"fmt"
	"time"

	"github.com/pterm/pterm"

	"github.com/openark-net/qa/pkg/qa/domain"
)

const durationDisplayThreshold = 500 * time.Millisecond

type Presenter struct {
	multi      *pterm.MultiPrinter
	spinners   map[string]*pterm.SpinnerPrinter
	startTimes map[string]time.Time
	done       chan struct{}
}

func New() *Presenter {
	return &Presenter{
		spinners:   make(map[string]*pterm.SpinnerPrinter),
		startTimes: make(map[string]time.Time),
		done:       make(chan struct{}),
	}
}

func (p *Presenter) Run(events <-chan domain.Event) {
	p.multi = &pterm.DefaultMultiPrinter
	p.multi.Start()

	for event := range events {
		switch e := event.(type) {
		case domain.CommandStarted:
			p.handleStart(e)
		case domain.CommandFinished:
			p.handleFinish(e)
		case domain.CommandCached:
			p.handleCached(e)
		}
	}

	p.multi.Stop()
	close(p.done)
}

func (p *Presenter) Wait() {
	<-p.done
}

func (p *Presenter) handleStart(e domain.CommandStarted) {
	spinner, _ := pterm.DefaultSpinner.
		WithWriter(p.multi.NewWriter()).
		WithShowTimer(true).
		Start(e.Command.Cmd)
	p.spinners[e.Command.ID()] = spinner
	p.startTimes[e.Command.ID()] = time.Now()
}

func (p *Presenter) handleFinish(e domain.CommandFinished) {
	cmdID := e.Result.Command.ID()
	spinner := p.spinners[cmdID]
	if spinner == nil {
		return
	}

	duration := time.Since(p.startTimes[cmdID])
	message := p.formatCompletionMessage(e.Result.Command.Cmd, duration)

	if e.Result.State == domain.Completed {
		spinner.MessageStyle = pterm.NewStyle(pterm.FgGreen)
		spinner.SuccessPrinter = &pterm.PrefixPrinter{Prefix: pterm.Prefix{Text: "✓", Style: pterm.NewStyle(pterm.FgGreen)}}
		spinner.Success(message)
	} else {
		spinner.MessageStyle = pterm.NewStyle(pterm.FgRed)
		spinner.FailPrinter = &pterm.PrefixPrinter{Prefix: pterm.Prefix{Text: "✗", Style: pterm.NewStyle(pterm.FgRed)}}
		spinner.Fail(message)
		p.printFailureOutput(e.Result)
	}
	delete(p.spinners, cmdID)
	delete(p.startTimes, cmdID)
}

func (p *Presenter) formatCompletionMessage(cmd string, duration time.Duration) string {
	if duration < durationDisplayThreshold {
		return cmd
	}
	durationText := pterm.FgGray.Sprint(formatDuration(duration))
	return fmt.Sprintf("%s %s", cmd, durationText)
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
}

func (p *Presenter) handleCached(e domain.CommandCached) {
	writer := p.multi.NewWriter()
	fmt.Fprintln(writer, pterm.FgGray.Sprintf("○ %s (cached)", e.Command.Cmd))
}

func (p *Presenter) printFailureOutput(result domain.CommandResult) {
	if result.Output == "" {
		return
	}
	pterm.Println()
	pterm.FgRed.Println(result.Output)
}
