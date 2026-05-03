package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestUpsertManagedBlockIntoEmptyContent(t *testing.T) {
	block := managedHookBlock("echo term")
	got := upsertManagedBlock("", block)
	if got != block {
		t.Fatalf("got %q, want %q", got, block)
	}
}

func TestUpsertManagedBlockAppendsToExistingContent(t *testing.T) {
	block := managedHookBlock("echo term")
	got := upsertManagedBlock("export PATH=$PATH:/bin\n", block)
	want := "export PATH=$PATH:/bin\n\n" + block
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestUpsertManagedBlockReplacesExistingBlock(t *testing.T) {
	oldBlock := managedHookBlock("echo old")
	newBlock := managedHookBlock("echo new")
	got := upsertManagedBlock("before\n\n"+oldBlock+"\nafter\n", newBlock)
	if strings.Count(got, hookStartMarker) != 1 {
		t.Fatalf("expected one start marker, got %q", got)
	}
	if strings.Contains(got, "echo old") {
		t.Fatalf("old hook was not replaced: %q", got)
	}
	if !strings.Contains(got, "echo new") {
		t.Fatalf("new hook missing: %q", got)
	}
}

func TestRemoveManagedBlockOnlyRemovesHook(t *testing.T) {
	block := managedHookBlock("echo term")
	got, removed := removeManagedBlock("before\n\n" + block + "\nafter\n")
	if !removed {
		t.Fatalf("expected block to be removed")
	}
	if strings.Contains(got, hookStartMarker) || strings.Contains(got, "echo term") {
		t.Fatalf("hook content remains: %q", got)
	}
	if !strings.Contains(got, "before") || !strings.Contains(got, "after") {
		t.Fatalf("non-hook content was removed: %q", got)
	}
}

func TestUnsupportedShellError(t *testing.T) {
	err := unsupportedShellError("fish")
	if err == nil || !strings.Contains(err.Error(), "unsupported shell: fish") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInstallAndUninstallHookCreateBackups(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".zshrc")
	if err := os.WriteFile(path, []byte("export PATH=$PATH:/bin\n"), 0644); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 5, 3, 12, 30, 0, 0, time.UTC)
	block := managedHookBlock("echo term")
	backup, _, err := installHook(path, "~/.zshrc", block, now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(backup); err != nil {
		t.Fatalf("backup missing: %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), hookStartMarker) {
		t.Fatalf("hook was not installed: %q", content)
	}
	removed, backup, _, err := uninstallHook(path, "~/.zshrc", now.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if !removed {
		t.Fatalf("expected hook to be removed")
	}
	if _, err := os.Stat(backup); err != nil {
		t.Fatalf("uninstall backup missing: %v", err)
	}
	content, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(content), hookStartMarker) {
		t.Fatalf("hook was not removed: %q", content)
	}
}
