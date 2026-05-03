package config

import (
	"os"
	"path/filepath"
	"strings"
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

func TestLoadCreatesCommentedConfigWithDefaults(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.PickerHeight != 14 {
		t.Fatalf("expected default picker height 14, got %d", cfg.PickerHeight)
	}
	if !cfg.FullScreen {
		t.Fatal("expected full screen default to be true")
	}
	if cfg.ShowPreview {
		t.Fatal("expected show preview default to be false")
	}
	if cfg.PreviewPane {
		t.Fatal("expected preview pane default to be false")
	}

	path, err := ConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	if !strings.Contains(text, "# aoo config") {
		t.Fatalf("expected commented config template, got %q", text)
	}
	if !strings.Contains(text, "picker_height: 14") {
		t.Fatalf("expected picker height in config, got %q", text)
	}
	if !strings.Contains(text, "full_screen: true") {
		t.Fatalf("expected full screen in config, got %q", text)
	}
	if !strings.Contains(text, "show_preview: false") {
		t.Fatalf("expected show preview in config, got %q", text)
	}
	if !strings.Contains(text, "preview_pane: false") {
		t.Fatalf("expected preview pane in config, got %q", text)
	}
}
