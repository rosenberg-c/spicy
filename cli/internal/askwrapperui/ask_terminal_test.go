//go:build !gio

package askwrapperui

import (
	"strings"
	"testing"

	"module/lib/internal/askwrapper"
)

func TestParseIndex_ValidAndInvalid(t *testing.T) {
	// @req CLI-ASKWRAPPER-007
	entries := []askwrapper.HistoryEntry{{Question: "q1"}, {Question: "q2"}}

	idx, err := parseIndex(entries, "2")
	if err != nil || idx != 1 {
		t.Fatalf("parseIndex valid = (%d, %v), want (1, nil)", idx, err)
	}

	_, err = parseIndex(entries, "3")
	if err == nil {
		t.Fatal("expected out-of-range error")
	}
}

func TestDeleteHistory_RemovesRequestedEntry(t *testing.T) {
	// @req CLI-ASKWRAPPER-008
	home := t.TempDir()
	t.Setenv("HOME", home)

	_ = askwrapper.AppendHistory("first", "a1")
	_ = askwrapper.AppendHistory("second", "a2")
	history, _ := askwrapper.LoadHistory()

	if err := deleteHistory(&history, "1"); err != nil {
		t.Fatalf("deleteHistory error = %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("history len = %d, want 1", len(history))
	}
}

func TestOneLine_NormalizesMultiline(t *testing.T) {
	got := oneLine("a\n b\r\n c")
	if strings.Contains(got, "\n") || strings.Contains(got, "\r") {
		t.Fatalf("oneLine still has newlines: %q", got)
	}
}
