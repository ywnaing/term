package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadParsesStringAndObjectSteps(t *testing.T) {
	dir := t.TempDir()
	content := []byte(`project: demo
shortcuts:
  dev:
    description: Run dev
    steps:
      - npm run dev
      - name: frontend
        command: cd frontend && npm run dev
history:
  capture_stderr: true
`)
	if err := os.WriteFile(filepath.Join(dir, FileName), content, 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	steps := cfg.Shortcuts["dev"].Steps
	if steps[0].Command != "npm run dev" {
		t.Fatalf("string step not parsed: %#v", steps[0])
	}
	if steps[1].Name != "frontend" || steps[1].Command != "cd frontend && npm run dev" {
		t.Fatalf("object step not parsed: %#v", steps[1])
	}
	if !cfg.History.ShouldCaptureStderr() {
		t.Fatalf("expected stderr capture to be enabled")
	}
}

func TestWriteDefaultConfig(t *testing.T) {
	dir := t.TempDir()
	cfg := TermConfig{Project: "demo", Shortcuts: map[string]Shortcut{"dev": {Description: "Run", Steps: []Step{{Command: "npm run dev"}}}}}
	if err := Write(dir, cfg); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, FileName)); err != nil {
		t.Fatal(err)
	}
}
