package cmd

import (
	"strings"
	"testing"

	"github.com/ywnaing/term/internal/config"
)

func TestCommandExecutable(t *testing.T) {
	cases := map[string]string{
		"npm run dev":                "npm",
		`"go" test ./...`:            "go",
		"cd frontend && npm run dev": "cd",
		"":                           "",
		"  docker compose up db  ":   "docker",
	}
	for input, want := range cases {
		if got := commandExecutable(input); got != want {
			t.Fatalf("commandExecutable(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestCommandExecutables(t *testing.T) {
	got := strings.Join(commandExecutables("cd frontend && npm run dev"), ",")
	if got != "cd,npm" {
		t.Fatalf("got %q", got)
	}
	got = strings.Join(commandExecutables("docker compose up db | tee output.log"), ",")
	if got != "docker,tee" {
		t.Fatalf("got %q", got)
	}
}

func TestShortcutWarnings(t *testing.T) {
	cfg := &config.TermConfig{Shortcuts: map[string]config.Shortcut{
		"empty": {Description: "Empty"},
		"blank": {Description: "Blank", Steps: []config.Step{{Command: "   "}}},
	}}
	warnings := strings.Join(shortcutWarnings(cfg), "\n")
	for _, want := range []string{"empty has no steps", "blank step 1 has an empty command"} {
		if !strings.Contains(warnings, want) {
			t.Fatalf("expected warning %q in %q", want, warnings)
		}
	}
}

func TestDoctorHasFailures(t *testing.T) {
	if !doctorHasFailures([]doctorResult{{Status: "FAIL"}}) {
		t.Fatalf("expected FAIL to count as failure")
	}
	if doctorHasFailures([]doctorResult{{Status: "WARN"}, {Status: "OK"}}) {
		t.Fatalf("did not expect WARN to count as failure")
	}
}

func TestShellBuiltinOrSyntax(t *testing.T) {
	for _, exe := range []string{"cd", "echo", "API_KEY=value"} {
		if !shellBuiltinOrSyntax(exe) {
			t.Fatalf("expected %q to be treated as shell syntax/builtin", exe)
		}
	}
	if shellBuiltinOrSyntax("npm") {
		t.Fatalf("did not expect npm to be shell syntax/builtin")
	}
}
