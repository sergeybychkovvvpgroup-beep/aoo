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
	if cfg.Theme != "fzf-dark" {
		t.Fatalf("expected default theme fzf-dark, got %q", cfg.Theme)
	}
	if cfg.Layout != "bottom" {
		t.Fatalf("expected default layout bottom, got %q", cfg.Layout)
	}
	if cfg.SearchMode != "hybrid" {
		t.Fatalf("expected default search mode hybrid, got %q", cfg.SearchMode)
	}
	if !cfg.FullScreen {
		t.Fatal("expected full screen default to be true")
	}
	if cfg.ShowPreview {
		t.Fatal("expected show preview default to be false")
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
	if !strings.Contains(text, "# themes: fzf-dark, catppuccin-mocha, catppuccin-latte, dracula, nord, solarized-dark, solarized-light") {
		t.Fatalf("expected compact theme list in config, got %q", text)
	}
	if !strings.Contains(text, "theme: fzf-dark") {
		t.Fatalf("expected default theme in config, got %q", text)
	}
	if !strings.Contains(text, "layout: bottom") {
		t.Fatalf("expected default layout in config, got %q", text)
	}
	if !strings.Contains(text, "search_mode: hybrid") {
		t.Fatalf("expected default search mode in config, got %q", text)
	}
	if !strings.Contains(text, "full_screen: true") {
		t.Fatalf("expected full screen in config, got %q", text)
	}
	if !strings.Contains(text, "show_preview: false") {
		t.Fatalf("expected show preview in config, got %q", text)
	}
}

func TestLoadBackfillsMissingSearchModeIntoConfigFile(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)

	path, err := ConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}

	raw := strings.Join([]string{
		"notes_dir: \"/tmp/notes\"",
		"theme: fzf-dark",
		"layout: bottom",
		"full_screen: true",
		"picker_height: 14",
		"show_preview: false",
		"",
	}, "\n")
	if err := os.WriteFile(path, []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.SearchMode != "hybrid" {
		t.Fatalf("expected hybrid search mode after backfill, got %q", cfg.SearchMode)
	}

	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(updated), "search_mode: hybrid") {
		t.Fatalf("expected config file to be backfilled with search_mode, got %q", string(updated))
	}
}
