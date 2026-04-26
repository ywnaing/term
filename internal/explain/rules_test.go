package explain

import "testing"

func TestFindPortInUse(t *testing.T) {
	rules, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	match, err := Find("EADDRINUSE: address already in use :::3000", rules)
	if err != nil {
		t.Fatal(err)
	}
	if match == nil || match.Rule.ID != "port-in-use" {
		t.Fatalf("unexpected match: %#v", match)
	}
	if match.Port != "3000" {
		t.Fatalf("got port %q", match.Port)
	}
}

func TestExtractPortFromError(t *testing.T) {
	if got := ExtractPort("Port 8080 was already in use"); got != "8080" {
		t.Fatalf("got %q", got)
	}
}
