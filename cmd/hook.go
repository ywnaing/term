package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var hookCmd = &cobra.Command{
	Use:   "hook",
	Short: "Shell hook helpers",
}

var hookInstallCmd = &cobra.Command{
	Use:   "install [zsh|bash]",
	Short: "Print experimental shell hook snippet",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		shell := ""
		if len(args) > 0 {
			shell = strings.ToLower(args[0])
		} else {
			shell = detectShell()
		}
		switch shell {
		case "zsh":
			_, _ = os.Stdout.WriteString(zshHookSnippet)
		case "bash":
			_, _ = os.Stdout.WriteString(bashHookSnippet)
		default:
			return fmt.Errorf("unsupported shell: %s\nRun:\n  term hook install zsh\n  term hook install bash", shell)
		}
		return nil
	},
}

const zshHookSnippet = `# term shell hook (experimental)
# Add this to ~/.zshrc after reviewing it.
# Records command text, exit code, cwd, project, timestamp, shell, OS, and duration.
# It does not capture stdout or stderr.
# Commands that start with a space are skipped.
#
# - term history search <query>
# - term explain last when stderr is recorded manually
#
autoload -Uz add-zsh-hook
zmodload zsh/datetime 2>/dev/null || true

__term_last_command=""
__term_started_at=0
__term_recording=0

__term_preexec() {
  [[ $__term_recording -eq 1 ]] && return
  __term_last_command="$1"
  __term_started_at=${EPOCHSECONDS:-$(date +%s)}
}

__term_precmd() {
  local exit_code=$?
  [[ $__term_recording -eq 1 ]] && return
  [[ -z "$__term_last_command" ]] && return
  [[ "$__term_last_command" == " "* ]] && { __term_last_command=""; return; }

  local now=${EPOCHSECONDS:-$(date +%s)}
  local duration_ms=$(( (now - __term_started_at) * 1000 ))
  __term_recording=1
  term record --quiet --command "$__term_last_command" --exit-code "$exit_code" --duration-ms "$duration_ms" >/dev/null 2>&1
  __term_recording=0
  __term_last_command=""
}

add-zsh-hook preexec __term_preexec
add-zsh-hook precmd __term_precmd
`

const bashHookSnippet = `# term shell hook (experimental)
# Add this to ~/.bashrc after reviewing it.
# Records command text, exit code, cwd, project, timestamp, shell, OS, and duration.
# It does not capture stdout or stderr.
# Commands that start with a space are skipped.
#
# - term history search <query>
# - term explain last when stderr is recorded manually
#
__term_last_command=""
__term_started_at=0
__term_recording=0

__term_debug_trap() {
  [[ $__term_recording -eq 1 ]] && return
  case "$BASH_COMMAND" in
    __term_*|term\ record*) return ;;
  esac
  __term_last_command="$BASH_COMMAND"
  __term_started_at=$(date +%s)
}

__term_prompt_command() {
  local exit_code=$?
  [[ $__term_recording -eq 1 ]] && return
  [[ -z "$__term_last_command" ]] && return
  [[ "$__term_last_command" == " "* ]] && { __term_last_command=""; return; }

  local now=$(date +%s)
  local duration_ms=$(( (now - __term_started_at) * 1000 ))
  __term_recording=1
  term record --quiet --command "$__term_last_command" --exit-code "$exit_code" --duration-ms "$duration_ms" >/dev/null 2>&1
  __term_recording=0
  __term_last_command=""
}

trap '__term_debug_trap' DEBUG
PROMPT_COMMAND="__term_prompt_command${PROMPT_COMMAND:+;$PROMPT_COMMAND}"
`

func init() {
	hookCmd.AddCommand(hookInstallCmd)
}

func detectShell() string {
	shell := filepath.Base(os.Getenv("SHELL"))
	switch shell {
	case "zsh", "bash":
		return shell
	default:
		return shell
	}
}
