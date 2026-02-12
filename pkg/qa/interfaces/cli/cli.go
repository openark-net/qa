package cli

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/openark-net/qa/pkg/qa/application"
	"github.com/openark-net/qa/pkg/qa/domain"
	"github.com/openark-net/qa/pkg/qa/infrastructure/cache"
	"github.com/openark-net/qa/pkg/qa/infrastructure/config"
	"github.com/openark-net/qa/pkg/qa/infrastructure/runner"
	"github.com/openark-net/qa/pkg/qa/interfaces/presenter"
)

func Command() *cobra.Command {
	var noCache bool
	var cacheDir string

	cmd := &cobra.Command{
		Use:   "qa",
		Short: "Federated QA runner for monorepos",
		Long: `qa runs format commands and checks defined in .qa.yml files.

Checks are cached using git tree hashes—unchanged code is skipped.
Format commands run sequentially, then checks run in parallel.

Configuration (.qa.yml):
  format:   Commands to run before checks (e.g., formatters)
  checks:   Commands to run in parallel with caching
  includes: Paths to other .qa.yml files to compose`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			configDir, err := config.FindConfig(cwd)
			if err != nil {
				return err
			}

			loader := config.New(os.DirFS(configDir))
			cfg, err := loader.Load(".")
			if err != nil {
				return err
			}

			cfg = resolveWorkingDirs(cfg, configDir)

			var c domain.Cache
			if noCache {
				c = cache.NoOp{}
			} else {
				realCache, err := cache.New(cmd.Context(), cacheDir)
				if err != nil {
					c = cache.NoOp{}
				} else {
					c = realCache
				}
			}

			cmdRunner := runner.New()
			executor := application.New(cmdRunner, c)
			pres := presenter.New()

			go pres.Run(executor.Events())

			success := executor.Run(cmd.Context(), cfg)
			pres.Wait()

			if !success {
				return errors.New("checks failed")
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&noCache, "no-cache", false, "Skip cache, run all checks")
	cmd.Flags().StringVar(&cacheDir, "cache-dir", defaultCacheDir(), "Cache directory")

	return cmd
}

func Run() int {
	if err := Command().Execute(); err != nil {
		return 1
	}
	return 0
}

func defaultCacheDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".cache/qa"
	}
	return filepath.Join(home, ".cache", "qa")
}

func resolveWorkingDirs(cfg domain.ConfigSet, baseDir string) domain.ConfigSet {
	resolved := domain.ConfigSet{
		Format: make(map[string][]domain.Command),
	}

	for dir, cmds := range cfg.Format {
		absDir := filepath.Join(baseDir, dir)
		for _, cmd := range cmds {
			resolved.Format[absDir] = append(resolved.Format[absDir], domain.Command{
				Cmd:        cmd.Cmd,
				WorkingDir: absDir,
			})
		}
	}

	for _, cmd := range cfg.Checks {
		resolved.Checks = append(resolved.Checks, domain.Command{
			Cmd:        cmd.Cmd,
			WorkingDir: filepath.Join(baseDir, cmd.WorkingDir),
		})
	}

	return resolved
}
