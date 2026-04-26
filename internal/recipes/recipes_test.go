package recipes

import "testing"

func TestSearchRecipes(t *testing.T) {
	all, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	hits := Search("undo last commit", all)
	if len(hits) == 0 || hits[0].ID != "git-undo-last-commit" {
		t.Fatalf("unexpected hits: %#v", hits[:min(1, len(hits))])
	}
}

func TestExtractPortFromQuery(t *testing.T) {
	if got := ExtractPort("kill port 8080"); got != "8080" {
		t.Fatalf("got %q", got)
	}
}
