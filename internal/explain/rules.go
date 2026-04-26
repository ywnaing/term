package explain

import (
	_ "embed"
	"encoding/json"
	"regexp"
	"strings"
)

//go:embed error-rules.json
var embedded []byte

type Rule struct {
	ID       string              `json:"id"`
	Title    string              `json:"title"`
	Patterns []string            `json:"patterns"`
	Meaning  string              `json:"meaning"`
	Fixes    map[string][]string `json:"fixes"`
	Notes    []string            `json:"notes"`
}

type Match struct {
	Rule Rule
	Port string
}

func Load() ([]Rule, error) {
	var rules []Rule
	err := json.Unmarshal(embedded, &rules)
	return rules, err
}

func Find(text string, rules []Rule) (*Match, error) {
	for _, rule := range rules {
		for _, pattern := range rule.Patterns {
			re, err := regexp.Compile("(?i)" + pattern)
			if err != nil {
				return nil, err
			}
			if re.MatchString(text) {
				return &Match{Rule: rule, Port: ExtractPort(text)}, nil
			}
		}
	}
	return nil, nil
}

func ExtractPort(text string) string {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`:{1,3}([1-9][0-9]{1,4})\b`),
		regexp.MustCompile(`(?i)\bport\s+([1-9][0-9]{1,4})\b`),
		regexp.MustCompile(`\b([1-9][0-9]{3,4})\b`),
	}
	for _, re := range patterns {
		if match := re.FindStringSubmatch(text); len(match) > 1 {
			return match[1]
		}
	}
	return ""
}

func RenderTemplate(text, port string) string {
	if port == "" {
		port = "<PORT>"
	}
	return strings.ReplaceAll(text, "{{port}}", port)
}

func RenderCommands(commands []string, port string) []string {
	out := make([]string, len(commands))
	for i, command := range commands {
		out[i] = RenderTemplate(command, port)
	}
	return out
}
