package bundled

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"aoo/internal/notes"
)

//go:embed service.yaml
var serviceYAML []byte

//go:embed command_templates.yaml
var templatesYAML []byte

const (
	serviceSourcePath   = "builtin/service.yaml"
	templatesSourcePath = "builtin/command_templates.yaml"
)

func Load() notes.LoadResult {
	service := notes.LoadBytes(serviceSourcePath, serviceYAML)
	templates := notes.LoadBytes(templatesSourcePath, templatesYAML)

	return notes.LoadResult{
		Entries: append(service.Entries, templates.Entries...),
		Errors:  append(service.Errors, templates.Errors...),
	}
}

func Materialize(sourcePath string, targetDir string) (string, error) {
	filename, content, ok := lookupBundledSource(sourcePath)
	if !ok {
		return "", fmt.Errorf("unknown bundled source: %s", sourcePath)
	}

	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return "", err
	}

	targetPath := filepath.Join(targetDir, filename)
	if _, err := os.Stat(targetPath); err == nil {
		return targetPath, nil
	} else if !os.IsNotExist(err) {
		return "", err
	}

	if err := os.WriteFile(targetPath, content, 0o644); err != nil {
		return "", err
	}

	return targetPath, nil
}

func FindEntryLine(path string, desc string) int {
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0
	}

	pattern := "- desc: " + desc
	lines := strings.Split(string(raw), "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) == pattern {
			return i + 1
		}
	}
	return 0
}

func lookupBundledSource(sourcePath string) (string, []byte, bool) {
	switch strings.TrimSpace(sourcePath) {
	case serviceSourcePath:
		return "aoo_service.yaml", serviceYAML, true
	case templatesSourcePath:
		return "aoo_command_templates.yaml", templatesYAML, true
	default:
		return "", nil, false
	}
}
