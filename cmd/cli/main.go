package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/openark-net/qa"
	"github.com/openark-net/qa/pkg/init"
	"github.com/openark-net/qa/pkg/qa/interfaces/cli"
)

func main() {
	rootCmd := cli.Command()

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize project tooling",
	}

	hookCmd := &cobra.Command{
		Use:   "hook",
		Short: "Install pre-commit hook that runs qa",
		RunE: func(cmd *cobra.Command, args []string) error {
			return setup.InstallHook()
		},
	}

	expectationsCmd := &cobra.Command{
		Use:   "expectations",
		Short: "Create CLAUDE.md with code quality expectations",
		RunE: func(cmd *cobra.Command, args []string) error {
			dest := "./CLAUDE.md"
			if len(args) > 0 {
				dest = args[0]
			}
			return setup.CopyExpectations(dest)
		},
	}

	readmeCmd := &cobra.Command{
		Use:   "readme",
		Short: "Print the README",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print(qa.ReadmeContent)
		},
	}

	initCmd.AddCommand(hookCmd, expectationsCmd)
	rootCmd.AddCommand(initCmd, readmeCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
