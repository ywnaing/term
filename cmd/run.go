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

	"github.com/ywnaing/term/internal/config"
	"github.com/ywnaing/term/internal/executor"
	tmpl "github.com/ywnaing/term/internal/template"
)

var runFlags struct {
	dryRun bool
}

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
		steps, err := resolveRunSteps(shortcut, args[1:])
		if err != nil {
			return err
		}
		if runFlags.dryRun {
			printRunDryRun(name, shortcut, args[1:], steps)
			return nil
		}
		if shortcut.Confirm || strings.EqualFold(shortcut.Danger, "high") {
			if !confirm("This shortcut is marked as dangerous. Continue? y/N ") {
				fmt.Println("Cancelled.")
				return nil
			}
		}
		commands := make([]string, 0, len(steps))
		for _, step := range steps {
			if step.Name != "" {
				fmt.Printf("[%s]\n", step.Name)
			}
			commands = append(commands, step.Command)
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

func init() {
	runCmd.Flags().BoolVar(&runFlags.dryRun, "dry-run", false, "preview commands without running them")
}

func resolveRunSteps(shortcut config.Shortcut, values []string) ([]config.Step, error) {
	steps := make([]config.Step, 0, len(shortcut.Steps))
	for _, step := range shortcut.Steps {
		command, err := tmpl.Apply(step.Command, shortcut.Args, values)
		if err != nil {
			return nil, err
		}
		steps = append(steps, config.Step{Name: step.Name, Command: command})
	}
	return steps, nil
}

func printRunDryRun(name string, shortcut config.Shortcut, values []string, steps []config.Step) {
	fmt.Printf("Shortcut: %s\n", name)
	if shortcut.Parallel {
		fmt.Println("Mode: parallel")
	} else {
		fmt.Println("Mode: sequential")
	}
	if shortcut.Danger != "" {
		fmt.Printf("Danger: %s\n", shortcut.Danger)
	}
	if shortcut.Confirm || strings.EqualFold(shortcut.Danger, "high") {
		fmt.Println("Confirmation required: yes")
	}
	if len(shortcut.Args) > 0 {
		fmt.Println()
		fmt.Println("Args:")
		for i, arg := range shortcut.Args {
			value := ""
			if i < len(values) {
				value = values[i]
			}
			fmt.Printf("  %s = %s\n", arg, value)
		}
	}
	fmt.Println()
	fmt.Println("Steps:")
	for i, step := range steps {
		if step.Name != "" {
			fmt.Printf("  %d. [%s] %s\n", i+1, step.Name, step.Command)
			continue
		}
		fmt.Printf("  %d. %s\n", i+1, step.Command)
	}
}

func confirm(prompt string) bool {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	return line == "y" || line == "yes"
}
