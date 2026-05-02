package app

import (
	"bytes"
	"strings"
	"testing"
)

func TestPromptCommandRunConfirmsYes(t *testing.T) {
	var stdout bytes.Buffer

	confirmed, err := promptCommandRun("List files", "ls -la", strings.NewReader("y\n"), &stdout)
	if err != nil {
		t.Fatalf("promptCommandRun returned error: %v", err)
	}
	if !confirmed {
		t.Fatal("expected confirmation to be accepted")
	}

	output := stdout.String()
	if !strings.Contains(output, "[run] List files") {
		t.Fatalf("expected run header in output, got %q", output)
	}
	if !strings.Contains(output, "[command]\nls -la") {
		t.Fatalf("expected command to be printed, got %q", output)
	}
}

func TestPromptCommandRunRejectsDefault(t *testing.T) {
	var stdout bytes.Buffer

	confirmed, err := promptCommandRun("List files", "ls -la", strings.NewReader("\n"), &stdout)
	if err != nil {
		t.Fatalf("promptCommandRun returned error: %v", err)
	}
	if confirmed {
		t.Fatal("expected empty answer to cancel execution")
	}
}
