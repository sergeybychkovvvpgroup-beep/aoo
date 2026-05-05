package bundled

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"aoo/internal/notes"
)

//go:embed *.yaml
var bundledFS embed.FS

func Load() notes.LoadResult {
	files, err := fs.Glob(bundledFS, "*.yaml")
	if err != nil {
		return notes.LoadResult{Errors: []error{err}}
	}
	sort.Strings(files)

	result := notes.LoadResult{}
	for _, name := range files {
		raw, readErr := bundledFS.ReadFile(name)
		if readErr != nil {
			result.Errors = append(result.Errors, readErr)
			continue
		}
		loaded := notes.LoadBytes("builtin/"+name, raw)
		result.Entries = append(result.Entries, loaded.Entries...)
		result.Errors = append(result.Errors, loaded.Errors...)
	}
	return result
}

func Materialize(sourcePath string, targetDir string) (string, error) {
	filename := strings.TrimSpace(strings.TrimPrefix(sourcePath, "builtin/"))
	if filename == "" || filename == sourcePath {
		return "", fmt.Errorf("unknown bundled source: %s", sourcePath)
	}

	content, err := bundledFS.ReadFile(filename)
	if err != nil {
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
