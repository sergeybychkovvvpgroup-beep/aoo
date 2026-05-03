package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	envNotesDir       = "AOO_NOTES_DIR"
	legacyEnvNotesDir = "TERM_NOTES_DIR"
	envTheme          = "AOO_THEME"
	envAppDir         = "AOO_APP_DIR"
)

type File struct {
	NotesDir      string `yaml:"notes_dir"`
	NotesRepo     string `yaml:"notes_repo"`
	AppDir        string `yaml:"app_dir"`
	Theme         string `yaml:"theme"`
	FullScreen    bool   `yaml:"full_screen"`
	PickerHeight  int    `yaml:"picker_height"`
	ShowPreview   bool   `yaml:"show_preview"`
	PreviewPane   bool   `yaml:"preview_pane"`
	LastRepoCheck string `yaml:"last_repo_check"`
}

type rawFile struct {
	NotesDir      string `yaml:"notes_dir"`
	NotesRepo     string `yaml:"notes_repo"`
	AppDir        string `yaml:"app_dir"`
	Theme         string `yaml:"theme"`
	FullScreen    *bool  `yaml:"full_screen"`
	PickerHeight  *int   `yaml:"picker_height"`
	ShowPreview   *bool  `yaml:"show_preview"`
	PreviewPane   *bool  `yaml:"preview_pane"`
	LastRepoCheck string `yaml:"last_repo_check"`
}

type SetupRequiredError struct{}

func (e SetupRequiredError) Error() string {
	return strings.TrimSpace(`
notes directory is not configured

Initial setup:
  edit ~/.config/aoo/config.yaml
  and set notes_dir

Or use commands:
  aoo set-folder /path/to/notes

Temporary override:
  aoo --dir /path/to/notes
  AOO_NOTES_DIR=/path/to/notes aoo

Check current config:
  aoo config show
`)
}

func ConfigPath() (string, error) {
	if custom := strings.TrimSpace(os.Getenv("AOO_CONFIG_FILE")); custom != "" {
		return filepath.Abs(custom)
	}

	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("cannot resolve user config dir: %w", err)
	}

	return filepath.Join(dir, "aoo", "config.yaml"), nil
}

func DefaultFile() File {
	return File{
		Theme:        "auto",
		FullScreen:   true,
		PickerHeight: 14,
		ShowPreview:  true,
		PreviewPane:  false,
	}
}

func Load() (File, error) {
	path, err := ConfigPath()
	if err != nil {
		return File{}, err
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			cfg := DefaultFile()
			if saveErr := Save(cfg); saveErr != nil {
				return File{}, saveErr
			}
			return cfg, nil
		}
		return File{}, fmt.Errorf("read config %s: %w", path, err)
	}

	cfg := DefaultFile()
	var parsed rawFile
	if err := yaml.Unmarshal(raw, &parsed); err != nil {
		return File{}, fmt.Errorf("parse config %s: %w", path, err)
	}

	cfg.NotesDir = strings.TrimSpace(parsed.NotesDir)
	cfg.NotesRepo = strings.TrimSpace(parsed.NotesRepo)
	cfg.AppDir = strings.TrimSpace(parsed.AppDir)
	if value := strings.TrimSpace(parsed.Theme); value != "" {
		cfg.Theme = value
	}
	if parsed.FullScreen != nil {
		cfg.FullScreen = *parsed.FullScreen
	}
	if parsed.PickerHeight != nil {
		cfg.PickerHeight = *parsed.PickerHeight
	}
	if parsed.ShowPreview != nil {
		cfg.ShowPreview = *parsed.ShowPreview
	}
	if parsed.PreviewPane != nil {
		cfg.PreviewPane = *parsed.PreviewPane
	}
	cfg.LastRepoCheck = strings.TrimSpace(parsed.LastRepoCheck)
	if cfg.PickerHeight < 6 {
		cfg.PickerHeight = 6
	}
	return cfg, nil
}

