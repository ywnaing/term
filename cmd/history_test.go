package cmd

import "testing"

func TestParseHistoryID(t *testing.T) {
	id, err := parseHistoryID("12")
	if err != nil {
		t.Fatal(err)
	}
	if id != 12 {
		t.Fatalf("got %d", id)
	}
	for _, input := range []string{"", "abc", "0", "-1"} {
		if _, err := parseHistoryID(input); err == nil {
			t.Fatalf("expected invalid id error for %q", input)
		}
	}
}
