package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const (
	hookStartMarker = "# >>> term hook >>>"
	hookEndMarker   = "# <<< term hook <<<"
)

var hookCmd = &cobra.Command{
	Use:   "hook",
	Short: "Shell hook helpers",
}

var hookInstallFlags struct {
	print  bool
	dryRun bool
}

var hookInstallCmd = &cobra.Command{
	Use:   "install [zsh|bash]",
	Short: "Install shell hook",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		shell := hookShell(args)
		snippet, err := hookSnippet(shell)
		if err != nil {
			return err
		}
		block := managedHookBlock(snippet)
		if hookInstallFlags.print {
			_, _ = os.Stdout.WriteString(snippet)
			return nil
		}
		path, display, err := hookRCPath(shell)
		if err != nil {
			return err
		}
		if hookInstallFlags.dryRun {
			fmt.Printf("Would install term hook in %s\n\n%s", display, block)
			return nil
		}
		backup, backupDisplay, err := installHook(path, display, block, time.Now())
		if err != nil {
			return err
		}
		fmt.Printf("Installed term hook in %s\n", display)
		fmt.Printf("Backup created: %s\n", backupDisplay)
		fmt.Println("Run:")
		fmt.Printf("  source %s\n", display)
		_ = backup
		return nil
	},
}

var hookUninstallCmd = &cobra.Command{
	Use:   "uninstall [zsh|bash]",
	Short: "Uninstall shell hook",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		shell := hookShell(args)
		if _, err := hookSnippet(shell); err != nil {
			return err
		}
		path, display, err := hookRCPath(shell)
		if err != nil {
			return err
		}
		removed, _, backupDisplay, err := uninstallHook(path, display, time.Now())
		if err != nil {
			return err
		}
		if !removed {
			fmt.Printf("No term hook found in %s\n", display)
			return nil
		}
		fmt.Printf("Removed term hook from %s\n", display)
		fmt.Printf("Backup created: %s\n", backupDisplay)
		fmt.Println("Run:")
		fmt.Printf("  source %s\n", display)
		return nil
	},
}

var hookStatusCmd = &cobra.Command{
	Use:   "status [zsh|bash]",
	Short: "Show shell hook status",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		shell := hookShell(args)
		if _, err := hookSnippet(shell); err != nil {
			return err
		}
		path, display, err := hookRCPath(shell)
		if err != nil {
			return err
		}
		status, err := hookStatus(path)
		if err != nil {
			return err
		}
		fmt.Printf("Shell: %s\n", shell)
		fmt.Printf("Config: %s\n", display)
		if status {
			fmt.Println("Status: installed")
		} else {
			fmt.Println("Status: not installed")
			fmt.Println("Run:")
			fmt.Printf("  term hook install %s\n", shell)
		}
		return nil
	},
}

const zshHookSnippet = `# term shell hook (experimental)
# Records command text, exit code, cwd, project, timestamp, shell, OS, and duration.
# It captures stderr for failed commands only when history.capture_stderr is enabled.
# Commands that start with a space are skipped.
#
# - term history search <query>
# - term explain last when stderr capture is enabled
#
autoload -Uz add-zsh-hook
zmodload zsh/datetime 2>/dev/null || true

__term_last_command=""
__term_started_at=0
__term_recording=0
__term_stderr_file=""
__term_stderr_active=0

__term_preexec() {
  [[ $__term_recording -eq 1 ]] && return
  __term_last_command="$1"
  __term_started_at=${EPOCHSECONDS:-$(date +%s)}
  __term_stderr_file=""
  __term_stderr_active=0

  if [[ "$(term record --capture-stderr-enabled 2>/dev/null)" == "true" ]]; then
    __term_stderr_file="$(mktemp "${TMPDIR:-/tmp}/term-stderr.XXXXXX" 2>/dev/null || true)"
    if [[ -n "$__term_stderr_file" ]]; then
      exec 9>&2
      exec 2> >(tee "$__term_stderr_file" >&9)
      __term_stderr_active=1
    fi
  fi
}

__term_precmd() {
  local exit_code=$?
  [[ $__term_recording -eq 1 ]] && return
  [[ -z "$__term_last_command" ]] && return

  if [[ $__term_stderr_active -eq 1 ]]; then
    exec 2>&9
    exec 9>&-
  fi
  [[ "$__term_last_command" == " "* ]] && { [[ -n "$__term_stderr_file" ]] && rm -f "$__term_stderr_file"; __term_last_command=""; __term_stderr_file=""; __term_stderr_active=0; return; }

  local now=${EPOCHSECONDS:-$(date +%s)}
  local duration_ms=$(( (now - __term_started_at) * 1000 ))
  local stderr_arg=()
  if [[ "$exit_code" -ne 0 && -n "$__term_stderr_file" && -s "$__term_stderr_file" ]]; then
    stderr_arg=(--stderr "$(head -c 16384 "$__term_stderr_file")")
  fi
  __term_recording=1
  term record --quiet --command "$__term_last_command" --exit-code "$exit_code" --duration-ms "$duration_ms" "${stderr_arg[@]}" >/dev/null 2>&1
  __term_recording=0
  [[ -n "$__term_stderr_file" ]] && rm -f "$__term_stderr_file"
  __term_last_command=""
  __term_stderr_file=""
  __term_stderr_active=0
}

add-zsh-hook preexec __term_preexec
add-zsh-hook precmd __term_precmd
`

