package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"term/internal/output"
)

const version = "0.1.0"

var rootCmd = &cobra.Command{
	Use:           "term",
	Short:         "Developer terminal assistant",
	SilenceUsage:  true,
	SilenceErrors: true,
	Version:       version,
}

func Execute() {
	rootCmd.PersistentFlags().BoolVar(&output.NoColor, "no-color", false, "disable colored output")
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, output.Error(err.Error()))
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(initCmd, listCmd, runCmd, findCmd, explainCmd, recordCmd, historyCmd, hookCmd)
}
