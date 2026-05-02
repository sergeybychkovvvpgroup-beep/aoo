package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	envNotesDir       = "AOO_NOTES_DIR"
	legacyEnvNotesDir = "TERM_NOTES_DIR"
	envTheme          = "AOO_THEME"
)

type File struct {
	NotesDir string `yaml:"notes_dir"`
	Theme    string `yaml:"theme"`
}

type SetupRequiredError struct{}

func (e SetupRequiredError) Error() string {
	return strings.TrimSpace(`
notes directory is not configured

Initial setup:
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

func Load() (File, error) {
	path, err := ConfigPath()
	if err != nil {
		return File{}, err
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return File{}, nil
		}
		return File{}, fmt.Errorf("read config %s: %w", path, err)
	}

	var cfg File
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return File{}, fmt.Errorf("parse config %s: %w", path, err)
	}

	cfg.NotesDir = strings.TrimSpace(cfg.NotesDir)
	cfg.Theme = strings.TrimSpace(cfg.Theme)
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

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
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