const bashHookSnippet = `# term shell hook (experimental)
# Records command text, exit code, cwd, project, timestamp, shell, OS, and duration.
# It captures stderr for failed commands only when history.capture_stderr is enabled.
# Commands that start with a space are skipped.
#
# - term history search <query>
# - term explain last when stderr capture is enabled
#
__term_last_command=""
__term_started_at=0
__term_recording=0
__term_stderr_file=""
__term_stderr_active=0

__term_debug_trap() {
  [[ $__term_recording -eq 1 ]] && return
  case "$BASH_COMMAND" in
    __term_*|term\ record*) return ;;
  esac
  __term_last_command="$BASH_COMMAND"
  __term_started_at=$(date +%s)
  __term_stderr_file=""
  __term_stderr_active=0

  if [[ "$(term record --capture-stderr-enabled 2>/dev/null)" == "true" ]]; then
    __term_stderr_file="$(mktemp "${TMPDIR:-/tmp}/term-stderr.XXXXXX" 2>/dev/null || true)"
    if [[ -n "$__term_stderr_file" ]]; then
      exec 9>&2
      exec 2> >(tee "$__term_stderr_file" >&9)
      __term_stderr_active=1
    fi
  fi
}

__term_prompt_command() {
  local exit_code=$?
  [[ $__term_recording -eq 1 ]] && return
  [[ -z "$__term_last_command" ]] && return

  if [[ $__term_stderr_active -eq 1 ]]; then
    exec 2>&9
    exec 9>&-
  fi
  [[ "$__term_last_command" == " "* ]] && { [[ -n "$__term_stderr_file" ]] && rm -f "$__term_stderr_file"; __term_last_command=""; __term_stderr_file=""; __term_stderr_active=0; return; }

  local now=$(date +%s)
  local duration_ms=$(( (now - __term_started_at) * 1000 ))
  local stderr_arg=()
  if [[ "$exit_code" -ne 0 && -n "$__term_stderr_file" && -s "$__term_stderr_file" ]]; then
    stderr_arg=(--stderr "$(head -c 16384 "$__term_stderr_file")")
  fi
  __term_recording=1
  term record --quiet --command "$__term_last_command" --exit-code "$exit_code" --duration-ms "$duration_ms" "${stderr_arg[@]}" >/dev/null 2>&1
  __term_recording=0
  [[ -n "$__term_stderr_file" ]] && rm -f "$__term_stderr_file"
  __term_last_command=""
  __term_stderr_file=""
  __term_stderr_active=0
}

trap '__term_debug_trap' DEBUG
PROMPT_COMMAND="__term_prompt_command${PROMPT_COMMAND:+;$PROMPT_COMMAND}"
`

