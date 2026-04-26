package template

import "testing"

func TestApplyReplacesArgs(t *testing.T) {
	got, err := Apply("dotnet ef migrations add {{name}}", []string{"name"}, []string{"CreateUsersTable"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "dotnet ef migrations add CreateUsersTable" {
		t.Fatalf("got %q", got)
	}
}

func TestApplyMissingArg(t *testing.T) {
	if _, err := Apply("echo {{name}}", []string{"name"}, nil); err == nil {
		t.Fatalf("expected missing arg error")
	}
}
