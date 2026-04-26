package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"term/internal/config"
	"term/internal/executor"
	tmpl "term/internal/template"
)

var runCmd = &cobra.Command{
	Use:   "run <shortcut> [args...]",
	Short: "Run a project shortcut",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, _ := os.Getwd()
		cfg, err := config.Load(cwd)
		if os.IsNotExist(err) {
			fmt.Println("No .term.yml found.")
			fmt.Println("Run:")
			fmt.Println("  term init")
			return nil
		}
		if err != nil {
			return err
		}
		name := args[0]
		shortcut, ok := cfg.Shortcuts[name]
		if !ok {
			fmt.Printf("Unknown shortcut: %s\n\n", name)
			fmt.Println("Available shortcuts:")
			keys := make([]string, 0, len(cfg.Shortcuts))
			for key := range cfg.Shortcuts {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for _, key := range keys {
				fmt.Printf("  %s\n", key)
			}
			return fmt.Errorf("shortcut not found")
		}
		if shortcut.Confirm || strings.EqualFold(shortcut.Danger, "high") {
			if !confirm("This shortcut is marked as dangerous. Continue? y/N ") {
				fmt.Println("Cancelled.")
				return nil
			}
		}
		commands := make([]string, 0, len(shortcut.Steps))
		for _, step := range shortcut.Steps {
			command, err := tmpl.Apply(step.Command, shortcut.Args, args[1:])
			if err != nil {
				return err
			}
			if step.Name != "" {
				fmt.Printf("[%s]\n", step.Name)
			}
			commands = append(commands, command)
		}
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		runner := executor.Runner{Dir: cwd}
		if shortcut.Parallel {
			return runner.RunParallel(ctx, commands)
		}
		return runner.RunSequential(ctx, commands)
	},
}

func confirm(prompt string) bool {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	return line == "y" || line == "yes"
}