func init() {
	hookInstallCmd.Flags().BoolVar(&hookInstallFlags.print, "print", false, "print hook snippet without installing")
	hookInstallCmd.Flags().BoolVar(&hookInstallFlags.dryRun, "dry-run", false, "preview hook installation without changing files")
	hookCmd.AddCommand(hookInstallCmd, hookUninstallCmd, hookStatusCmd)
}

func hookShell(args []string) string {
	if len(args) > 0 {
		return strings.ToLower(args[0])
	}
	return detectShell()
}

func hookSnippet(shell string) (string, error) {
	switch shell {
	case "zsh":
		return zshHookSnippet, nil
	case "bash":
		return bashHookSnippet, nil
	default:
		return "", unsupportedShellError(shell)
	}
}

func unsupportedShellError(shell string) error {
	return fmt.Errorf("unsupported shell: %s\nRun:\n  term hook install zsh\n  term hook install bash", shell)
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

func hookRCPath(shell string) (string, string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", err
	}
	switch shell {
	case "zsh":
		return filepath.Join(home, ".zshrc"), "~/.zshrc", nil
	case "bash":
		return filepath.Join(home, ".bashrc"), "~/.bashrc", nil
	default:
		return "", "", unsupportedShellError(shell)
	}
}

func managedHookBlock(snippet string) string {
	return hookStartMarker + "\n" + strings.TrimSpace(snippet) + "\n" + hookEndMarker + "\n"
}

func installHook(path, display, block string, now time.Time) (string, string, error) {
	current, err := readOptionalFile(path)
	if err != nil {
		return "", "", err
	}
	backup, backupDisplay, err := backupFile(path, display, current, now)
	if err != nil {
		return "", "", err
	}
	next := upsertManagedBlock(current, block)
	if err := os.WriteFile(path, []byte(next), 0644); err != nil {
		return "", "", err
	}
	return backup, backupDisplay, nil
}

func uninstallHook(path, display string, now time.Time) (bool, string, string, error) {
	current, err := readOptionalFile(path)
	if err != nil {
		return false, "", "", err
	}
	next, removed := removeManagedBlock(current)
	if !removed {
		return false, "", "", nil
	}
	backup, backupDisplay, err := backupFile(path, display, current, now)
	if err != nil {
		return false, "", "", err
	}
	if err := os.WriteFile(path, []byte(next), 0644); err != nil {
		return false, "", "", err
	}
	return true, backup, backupDisplay, nil
}

func readOptionalFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	return string(data), err
}

func hookStatus(path string) (bool, error) {
	content, err := readOptionalFile(path)
	if err != nil {
		return false, err
	}
	return hasManagedBlock(content), nil
}

func hasManagedBlock(content string) bool {
	start := strings.Index(content, hookStartMarker)
	if start == -1 {
		return false
	}
	end := strings.Index(content[start:], hookEndMarker)
	return end != -1
}

func backupFile(path, display, content string, now time.Time) (string, string, error) {
	timestamp := now.Format("20060102T150405.000000000")
	backup := fmt.Sprintf("%s.term.bak.%s", path, timestamp)
	backupDisplay := fmt.Sprintf("%s.term.bak.%s", display, timestamp)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", "", err
	}
	if err := os.WriteFile(backup, []byte(content), 0644); err != nil {
		return "", "", err
	}
	return backup, backupDisplay, nil
}

func upsertManagedBlock(content, block string) string {
	next, removed := removeManagedBlock(content)
	if removed {
		content = strings.TrimRight(next, "\n")
		if content == "" {
			return block
		}
		return content + "\n\n" + block
	}
	content = strings.TrimRight(content, "\n")
	if content == "" {
		return block
	}
	return content + "\n\n" + block
}

func removeManagedBlock(content string) (string, bool) {
	if !hasManagedBlock(content) {
		return content, false
	}
	start := strings.Index(content, hookStartMarker)
	end := strings.Index(content[start:], hookEndMarker)
	end = start + end + len(hookEndMarker)
	if end < len(content) && content[end] == '\n' {
		end++
	}
	next := content[:start] + content[end:]
	next = strings.TrimRight(next, "\n")
	if next != "" {
		next += "\n"
	}
	return next, true
}
