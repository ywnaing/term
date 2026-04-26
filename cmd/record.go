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
}

var recordCmd = &cobra.Command{
	Use:   "record",
	Short: "Record a command in history",
	RunE: func(cmd *cobra.Command, args []string) error {
		if recordFlags.command == "" {
			return fmt.Errorf("--command is required")
		}
		cwd, _ := os.Getwd()
		project := config.ProjectNameFromDir(cwd)
		if _, cfg, err := config.FindNearest(cwd); err == nil && cfg.Project != "" {
			project = cfg.Project
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
		record := history.NewRecord(recordFlags.command, recordFlags.exitCode, recordFlags.stdout, recordFlags.stderr, recordFlags.durationMS, cwd, project)
		id, err := store.Insert(record)
		if err != nil {
			return err
		}
		fmt.Printf("Recorded command history: %d\n", id)
		return nil
	},
}

func init() {
	recordCmd.Flags().StringVar(&recordFlags.command, "command", "", "command text")
	recordCmd.Flags().IntVar(&recordFlags.exitCode, "exit-code", 0, "command exit code")
	recordCmd.Flags().StringVar(&recordFlags.stdout, "stdout", "", "command stdout")
	recordCmd.Flags().StringVar(&recordFlags.stderr, "stderr", "", "command stderr")
	recordCmd.Flags().Int64Var(&recordFlags.durationMS, "duration-ms", 0, "command duration in milliseconds")
}
