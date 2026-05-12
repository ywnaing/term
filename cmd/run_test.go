package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/ywnaing/term/internal/config"
)

func TestResolveRunStepsReplacesArgs(t *testing.T) {
	shortcut := config.Shortcut{
		Args:  []string{"name"},
		Steps: []config.Step{{Command: "dotnet ef migrations add {{name}}"}},
	}
	steps, err := resolveRunSteps(shortcut, []string{"CreateUsersTable"})
	if err != nil {
		t.Fatal(err)
	}
	if steps[0].Command != "dotnet ef migrations add CreateUsersTable" {
		t.Fatalf("got %q", steps[0].Command)
	}
}

func TestPrintRunDryRun(t *testing.T) {
	shortcut := config.Shortcut{
		Parallel: true,
		Confirm:  true,
		Danger:   "high",
		Args:     []string{"name"},
	}
	steps := []config.Step{{Name: "frontend", Command: "npm run dev"}}
	output := captureStdout(t, func() {
		printRunDryRun("dev", shortcut, []string{"web"}, steps)
	})
	for _, want := range []string{
		"Shortcut: dev",
		"Mode: parallel",
		"Danger: high",
		"Confirmation required: yes",
		"name = web",
		"1. [frontend] npm run dev",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected %q in output:\n%s", want, output)
		}
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	fn()
	_ = w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}