func Save(cfg File) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	if err := os.WriteFile(path, []byte(renderConfig(cfg)), 0o644); err != nil {
		return fmt.Errorf("write config %s: %w", path, err)
	}

	return nil
}

func ResolveNotesDir(cliValue string) (string, string, error) {
	if value := strings.TrimSpace(cliValue); value != "" {
		path, err := filepath.Abs(value)
		return path, "flag --dir", err
	}

	if value := strings.TrimSpace(os.Getenv(envNotesDir)); value != "" {
		path, err := filepath.Abs(value)
		return path, envNotesDir, err
	}

	if value := strings.TrimSpace(os.Getenv(legacyEnvNotesDir)); value != "" {
		path, err := filepath.Abs(value)
		return path, legacyEnvNotesDir, err
	}

	cfg, err := Load()
	if err != nil {
		return "", "", err
	}

	if value := strings.TrimSpace(cfg.NotesDir); value != "" {
		path, err := filepath.Abs(value)
		return path, "config", err
	}

	cwd, err := os.Getwd()
	if err == nil {
		matches, globErr := filepath.Glob(filepath.Join(cwd, "*.yaml"))
		if globErr == nil && hasVisibleYAML(matches) {
			return cwd, "current directory", nil
		}
	}

	return "", "", SetupRequiredError{}
}

func SetNotesDir(dir string) (string, error) {
	path, err := filepath.Abs(strings.TrimSpace(dir))
	if err != nil {
		return "", err
	}

	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("check notes dir %s: %w", path, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", path)
	}

	cfg, err := Load()
	if err != nil {
		return "", err
	}
	cfg.NotesDir = path

	if err := Save(cfg); err != nil {
		return "", err
	}

	return path, nil
}

func SetNotesRepo(repo string) error {
	cfg, err := Load()
	if err != nil {
		return err
	}
	cfg.NotesRepo = strings.TrimSpace(repo)
	return Save(cfg)
}

func SetLastRepoCheck(value string) error {
	cfg, err := Load()
	if err != nil {
		return err
	}
	cfg.LastRepoCheck = strings.TrimSpace(value)
	return Save(cfg)
}

func ResolveAppDir(cliValue string) (string, string, error) {
	if value := strings.TrimSpace(cliValue); value != "" {
		path, err := filepath.Abs(value)
		return path, "flag --app-dir", err
	}

	if value := strings.TrimSpace(os.Getenv(envAppDir)); value != "" {
		path, err := filepath.Abs(value)
		return path, envAppDir, err
	}

	cfg, err := Load()
	if err != nil {
		return "", "", err
	}

	if value := strings.TrimSpace(cfg.AppDir); value != "" {
		path, err := filepath.Abs(value)
		return path, "config", err
	}

	cwd, err := os.Getwd()
	if err == nil {
		if _, statErr := os.Stat(filepath.Join(cwd, "go.mod")); statErr == nil {
			return cwd, "current directory", nil
		}
	}

	return "", "", nil
}

func SetAppDir(dir string) (string, error) {
	path, err := filepath.Abs(strings.TrimSpace(dir))
	if err != nil {
		return "", err
	}

	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("check app dir %s: %w", path, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", path)
	}

	cfg, err := Load()
	if err != nil {
		return "", err
	}
	cfg.AppDir = path

	if err := Save(cfg); err != nil {
		return "", err
	}

	return path, nil
}

func ResolveTheme(cliValue string) (string, string, error) {
	if value := strings.TrimSpace(cliValue); value != "" {
		return value, "flag --theme", nil
	}

	if value := strings.TrimSpace(os.Getenv(envTheme)); value != "" {
		return value, envTheme, nil
	}

	cfg, err := Load()
	if err != nil {
		return "", "", err
	}

	if value := strings.TrimSpace(cfg.Theme); value != "" {
		return value, "config", nil
	}

	return "auto", "default", nil
}

