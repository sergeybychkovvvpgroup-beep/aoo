package app

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"aoo/internal/ui"
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

func TestShortErrorUsesFirstLine(t *testing.T) {
	err := errors.New("first line\nsecond line")
	if got := shortError(err); got != "first line" {
		t.Fatalf("unexpected short error: %q", got)
	}
}

func TestStartNotesSyncReturnsStatus(t *testing.T) {
	ch := make(chan ui.SyncStatus, 1)
	ch <- ui.SyncStatus{State: ui.SyncStateOK}
	close(ch)

	status, ok := <-ch
	if !ok {
		t.Fatal("expected sync status")
	}
	if status.State != ui.SyncStateOK {
		t.Fatalf("unexpected sync status: %#v", status)
	}
}
