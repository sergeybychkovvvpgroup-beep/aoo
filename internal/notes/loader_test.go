package notes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadDirSkipsHiddenYAML(t *testing.T) {
	workingDir := t.TempDir()

	noteFile := filepath.Join(workingDir, "notes.yaml")
	if err := os.WriteFile(noteFile, []byte("- desc: router\n  note: |\n    ok\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	hiddenFile := filepath.Join(workingDir, ".goreleaser.yaml")
	if err := os.WriteFile(hiddenFile, []byte("project_name: term-notes\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	result := LoadDir(workingDir)
	if len(result.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", result.Errors)
	}

	if len(result.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result.Entries))
	}
}

func TestLoadBytesAcceptsExplicitMode(t *testing.T) {
	result := LoadBytes("notes.yaml", []byte(`
- desc: show interfaces
  mode: show
  note: |
    show interfaces
- desc: ssh router
  mode: run
  run: ssh root@router
`))

	if len(result.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", result.Errors)
	}
	if len(result.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result.Entries))
	}
	if got := result.Entries[0].Type(); got != TypeShow {
		t.Fatalf("expected first entry type %s, got %s", TypeShow, got)
	}
	if got := result.Entries[1].Type(); got != TypeRun {
		t.Fatalf("expected second entry type %s, got %s", TypeRun, got)
	}
}

func TestLoadBytesRejectsInvalidMode(t *testing.T) {
	result := LoadBytes("notes.yaml", []byte(`
- desc: invalid
  mode: template
  template: echo hi
`))

	if len(result.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(result.Errors))
	}
	if !strings.Contains(result.Errors[0].Error(), "invalid mode") {
		t.Fatalf("expected invalid mode error, got %v", result.Errors[0])
	}
}

func TestLoadBytesRejectsShowModeWithoutNote(t *testing.T) {
	result := LoadBytes("notes.yaml", []byte(`
- desc: invalid show
  mode: show
  run: echo hi
`))

	if len(result.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(result.Errors))
	}
	if !strings.Contains(result.Errors[0].Error(), "must have note") {
		t.Fatalf("expected mode show validation error, got %v", result.Errors[0])
	}
}
