package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var hookCmd = &cobra.Command{
	Use:   "hook",
	Short: "Shell hook helpers",
}

var hookInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Print experimental shell hook snippet",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(`# term shell hook (experimental)
# Add this to ~/.zshrc or ~/.bashrc after reviewing it.
# Future versions will record commands automatically, enabling:
# - term explain last
# - term history search <query>
# - richer project command history
#
# MVP placeholder:
# term record --command "<command>" --exit-code <code> --stderr "<stderr>"
`)
	},
}

func init() {
	hookCmd.AddCommand(hookInstallCmd)
}
