package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"term/internal/config"
	"term/internal/history"
)

var recordFlags struct {
	command    string
	exitCode   int
	stdout     string
	stderr     string
	durationMS int64
	quiet      bool
}

var recordCmd = &cobra.Command{
	Use:   "record",
	Short: "Record a command in history",
	RunE: func(cmd *cobra.Command, args []string) error {
		if recordFlags.command == "" {
			return fmt.Errorf("--command is required")
		}
		if history.ShouldSkipCommand(recordFlags.command) {
			return nil
		}
		cwd, _ := os.Getwd()
		project := config.ProjectNameFromDir(cwd)
		if _, cfg, err := config.FindNearest(cwd); err == nil {
			if !cfg.History.IsEnabled() {
				if !recordFlags.quiet {
					fmt.Println("Command history is disabled for this project.")
				}
				return nil
			}
			if cfg.Project != "" {
				project = cfg.Project
			}
		}
		path, err := history.DefaultPath()
		if err != nil {
			return err
		}
		store, err := history.Open(path)
		if err != nil {
			return err
		}
		defer store.Close()
		record := history.NewRecord(
			history.Redact(recordFlags.command),
			recordFlags.exitCode,
			history.Redact(recordFlags.stdout),
			history.Redact(recordFlags.stderr),
			recordFlags.durationMS,
			cwd,
			project,
		)
		id, err := store.Insert(record)
		if err != nil {
			return err
		}
		if !recordFlags.quiet {
			fmt.Printf("Recorded command history: %d\n", id)
		}
		return nil
	},
}

func init() {
	recordCmd.Flags().StringVar(&recordFlags.command, "command", "", "command text")
	recordCmd.Flags().IntVar(&recordFlags.exitCode, "exit-code", 0, "command exit code")
	recordCmd.Flags().StringVar(&recordFlags.stdout, "stdout", "", "command stdout")
	recordCmd.Flags().StringVar(&recordFlags.stderr, "stderr", "", "command stderr")
	recordCmd.Flags().Int64Var(&recordFlags.durationMS, "duration-ms", 0, "command duration in milliseconds")
	recordCmd.Flags().BoolVar(&recordFlags.quiet, "quiet", false, "suppress output")
}
