package template

import (
	"fmt"
	"regexp"
	"strings"
)

var tokenRE = regexp.MustCompile(`\{\{\s*([A-Za-z0-9_-]+)\s*\}\}`)

func Apply(command string, names, values []string) (string, error) {
	if len(values) < len(names) {
		return "", fmt.Errorf("missing argument: %s", names[len(values)])
	}
	replacements := map[string]string{}
	for i, name := range names {
		replacements[name] = values[i]
	}
	missing := ""
	out := tokenRE.ReplaceAllStringFunc(command, func(match string) string {
		key := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(match, "{{"), "}}"))
		value, ok := replacements[key]
		if !ok {
			missing = key
			return match
		}
		return value
	})
	if missing != "" {
		return "", fmt.Errorf("unknown argument placeholder: %s", missing)
	}
	return out, nil
}
