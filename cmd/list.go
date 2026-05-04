package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"

	"github.com/ywnaing/term/internal/config"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List project shortcuts",
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
		fmt.Println("Available shortcuts:")
		fmt.Println()
		keys := make([]string, 0, len(cfg.Shortcuts))
		for key := range cfg.Shortcuts {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			fmt.Printf("%-10s %s\n", key, cfg.Shortcuts[key].Description)
		}
		return nil
	},
}
