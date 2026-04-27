package history

import "testing"

func TestRedactSecrets(t *testing.T) {
	cases := map[string]string{
		`curl --token abc123`:                     `curl --token <redacted>`,
		`psql --password "secret value"`:          `psql --password <redacted>`,
		`API_KEY=supersecret npm test`:            `API_KEY=<redacted> npm test`,
		`Authorization: Bearer abc.def.ghi`:       `Authorization: Bearer <redacted>`,
		`docker login --password=abc registry.io`: `docker login --password=<redacted> registry.io`,
	}
	for input, want := range cases {
		if got := Redact(input); got != want {
			t.Fatalf("Redact(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestShouldSkipCommand(t *testing.T) {
	for _, command := range []string{"", "   ", " secret command", "term record --command x"} {
		if !ShouldSkipCommand(command) {
			t.Fatalf("expected %q to be skipped", command)
		}
	}
	if ShouldSkipCommand("npm test") {
		t.Fatalf("did not expect npm test to be skipped")
	}
}
