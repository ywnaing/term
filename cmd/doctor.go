package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ywnaing/term/internal/config"
	"github.com/ywnaing/term/internal/history"
)

type doctorResult struct {
	Status string
	Name   string
	Detail string
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check term setup",
	RunE: func(cmd *cobra.Command, args []string) error {
		results := runDoctor()
		fmt.Println("term doctor")
		fmt.Println()
		for _, result := range results {
			fmt.Printf("%s %s", result.Status, result.Name)
			if result.Detail != "" {
				fmt.Printf(" - %s", result.Detail)
			}
			fmt.Println()
		}
		if doctorHasFailures(results) {
			return fmt.Errorf("doctor found issues")
		}
		return nil
	},
}

func runDoctor() []doctorResult {
	var results []doctorResult
	results = append(results, checkTermOnPath())
	results = append(results, checkShellAndHook()...)
	results = append(results, checkHistoryDB())
	results = append(results, checkProjectConfig()...)
	return results
}

func checkTermOnPath() doctorResult {
	path, err := exec.LookPath("term")
	if err != nil {
		return doctorResult{Status: "FAIL", Name: "term on PATH", Detail: "not found; run go install . or add your Go bin directory to PATH"}
	}
	return doctorResult{Status: "OK", Name: "term on PATH", Detail: path}
}

func checkShellAndHook() []doctorResult {
	shell := detectShell()
	if _, err := hookSnippet(shell); err != nil {
		return []doctorResult{{Status: "WARN", Name: "shell", Detail: fmt.Sprintf("%s is not supported yet; zsh and bash are supported", shell)}}
	}
	path, display, err := hookRCPath(shell)
	if err != nil {
		return []doctorResult{{Status: "FAIL", Name: "shell config", Detail: err.Error()}}
	}
	installed, err := hookStatus(path)
	if err != nil {
		return []doctorResult{{Status: "FAIL", Name: "shell hook", Detail: err.Error()}}
	}
	results := []doctorResult{{Status: "OK", Name: "shell", Detail: shell}}
	if installed {
		results = append(results, doctorResult{Status: "OK", Name: "shell hook", Detail: display + " has term hook"})
	} else {
		results = append(results, doctorResult{Status: "WARN", Name: "shell hook", Detail: fmt.Sprintf("not installed; run term hook install %s", shell)})
	}
	return results
}

func checkHistoryDB() doctorResult {
	path, err := history.DefaultPath()
	if err != nil {
		return doctorResult{Status: "FAIL", Name: "history database", Detail: err.Error()}
	}
	store, err := history.Open(path)
	if err != nil {
		return doctorResult{Status: "FAIL", Name: "history database", Detail: err.Error()}
	}
	if err := store.Close(); err != nil {
		return doctorResult{Status: "FAIL", Name: "history database", Detail: err.Error()}
	}
	return doctorResult{Status: "OK", Name: "history database", Detail: path}
}

func checkProjectConfig() []doctorResult {
	cwd, _ := os.Getwd()
	dir, cfg, err := config.FindNearest(cwd)
	if os.IsNotExist(err) {
		return []doctorResult{{Status: "WARN", Name: ".term.yml", Detail: "not found; run term init"}}
	}
	if err != nil {
		return []doctorResult{{Status: "FAIL", Name: ".term.yml", Detail: err.Error()}}
	}
	results := []doctorResult{{Status: "OK", Name: ".term.yml", Detail: filepath.Join(dir, config.FileName)}}
	if cfg.Project == "" {
		results = append(results, doctorResult{Status: "WARN", Name: "project name", Detail: "missing project field"})
	} else {
		results = append(results, doctorResult{Status: "OK", Name: "project name", Detail: cfg.Project})
	}
	if len(cfg.Shortcuts) == 0 {
		results = append(results, doctorResult{Status: "WARN", Name: "shortcuts", Detail: "no shortcuts configured"})
		return results
	}
	results = append(results, doctorResult{Status: "OK", Name: "shortcuts", Detail: fmt.Sprintf("%d configured", len(cfg.Shortcuts))})
	if cfg.History.ShouldCaptureStderr() {
		results = append(results, doctorResult{Status: "OK", Name: "stderr capture", Detail: "enabled for failed commands"})
	} else {
		results = append(results, doctorResult{Status: "OK", Name: "stderr capture", Detail: "disabled; set history.capture_stderr: true to enable"})
	}
	for _, warning := range shortcutWarnings(cfg) {
		results = append(results, doctorResult{Status: "WARN", Name: "shortcut", Detail: warning})
	}
	for _, missing := range missingShortcutExecutables(cfg) {
		results = append(results, doctorResult{Status: "WARN", Name: "command dependency", Detail: missing})
	}
	return results
}

func shortcutWarnings(cfg *config.TermConfig) []string {
	var warnings []string
	keys := sortedShortcutKeys(cfg)
	for _, name := range keys {
		shortcut := cfg.Shortcuts[name]
		if len(shortcut.Steps) == 0 {
			warnings = append(warnings, fmt.Sprintf("%s has no steps", name))
			continue
		}
		for i, step := range shortcut.Steps {
			if strings.TrimSpace(step.Command) == "" {
				warnings = append(warnings, fmt.Sprintf("%s step %d has an empty command", name, i+1))
			}
		}
	}
	return warnings
}

func missingShortcutExecutables(cfg *config.TermConfig) []string {
	seen := map[string]bool{}
	var missing []string
	for _, name := range sortedShortcutKeys(cfg) {
		for _, step := range cfg.Shortcuts[name].Steps {
			for _, exe := range commandExecutables(step.Command) {
				if exe == "" || seen[exe] || shellBuiltinOrSyntax(exe) {
					continue
				}
				seen[exe] = true
				if _, err := exec.LookPath(exe); err != nil {
					missing = append(missing, fmt.Sprintf("%s referenced by shortcut %s was not found on PATH", exe, name))
				}
			}
		}
	}
	sort.Strings(missing)
	return missing
}

func commandExecutable(command string) string {
	fields := strings.Fields(strings.TrimSpace(command))
	if len(fields) == 0 {
		return ""
	}
	return strings.Trim(fields[0], `"'`)
}

func commandExecutables(command string) []string {
	fields := strings.Fields(strings.TrimSpace(command))
	var executables []string
	nextIsCommand := true
	for _, field := range fields {
		token := strings.Trim(field, `"'`)
		if token == "" {
			continue
		}
		if shellOperator(token) {
			nextIsCommand = true
			continue
		}
		if nextIsCommand {
			executables = append(executables, token)
			nextIsCommand = false
		}
	}
	return executables
}

func shellOperator(token string) bool {
	switch token {
	case "&&", "||", "|", ";":
		return true
	default:
		return false
	}
}

func shellBuiltinOrSyntax(exe string) bool {
	switch exe {
	case "cd", "echo", "export", "set", "source", ".", "alias", "unalias", "test", "[", "true", "false":
		return true
	default:
		return strings.ContainsAny(exe, "=<>|&;")
	}
}

func sortedShortcutKeys(cfg *config.TermConfig) []string {
	keys := make([]string, 0, len(cfg.Shortcuts))
	for key := range cfg.Shortcuts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func doctorHasFailures(results []doctorResult) bool {
	for _, result := range results {
		if result.Status == "FAIL" {
			return true
		}
	}
	return false
}
