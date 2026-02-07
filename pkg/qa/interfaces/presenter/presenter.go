package presenter

import (
	"github.com/pterm/pterm"

	"github.com/openark-net/qa/pkg/qa/domain"
)

type Presenter struct {
	multi    *pterm.MultiPrinter
	spinners map[string]*pterm.SpinnerPrinter
	done     chan struct{}
}

func New() *Presenter {
	return &Presenter{
		spinners: make(map[string]*pterm.SpinnerPrinter),
		done:     make(chan struct{}),
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
		Start(e.Command.Cmd)
	p.spinners[e.Command.ID()] = spinner
}

func (p *Presenter) handleFinish(e domain.CommandFinished) {
	spinner := p.spinners[e.Result.Command.ID()]
	if spinner == nil {
		return
	}

	if e.Result.State == domain.Completed {
		spinner.MessageStyle = pterm.NewStyle(pterm.FgGreen)
		spinner.SuccessPrinter = &pterm.PrefixPrinter{Prefix: pterm.Prefix{Text: "✓", Style: pterm.NewStyle(pterm.FgGreen)}}
		spinner.Success(e.Result.Command.Cmd)
	} else {
		spinner.MessageStyle = pterm.NewStyle(pterm.FgRed)
		spinner.FailPrinter = &pterm.PrefixPrinter{Prefix: pterm.Prefix{Text: "✗", Style: pterm.NewStyle(pterm.FgRed)}}
		spinner.Fail(e.Result.Command.Cmd)
		p.printFailureOutput(e.Result)
	}
	delete(p.spinners, e.Result.Command.ID())
}

func (p *Presenter) printFailureOutput(result domain.CommandResult) {
	if result.Output == "" {
		return
	}
	pterm.Println()
	pterm.FgRed.Println(result.Output)
}
