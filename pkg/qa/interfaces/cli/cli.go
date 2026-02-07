package cli

import (
	"errors"
	"os"

	"github.com/spf13/cobra"

	"github.com/openark-net/qa/pkg/qa/application"
	"github.com/openark-net/qa/pkg/qa/infrastructure/config"
	"github.com/openark-net/qa/pkg/qa/infrastructure/runner"
	"github.com/openark-net/qa/pkg/qa/interfaces/presenter"
)

func Run() int {
	loader := config.New(os.DirFS("."))
	cmdRunner := runner.New()
	executor := application.New(cmdRunner)
	pres := presenter.New()

	cmd := &cobra.Command{
		Use:   "qa",
		Short: "Run QA checks from .qa.yml",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loader.Load(".")
			if err != nil {
				return err
			}

			go pres.Run(executor.Events())

			success := executor.Run(cmd.Context(), cfg)
			pres.Wait()

			if !success {
				return errors.New("checks failed")
			}
			return nil
		},
	}

	if err := cmd.Execute(); err != nil {
		return 1
	}
	return 0
}
