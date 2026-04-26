package detect

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectNode(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "package.json")
	kinds := Detect(dir)
	if !contains(kinds, Node) {
		t.Fatalf("expected node, got %#v", kinds)
	}
}

func TestDefaultSpringFullstackConfig(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "pom.xml")
	touch(t, dir, "compose.yaml")
	if err := os.Mkdir(filepath.Join(dir, "frontend"), 0755); err != nil {
		t.Fatal(err)
	}
	touch(t, dir, "frontend", "package.json")
	cfg := DefaultConfig(dir)
	if !cfg.Shortcuts["dev"].Parallel {
		t.Fatalf("expected parallel dev shortcut")
	}
	if cfg.Shortcuts["reset-db"].Danger != "high" {
		t.Fatalf("expected reset-db danger")
	}
}

func touch(t *testing.T, dir string, parts ...string) {
	t.Helper()
	path := filepath.Join(append([]string{dir}, parts...)...)
	if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
}
