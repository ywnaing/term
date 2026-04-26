package history

import (
	"path/filepath"
	"testing"
)

func TestRecordSearchAndLatestFailed(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "term.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	_, err = store.Insert(Record{Command: "npm run dev", Cwd: "/tmp/project", ProjectName: "project", ExitCode: 1, Stderr: "EADDRINUSE :::3000", StartedAt: "2026-04-26T10:30:00Z"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.Insert(Record{Command: "docker compose exec db psql -U postgres", Cwd: "/tmp/project", ProjectName: "project", ExitCode: 0, StartedAt: "2026-04-26T10:31:00Z"})
	if err != nil {
		t.Fatal(err)
	}
	results, err := store.Search("postgres", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].ExitCode != 0 {
		t.Fatalf("unexpected results: %#v", results)
	}
	failed, err := store.LatestFailed()
	if err != nil {
		t.Fatal(err)
	}
	if failed == nil || failed.Command != "npm run dev" {
		t.Fatalf("unexpected failed record: %#v", failed)
	}
}
