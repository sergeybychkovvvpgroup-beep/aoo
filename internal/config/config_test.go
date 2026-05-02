package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveNotesDirUsesConfig(t *testing.T) {
	configHome := t.TempDir()
	notesDir := filepath.Join(t.TempDir(), "notes")
	if err := os.MkdirAll(notesDir, 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("XDG_CONFIG_HOME", configHome)
	t.Setenv("AOO_NOTES_DIR", "")
	t.Setenv("TERM_NOTES_DIR", "")

	if _, err := SetNotesDir(notesDir); err != nil {
		t.Fatal(err)
	}

	got, source, err := ResolveNotesDir("")
	if err != nil {
		t.Fatal(err)
	}

	if got != notesDir {
		t.Fatalf("expected %s, got %s", notesDir, got)
	}
	if source != "config" {
		t.Fatalf("expected source=config, got %s", source)
	}
}

func TestResolveNotesDirReturnsSetupErrorWithoutSources(t *testing.T) {
	configHome := t.TempDir()
	workdir := t.TempDir()

	t.Setenv("XDG_CONFIG_HOME", configHome)
	t.Setenv("AOO_NOTES_DIR", "")
	t.Setenv("TERM_NOTES_DIR", "")

	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldwd) }()

	if err := os.Chdir(workdir); err != nil {
		t.Fatal(err)
	}

	_, _, err = ResolveNotesDir("")
	if err == nil {
		t.Fatal("expected setup error")
	}

	if _, ok := err.(SetupRequiredError); !ok {
		t.Fatalf("expected SetupRequiredError, got %T", err)
	}
}
