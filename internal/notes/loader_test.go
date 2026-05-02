package notes

import (
	"os"
	"path/filepath"
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
