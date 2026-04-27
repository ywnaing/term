package history

import (
	"regexp"
	"strings"
)

var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(--(?:password|passwd|token|api-key|apikey|secret)(?:=|\s+))("[^"]*"|'[^']*'|\S+)`),
	regexp.MustCompile(`(?i)\b([A-Z0-9_]*(?:PASSWORD|PASSWD|TOKEN|SECRET|API_KEY|AUTH_KEY)[A-Z0-9_]*=)("[^"]*"|'[^']*'|[^\s]+)`),
	regexp.MustCompile(`(?i)(Authorization:\s*Bearer\s+)([A-Za-z0-9._~+/\-=]+)`),
}

func Redact(text string) string {
	out := text
	for _, pattern := range secretPatterns {
		out = pattern.ReplaceAllString(out, `${1}<redacted>`)
	}
	return out
}

func ShouldSkipCommand(command string) bool {
	if command == "" {
		return true
	}
	if strings.HasPrefix(command, " ") {
		return true
	}
	trimmed := strings.TrimSpace(command)
	if trimmed == "" {
		return true
	}
	if strings.HasPrefix(trimmed, "term record") {
		return true
	}
	return false
}
