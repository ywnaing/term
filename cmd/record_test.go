package cmd

import (
	"strings"
	"testing"
)

func TestLimitRecordedStderrLeavesSmallText(t *testing.T) {
	input := "EADDRINUSE :::3000"
	if got := limitRecordedStderr(input); got != input {
		t.Fatalf("got %q, want %q", got, input)
	}
}

func TestLimitRecordedStderrTruncatesLargeText(t *testing.T) {
	input := strings.Repeat("x", maxRecordedStderrBytes+10)
	got := limitRecordedStderr(input)
	if len(got) <= maxRecordedStderrBytes {
		t.Fatalf("expected truncation marker, got length %d", len(got))
	}
	if !strings.Contains(got, "stderr truncated to 16384 bytes") {
		t.Fatalf("missing truncation marker: %q", got[len(got)-80:])
	}
}