func SetTheme(theme string) (string, error) {
	theme = strings.TrimSpace(theme)
	if theme == "" {
		return "", errors.New("theme name is required")
	}

	cfg, err := Load()
	if err != nil {
		return "", err
	}
	cfg.Theme = theme

	if err := Save(cfg); err != nil {
		return "", err
	}

	return theme, nil
}

func hasVisibleYAML(matches []string) bool {
	for _, match := range matches {
		base := filepath.Base(match)
		if strings.HasPrefix(base, ".") {
			continue
		}
		return true
	}
	return false
}

func renderConfig(cfg File) string {
	cfg.NotesDir = strings.TrimSpace(cfg.NotesDir)
	cfg.NotesRepo = strings.TrimSpace(cfg.NotesRepo)
	cfg.AppDir = strings.TrimSpace(cfg.AppDir)
	cfg.Theme = strings.TrimSpace(cfg.Theme)
	cfg.LastRepoCheck = strings.TrimSpace(cfg.LastRepoCheck)
	if cfg.Theme == "" {
		cfg.Theme = DefaultFile().Theme
	}
	if cfg.PickerHeight < 6 {
		cfg.PickerHeight = DefaultFile().PickerHeight
	}

	lines := []string{
		"# aoo config",
		"#",
		"# All user settings live here.",
		"# Edit this file directly instead of keeping setup/help values in notes.",
		"#",
		"# Notes directory with your YAML files.",
		"# Example: /root/.local/share/aoo/notes",
		"notes_dir: " + yamlScalar(cfg.NotesDir),
		"",
		"# Optional git repo URL for the notes source.",
		"# Used by `aoo set-source` and repo status hints.",
		"notes_repo: " + yamlScalar(cfg.NotesRepo),
		"",
		"# Optional path to the aoo app checkout.",
		"# Built-in templates can use {{aoo_app_dir}}.",
		"app_dir: " + yamlScalar(cfg.AppDir),
		"",
		"# UI theme. Use `aoo themes` to list available values.",
		"# Recommended: auto, nord, catppuccin-mocha, solarized-dark.",
		"theme: " + yamlScalar(cfg.Theme),
		"",
		"# Enable full-screen TUI mode.",
		"# true  = use the terminal alternate screen and full terminal height",
		"# false = keep the compact inline picker mode",
		"full_screen: " + yamlScalarBool(cfg.FullScreen),
		"",
		"# Picker height in terminal rows.",
		"# Used only when full_screen is false.",
		"# Keep this small to see prior terminal output.",
		"# Minimum value is 6. Recommended range: 10-18.",
		"picker_height: " + strconv.Itoa(cfg.PickerHeight),
		"",
		"# Show second-line preview in the search results.",
		"# true  = title + preview line",
		"# false = one result per line",
		"show_preview: " + yamlScalarBool(cfg.ShowPreview),
		"",
		"# Show a side preview pane like fzf.",
		"# The pane renders the selected note/command details on the right side.",
		"preview_pane: " + yamlScalarBool(cfg.PreviewPane),
		"",
		"# Internal timestamp for repo update checks.",
		"# Safe to leave empty.",
		"last_repo_check: " + yamlScalar(cfg.LastRepoCheck),
		"",
		"# Quick reference:",
		"# - start: aoo",
		"# - validate notes: aoo validate --dir /path/to/notes",
		"# - edit config: aoo config",
		"# - inspect current config: aoo config show",
		"#",
		"# Environment overrides:",
		"# - AOO_CONFIG_FILE",
		"# - AOO_NOTES_DIR",
		"# - AOO_APP_DIR",
		"# - AOO_THEME",
		"",
	}
	return strings.Join(lines, "\n")
}

func yamlScalar(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return `""`
	}
	raw, err := yaml.Marshal(trimmed)
	if err != nil {
		return `""`
	}
	return strings.TrimSpace(string(raw))
}

func yamlScalarBool(value bool) string {
	if value {
		return "true"
	}
	return "false"
}
