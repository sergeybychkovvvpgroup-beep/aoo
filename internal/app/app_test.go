package app

import (
	"bytes"
	"strings"
	"testing"

	"aoo/internal/notes"
)

func TestPromptCommandRunPrintsCommandWithoutConfirmation(t *testing.T) {
	var stdout bytes.Buffer

	if err := promptCommandRun("List files", "ls -la", &stdout); err != nil {
		t.Fatalf("promptCommandRun returned error: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "[run] List files") {
		t.Fatalf("expected run header in output, got %q", output)
	}
	if !strings.Contains(output, "[command]\nls -la") {
		t.Fatalf("expected command to be printed, got %q", output)
	}
	if strings.Contains(output, "run?") {
		t.Fatalf("expected no confirmation prompt, got %q", output)
	}
}

func TestRunHeaderSkipsDuplicateActionDescription(t *testing.T) {
	entry := notes.Entry{Desc: "Vyos-DHCP Chashnikovo показать subnets"}
	action := &notes.Action{Desc: "Vyos-DHCP Chashnikovo показать subnets"}

	if got := runHeader(entry, action); got != entry.DisplayName() {
		t.Fatalf("expected deduplicated run header, got %q", got)
	}
}

func TestRunHeaderKeepsDistinctActionDescription(t *testing.T) {
	entry := notes.Entry{Desc: "Vyos-DHCP Chashnikovo"}
	action := &notes.Action{Desc: "показать subnets"}

	if got := runHeader(entry, action); got != "Vyos-DHCP Chashnikovo :: показать subnets" {
		t.Fatalf("unexpected run header: %q", got)
	}
}
