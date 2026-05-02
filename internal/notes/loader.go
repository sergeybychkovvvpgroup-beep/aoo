package notes

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

var ignoredDirs = map[string]struct{}{
	".git":      {},
	".obsidian": {},
	"_archive":  {},
	"bin":       {},
	"dist":      {},
}

type ValidationError struct {
	Path    string
	Problem string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Path, e.Problem)
}

type LoadResult struct {
	Entries []Entry
	Errors  []error
}

func LoadDir(root string) LoadResult {
	var result LoadResult
	counter := 0

	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			result.Errors = append(result.Errors, walkErr)
			return nil
		}

		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			if _, skip := ignoredDirs[d.Name()]; skip {
				return filepath.SkipDir
			}
			return nil
		}

		if !isYAMLFile(path) {
			return nil
		}

		if strings.HasPrefix(d.Name(), ".") {
			return nil
		}

		entries, err := loadFile(path)
		if err != nil {
			result.Errors = append(result.Errors, err)
			return nil
		}

		for i := range entries {
			counter++
			entries[i].index = counter
			result.Entries = append(result.Entries, entries[i])
		}

		return nil
	})

	sort.Slice(result.Entries, func(i, j int) bool {
		if result.Entries[i].SourceFile == result.Entries[j].SourceFile {
			return result.Entries[i].index < result.Entries[j].index
		}
		return result.Entries[i].SourceFile < result.Entries[j].SourceFile
	})

	return result
}

func LoadBytes(path string, raw []byte) LoadResult {
	var result LoadResult

	entries, err := loadYAML(path, raw)
	if err != nil {
		result.Errors = append(result.Errors, err)
		return result
	}

	for i := range entries {
		entries[i].index = i + 1
	}
	result.Entries = entries
	return result
}

func loadFile(path string) ([]Entry, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return loadYAML(path, raw)
}

func loadYAML(path string, raw []byte) ([]Entry, error) {

	var doc yaml.Node
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, ValidationError{Path: path, Problem: err.Error()}
	}

	if len(doc.Content) == 0 {
		return nil, nil
	}

	root := doc.Content[0]
	var nodes []*yaml.Node

	switch root.Kind {
	case yaml.SequenceNode:
		nodes = root.Content
	case yaml.MappingNode:
		nodes = []*yaml.Node{root}
	default:
		return nil, ValidationError{
			Path:    path,
			Problem: "top-level YAML must be a note object or a list of note objects",
		}
	}

	entries := make([]Entry, 0, len(nodes))
	for _, node := range nodes {
		if node.Kind != yaml.MappingNode {
			return nil, ValidationError{
				Path:    path,
				Problem: fmt.Sprintf("note at line %d must be a YAML object", node.Line),
			}
		}

		var entry Entry
		if err := node.Decode(&entry); err != nil {
			return nil, ValidationError{
				Path:    path,
				Problem: fmt.Sprintf("cannot decode note at line %d: %v", node.Line, err),
			}
		}

		if err := validateEntry(path, node.Line, entry); err != nil {
			return nil, err
		}

		entry.SourcePath = path
		entry.SourceFile = filepath.Base(path)
		entry.Tags = normalizedTags(entry.Tags, entry.SourceFile)
		entries = append(entries, entry)
	}

	return entries, nil
}

func validateEntry(path string, line int, entry Entry) error {
	if strings.TrimSpace(entry.Desc) == "" {
		return ValidationError{
			Path:    path,
			Problem: fmt.Sprintf("note at line %d is missing required field desc", line),
		}
	}

	mode := strings.ToLower(strings.TrimSpace(entry.Mode))
	if mode != "" && mode != "run" && mode != "show" {
		return ValidationError{
			Path:    path,
			Problem: fmt.Sprintf("note %q at line %d has invalid mode %q, expected run or show", entry.Desc, line, entry.Mode),
		}
	}

	if strings.TrimSpace(entry.Run) == "" && strings.TrimSpace(entry.Note) == "" && strings.TrimSpace(entry.Template) == "" {
		return ValidationError{
			Path:    path,
			Problem: fmt.Sprintf("note %q at line %d must have run, template or note", entry.Desc, line),
		}
	}

	if mode == "run" && strings.TrimSpace(entry.Run) == "" && strings.TrimSpace(entry.Template) == "" {
		return ValidationError{
			Path:    path,
			Problem: fmt.Sprintf("note %q at line %d with mode run must have run or template", entry.Desc, line),
		}
	}

	if mode == "show" && strings.TrimSpace(entry.Note) == "" {
		return ValidationError{
			Path:    path,
			Problem: fmt.Sprintf("note %q at line %d with mode show must have note", entry.Desc, line),
		}
	}

	return nil
}

func isYAMLFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".yaml" || ext == ".yml"
}

func normalizedTags(tags []string, sourceFile string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(tags)+4)

	appendTag := func(tag string) {
		tag = strings.TrimSpace(strings.ToLower(tag))
		if tag == "" {
			return
		}
		if _, exists := seen[tag]; exists {
			return
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}

	for _, tag := range tags {
		appendTag(tag)
	}

	base := strings.TrimSuffix(sourceFile, filepath.Ext(sourceFile))
	replacer := strings.NewReplacer("-", " ", "_", " ", ".", " ")
	for _, token := range strings.Fields(replacer.Replace(strings.ToLower(base))) {
		appendTag(token)
	}

	return out
}
