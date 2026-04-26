package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"term/internal/executor"
	"term/internal/history"
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Search and reuse command history",
}

var historySearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search command history",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := openHistory()
		if err != nil {
			return err
		}
		defer store.Close()
		records, err := store.Search(strings.Join(args, " "), 20)
		if err != nil {
			return err
		}
		if len(records) == 0 {
			fmt.Println("No command history found.")
			return nil
		}
		for _, r := range records {
			fmt.Printf("[%d] %s\n", r.ID, r.Command)
			fmt.Printf("    Project: %s\n", r.ProjectName)
			fmt.Printf("    Directory: %s\n", r.Cwd)
			fmt.Printf("    Exit: %d\n", r.ExitCode)
			fmt.Printf("    Time: %s\n\n", r.StartedAt)
		}
		return nil
	},
}

var historyShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show a command history record",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		record, err := getHistoryRecord(args[0])
		if err != nil || record == nil {
			if err != nil {
				return err
			}
			fmt.Println("No command history found.")
			return nil
		}
		fmt.Printf("ID: %d\nCommand: %s\nProject: %s\nDirectory: %s\nExit: %d\nTime: %s\nDuration: %dms\nShell: %s\nOS: %s\n\nStdout:\n%s\n\nStderr:\n%s\n",
			record.ID, record.Command, record.ProjectName, record.Cwd, record.ExitCode, record.StartedAt, record.DurationMS, record.Shell, record.OS, record.Stdout, record.Stderr)
		return nil
	},
}

var historyRunCmd = &cobra.Command{
	Use:   "run <id>",
	Short: "Run a command from history",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		record, err := getHistoryRecord(args[0])
		if err != nil || record == nil {
			if err != nil {
				return err
			}
			fmt.Println("No command history found.")
			return nil
		}
		fmt.Println("Run this command?")
		fmt.Printf("  %s\n\n", record.Command)
		if !confirm("Continue? y/N ") {
			fmt.Println("Cancelled.")
			return nil
		}
		return executor.Runner{Dir: record.Cwd}.RunOne(context.Background(), record.Command)
	},
}

var historyClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear command history",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !confirm("Delete all command history? y/N ") {
			fmt.Println("Cancelled.")
			return nil
		}
		store, err := openHistory()
		if err != nil {
			return err
		}
		defer store.Close()
		if err := store.Clear(); err != nil {
			return err
		}
		fmt.Println("Command history cleared.")
		return nil
	},
}

func init() {
	historyCmd.AddCommand(historySearchCmd, historyShowCmd, historyRunCmd, historyClearCmd)
}

func openHistory() (*history.Store, error) {
	path, err := history.DefaultPath()
	if err != nil {
		return nil, err
	}
	return history.Open(path)
}

func getHistoryRecord(idText string) (*history.Record, error) {
	id, err := strconv.ParseInt(idText, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid history id")
	}
	store, err := openHistory()
	if err != nil {
		return nil, err
	}
	defer store.Close()
	return store.Get(id)
}
