package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ywnaing/term/internal/recipes"
)

var findCmd = &cobra.Command{
	Use:   "find <query>",
	Short: "Search command recipes",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.Join(args, " ")
		all, err := recipes.Load()
		if err != nil {
			return err
		}
		hits := recipes.Search(query, all)
		if len(hits) == 0 {
			fmt.Println("No command recipe found.")
			return nil
		}
		printRecipe(hits[0], recipes.ExtractPort(query))
		return nil
	},
}

func printRecipe(recipe recipes.Recipe, port string) {
	fmt.Println(recipe.Title)
	fmt.Println()
	if len(recipe.Commands["darwin"]) > 0 || len(recipe.Commands["linux"]) > 0 {
		fmt.Println("macOS/Linux:")
		commands := recipe.Commands["darwin"]
		if len(commands) == 0 {
			commands = recipe.Commands["linux"]
		}
		for _, command := range recipes.ReplaceVars(commands, port) {
			fmt.Printf("  %s\n", command)
		}
		fmt.Println()
	}
	if len(recipe.Commands["windows"]) > 0 {
		fmt.Println("Windows:")
		for _, command := range recipes.ReplaceVars(recipe.Commands["windows"], port) {
			fmt.Printf("  %s\n", command)
		}
		fmt.Println()
	}
	if len(recipe.Commands["all"]) > 0 {
		fmt.Printf("%s:\n", recipes.CurrentOSGroup())
		for _, command := range recipes.ReplaceVars(recipe.Commands["all"], port) {
			fmt.Printf("  %s\n", command)
		}
		fmt.Println()
	}
	if len(recipe.Notes) > 0 {
		fmt.Println("Notes:")
		for _, note := range recipe.Notes {
			fmt.Printf("  %s\n", note)
		}
	}
}
