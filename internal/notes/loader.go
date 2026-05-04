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

		entry = normalizeLegacyEntry(entry)
		if err := validateEntry(path, node.Line, entry); err != nil {
			return nil, err
		}

		entry.SourcePath = path
		entry.SourceFile = filepath.Base(path)
		entry.SourceLine = node.Line
		entry.Tags = normalizedTags(entry.Tags, entry.SourceFile)
		entry.searchData = weightedFields(entry)
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

	if len(entry.Actions) == 0 {
		return ValidationError{
			Path:    path,
			Problem: fmt.Sprintf("note %q at line %d must have at least one action", entry.Desc, line),
		}
	}

	for i, action := range entry.Actions {
		actionLine := line
		if strings.TrimSpace(action.Desc) == "" {
			return ValidationError{
				Path:    path,
				Problem: fmt.Sprintf("note %q action %d at line %d is missing required field desc", entry.Desc, i+1, actionLine),
			}
		}

		kinds := 0
		if action.IsShow() {
			kinds++
		}
		if action.IsCmd() {
			kinds++
		}
		if action.IsTemplate() {
			kinds++
		}
		if kinds == 0 {
			return ValidationError{
				Path:    path,
				Problem: fmt.Sprintf("note %q action %q at line %d must have text, cmd, or template", entry.Desc, action.Desc, actionLine),
			}
		}
		if kinds > 1 {
			return ValidationError{
				Path:    path,
				Problem: fmt.Sprintf("note %q action %q at line %d must use only one of text, cmd, or template", entry.Desc, action.Desc, actionLine),
			}
		}
	}

	return nil
}

func normalizeLegacyEntry(entry Entry) Entry {
	if len(entry.Actions) > 0 {
		entry.Text = strings.TrimSpace(entry.Text)
		entry.Cmd = strings.TrimSpace(entry.Cmd)
		for i := range entry.Actions {
			entry.Actions[i].Desc = strings.TrimSpace(entry.Actions[i].Desc)
			entry.Actions[i].Cmd = strings.TrimSpace(entry.Actions[i].Cmd)
			entry.Actions[i].Text = strings.TrimSpace(entry.Actions[i].Text)
			entry.Actions[i].Template = strings.TrimSpace(entry.Actions[i].Template)
			entry.Actions[i].Banner = strings.TrimSpace(entry.Actions[i].Banner)
		}
		return entry
	}

	actions := make([]Action, 0, 1+len(entry.Run))
	entry.Text = strings.TrimSpace(entry.Text)
	entry.Cmd = strings.TrimSpace(entry.Cmd)
	if entry.Cmd != "" {
		actions = append(actions, Action{
			Desc:   inferActionDesc(entry.Cmd, "run"),
			Cmd:    entry.Cmd,
			Banner: strings.TrimSpace(entry.Banner),
		})
	}
	if entry.Text != "" {
		actions = append(actions, Action{
			Desc: "show",
			Text: entry.Text,
		})
	}
	if text := strings.TrimSpace(entry.Note); text != "" {
		actions = append(actions, Action{
			Desc: "show",
			Text: text,
		})
	}
	for i, option := range entry.Run {
		desc := strings.TrimSpace(option.Desc)
		if desc == "" {
			desc = inferActionDesc(option.Run, fmt.Sprintf("run %d", i+1))
		}
		banner := strings.TrimSpace(option.Banner)
		if banner == "" && i == 0 {
			banner = strings.TrimSpace(entry.Banner)
		}
		actions = append(actions, Action{
			Desc:   desc,
			Cmd:    strings.TrimSpace(option.Run),
			Banner: banner,
		})
	}
	if template := strings.TrimSpace(entry.Template); template != "" {
		actions = append(actions, Action{
			Desc:     inferActionDesc(template, "run"),
			Template: template,
			Args:     entry.Args,
			Banner:   strings.TrimSpace(entry.Banner),
		})
	}
	entry.Actions = actions
	return entry
}

func inferActionDesc(command string, fallback string) string {
	command = strings.TrimSpace(command)
	if command == "" {
		return fallback
	}

	if parts := strings.Split(command, "&&"); len(parts) > 1 {
		command = strings.TrimSpace(parts[len(parts)-1])
	}
	command = strings.TrimSpace(strings.TrimSuffix(command, ";"))

	fields := strings.Fields(command)
	if len(fields) == 0 {
		return fallback
	}

	first := fields[0]
	if first == "sudo" && len(fields) > 1 {
		first = fields[1]
		fields = fields[1:]
	}
	switch first {
	case "ssh":
		for i := 1; i < len(fields); i++ {
			if fields[i] == "-L" {
				if i+1 < len(fields) {
					if port := strings.SplitN(fields[i+1], ":", 2)[0]; port != "" {
						return "tunnel " + port
					}
				}
				return "tunnel"
			}
			if strings.HasPrefix(fields[i], "-L") && len(fields[i]) > 2 {
				if port := strings.SplitN(strings.TrimPrefix(fields[i], "-L"), ":", 2)[0]; port != "" {
					return "tunnel " + port
				}
				return "tunnel"
			}
		}
		return "ssh"
	case "grep", "find", "dig", "curl", "make", "git", "docker", "mount", "nmap", "echo", "aoo":
		return first
	default:
		return fallback
	}
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
