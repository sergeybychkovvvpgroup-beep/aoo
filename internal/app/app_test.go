package app

import (
	"bytes"
	"strings"
	"testing"
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
