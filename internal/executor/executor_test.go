package executor

import "testing"

func TestShell(t *testing.T) {
	name, args := shell("echo hi")
	if name == "" || len(args) == 0 {
		t.Fatalf("invalid shell: %s %#v", name, args)
	}
}
