package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"term/internal/config"
	"term/internal/detect"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create .term.yml",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, _ := os.Getwd()
		path := filepath.Join(cwd, config.FileName)
		if _, err := os.Stat(path); err == nil {
			fmt.Println(".term.yml already exists.")
			return nil
		}
		cfg := detect.DefaultConfig(cwd)
		if err := config.Write(cwd, cfg); err != nil {
			return err
		}
		fmt.Println("Created .term.yml")
		return nil
	},
}
