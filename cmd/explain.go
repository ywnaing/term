package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"term/internal/explain"
	"term/internal/history"
)

var explainCmd = &cobra.Command{
	Use:   "explain [errorText|last]",
	Short: "Explain common terminal errors",
	RunE: func(cmd *cobra.Command, args []string) error {
		text, err := explainInput(args)
		if err != nil {
			return err
		}
		rules, err := explain.Load()
		if err != nil {
			return err
		}
		match, err := explain.Find(text, rules)
		if err != nil {
			return err
		}
		if match == nil {
			fmt.Println("No known explanation found.")
			fmt.Println("Try:")
			fmt.Println(`  term find "<keyword>"`)
			return nil
		}
		printExplanation(*match)
		return nil
	},
}

func explainInput(args []string) (string, error) {
	if len(args) > 0 && args[0] == "last" {
		path, err := history.DefaultPath()
		if err != nil {
			return "", err
		}
		store, err := history.Open(path)
		if err != nil {
			return "", err
		}
		defer store.Close()
		record, err := store.LatestFailed()
		if err != nil {
			return "", err
		}
		if record == nil {
			return "", fmt.Errorf("no failed command history found")
		}
		return record.Stderr, nil
	}
	if len(args) > 0 {
		return strings.Join(args, " "), nil
	}
	stat, _ := os.Stdin.Stat()
	if stat != nil && (stat.Mode()&os.ModeCharDevice) == 0 {
		data, err := io.ReadAll(os.Stdin)
		return string(data), err
	}
	return "", fmt.Errorf("provide error text or pipe it into term explain")
}

func printExplanation(match explain.Match) {
	fmt.Printf("Detected: %s\n\n", match.Rule.Title)
	fmt.Println("Meaning:")
	fmt.Println(explain.RenderTemplate(match.Rule.Meaning, match.Port))
	fmt.Println()
	fmt.Println("Fix:")
	fmt.Println()
	if len(match.Rule.Fixes["darwin"]) > 0 || len(match.Rule.Fixes["linux"]) > 0 {
		fmt.Println("macOS/Linux:")
		commands := match.Rule.Fixes["darwin"]
		if len(commands) == 0 {
			commands = match.Rule.Fixes["linux"]
		}
		for _, command := range explain.RenderCommands(commands, match.Port) {
			fmt.Printf("  %s\n", command)
		}
		fmt.Println()
	}
	if len(match.Rule.Fixes["windows"]) > 0 {
		fmt.Println("Windows:")
		for _, command := range explain.RenderCommands(match.Rule.Fixes["windows"], match.Port) {
			fmt.Printf("  %s\n", command)
		}
		fmt.Println()
	}
	if len(match.Rule.Fixes["all"]) > 0 {
		for _, command := range explain.RenderCommands(match.Rule.Fixes["all"], match.Port) {
			fmt.Printf("  %s\n", command)
		}
		fmt.Println()
	}
	if len(match.Rule.Notes) > 0 {
		fmt.Println("Notes:")
		for _, note := range match.Rule.Notes {
			fmt.Printf("  %s\n", explain.RenderTemplate(note, match.Port))
		}
	}
}
