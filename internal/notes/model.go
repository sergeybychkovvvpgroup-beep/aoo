package notes

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Arg struct {
	Name        string `yaml:"name"`
	Prompt      string `yaml:"prompt"`
	Default     string `yaml:"default"`
	Example     string `yaml:"example"`
	Description string `yaml:"description"`
}

type Action struct {
	Desc     string `yaml:"desc"`
	Cmd      string `yaml:"cmd"`
	Text     string `yaml:"text"`
	Template string `yaml:"template"`
	Args     []Arg  `yaml:"args"`
	Banner   string `yaml:"banner"`
}

func (a Action) IsShow() bool {
	return strings.TrimSpace(a.Text) != ""
}

func (a Action) IsCmd() bool {
	return strings.TrimSpace(a.Cmd) != ""
}

func (a Action) IsTemplate() bool {
	return strings.TrimSpace(a.Template) != ""
}

func (a Action) DisplayValue() string {
	switch {
	case a.IsShow():
		return oneLine(a.Text, 90)
	case a.IsCmd():
		return oneLine(a.Cmd, 90)
	case a.IsTemplate():
		return oneLine(a.Template, 90)
	default:
		return ""
	}
}

type RunOption struct {
	Desc   string `yaml:"desc"`
	Run    string `yaml:"run"`
	Banner string `yaml:"banner"`
}

type RunCommands []RunOption

func (r *RunCommands) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case 0:
		*r = nil
		return nil
	case yaml.ScalarNode:
		var command string
		if err := node.Decode(&command); err != nil {
			return err
		}
		command = strings.TrimSpace(command)
		if command == "" {
			*r = nil
			return nil
		}
		*r = RunCommands{{Run: command}}
		return nil
	case yaml.SequenceNode:
		options := make([]RunOption, 0, len(node.Content))
		for _, item := range node.Content {
			switch item.Kind {
			case yaml.ScalarNode:
				var command string
				if err := item.Decode(&command); err != nil {
					return err
				}
				command = strings.TrimSpace(command)
				if command == "" {
					continue
				}
				options = append(options, RunOption{Run: command})
			case yaml.MappingNode:
				var option RunOption
				if err := item.Decode(&option); err != nil {
					return err
				}
				option.Desc = strings.TrimSpace(option.Desc)
				option.Run = strings.TrimSpace(option.Run)
				option.Banner = strings.TrimSpace(option.Banner)
				if option.Run == "" {
					return fmt.Errorf("run option at line %d is missing run", item.Line)
				}
				options = append(options, option)
			default:
				return fmt.Errorf("run must be a string or a list of strings/objects")
			}
		}
		*r = options
		return nil
	default:
		return fmt.Errorf("run must be a string or a list of strings/objects")
	}
}

type Entry struct {
	Desc       string      `yaml:"desc"`
	Text       string      `yaml:"text"`
	Cmd        string      `yaml:"cmd"`
	Mode       string      `yaml:"mode"`
	Actions    []Action    `yaml:"actions"`
	Run        RunCommands `yaml:"run"`
	Template   string      `yaml:"template"`
	Args       []Arg       `yaml:"args"`
	Note       string      `yaml:"note"`
	Banner     string      `yaml:"banner"`
	Tags       []string    `yaml:"tags"`
	SourcePath string      `yaml:"-"`
	SourceFile string      `yaml:"-"`
	SourceLine int         `yaml:"-"`
	searchData []weightedField `yaml:"-"`
	index      int
}

func (e Entry) PreviewHitLine(preview PreviewMatch, hitIndex int) int {
	if strings.TrimSpace(e.SourcePath) == "" {
		return e.SourceLine
	}
	if len(preview.Snippets) == 0 || hitIndex < 0 || hitIndex >= len(preview.Snippets) {
		return e.SourceLine
	}
	lineText := previewSnippetLine(preview.Snippets[hitIndex])
	if strings.TrimSpace(lineText) == "" {
		return e.SourceLine
	}
	raw, err := os.ReadFile(e.SourcePath)
	if err != nil {
		return e.SourceLine
	}
	lines := strings.Split(string(raw), "\n")
	start := e.SourceLine - 1
	if start < 0 {
		start = 0
	}
	for i := start; i < len(lines); i++ {
		if strings.Contains(lines[i], strings.TrimSpace(lineText)) {
			return i + 1
		}
	}
	return e.SourceLine
}

const TypeAll = "ALL"

func (e Entry) HasShow() bool {
	for _, action := range e.normalizedActions() {
		if action.IsShow() {
			return true
		}
	}
	return false
}

func (e Entry) HasNote() bool {
	return e.HasShow()
}

func (e Entry) HasCmd() bool {
	for _, action := range e.normalizedActions() {
		if action.IsCmd() {
			return true
		}
	}
	return false
}

func (e Entry) IsRun() bool {
	return e.HasCmd()
}

func (e Entry) HasTemplate() bool {
	for _, action := range e.normalizedActions() {
		if action.IsTemplate() {
			return true
		}
	}
	return false
}

func (e Entry) IsTemplate() bool {
	return e.HasTemplate()
}

func (e Entry) ActionCount() int {
	return len(e.normalizedActions())
}

func (e Entry) ActionsList() []Action {
	return e.normalizedActions()
}

func (e Entry) PrimaryAction() *Action {
	actions := e.normalizedActions()
	if len(actions) == 0 {
		return nil
	}
	return &actions[0]
}

func (e Entry) CmdActions() []Action {
	actions := e.normalizedActions()
	out := make([]Action, 0, len(actions))
	for _, action := range actions {
		if action.IsCmd() {
			out = append(out, action)
		}
	}
	return out
}

func (e Entry) RunCount() int {
	return len(e.CmdActions())
}

func (e Entry) RunOptions() []RunOption {
	cmds := e.CmdActions()
	out := make([]RunOption, 0, len(cmds))
	for _, action := range cmds {
		out = append(out, RunOption{
			Desc:   action.Desc,
			Run:    action.Cmd,
			Banner: action.Banner,
		})
	}
	return out
}

func (e Entry) PrimaryRun() string {
	cmds := e.CmdActions()
	if len(cmds) == 0 {
		return ""
	}
	return strings.TrimSpace(cmds[0].Cmd)
}

func (e Entry) PrimaryBanner() string {
	cmds := e.CmdActions()
	if len(cmds) == 0 {
		return ""
	}
	return strings.TrimSpace(cmds[0].Banner)
}

func (e Entry) FirstShowAction() *Action {
	actions := e.normalizedActions()
	for i := range actions {
		if actions[i].IsShow() {
			return &actions[i]
		}
	}
	return nil
}

func (e Entry) Action() string {
	if e.IsLite() {
		switch {
		case strings.TrimSpace(e.Cmd) != "":
			return "RUN"
		case strings.TrimSpace(e.Template) != "":
			return "TEMPLATE"
		case strings.TrimSpace(e.Text) != "":
			return "SHOW"
		}
	}

	parts := make([]string, 0, 3)
	if e.HasShow() {
		parts = append(parts, "SHOW")
	}
	cmdCount := len(e.CmdActions())
	if cmdCount > 0 {
		if cmdCount == 1 {
			parts = append(parts, "RUN")
		} else {
			parts = append(parts, fmt.Sprintf("RUN x%d", cmdCount))
		}
	}
	if e.HasTemplate() {
		parts = append(parts, "TEMPLATE")
	}
	if len(parts) == 0 {
		return TypeAll
	}
	return strings.Join(parts, " + ")
}

func (e Entry) DisplayValue() string {
	if e.IsLite() {
		switch {
		case strings.TrimSpace(e.Cmd) != "":
			return oneLine(e.Cmd, 90)
		case strings.TrimSpace(e.Template) != "":
			return oneLine(e.Template, 90)
		case strings.TrimSpace(e.Text) != "":
			return oneLine(e.Text, 90)
		}
	}

	actions := e.normalizedActions()
	if len(actions) == 0 {
		return ""
	}
	value := actions[0].DisplayValue()
	if len(actions) > 1 {
		value = fmt.Sprintf("%s (+%d more)", value, len(actions)-1)
	}
	return oneLine(value, 90)
}

func (e Entry) Title() string {
	base := strings.TrimSuffix(e.SourceFile, filepath.Ext(e.SourceFile))
	return fmt.Sprintf("%s [%s]", e.Desc, base)
}

func (e Entry) SearchFields() []string {
	fields := []string{e.Desc, e.SourceFile, e.Text, e.Cmd}
	for _, action := range e.normalizedActions() {
		fields = append(fields, action.Desc, action.Cmd, action.Text, action.Template, action.Banner)
		for _, arg := range action.Args {
			fields = append(fields, arg.Name, arg.Prompt, arg.Default, arg.Example, arg.Description)
		}
	}
	fields = append(fields, e.Tags...)
	return fields
}

func oneLine(text string, limit int) string {
	text = strings.TrimSpace(strings.Join(strings.Fields(text), " "))
	if len(text) <= limit {
		return text
	}
	return text[:limit-3] + "..."
}

func (e Entry) IsLite() bool {
	return strings.TrimSpace(e.Text) != "" || strings.TrimSpace(e.Cmd) != "" || strings.TrimSpace(e.Template) != ""
}

func (e Entry) QuickAction() *Action {
	if e.IsLite() {
		switch {
		case strings.TrimSpace(e.Cmd) != "":
			return &Action{
				Desc:   inferActionDesc(e.Cmd, "run"),
				Cmd:    strings.TrimSpace(e.Cmd),
				Banner: strings.TrimSpace(e.Banner),
			}
		case strings.TrimSpace(e.Template) != "":
			return &Action{
				Desc:     inferActionDesc(e.Template, "run"),
				Template: strings.TrimSpace(e.Template),
				Args:     e.Args,
				Banner:   strings.TrimSpace(e.Banner),
			}
		case strings.TrimSpace(e.Text) != "":
			return &Action{
				Desc: "show",
				Text: strings.TrimSpace(e.Text),
			}
		}
	}

	actions := e.normalizedActions()
	if len(actions) != 1 {
		return nil
	}
	return &actions[0]
}

func (e Entry) normalizedActions() []Action {
	if len(e.Actions) > 0 {
		return e.Actions
	}

	actions := make([]Action, 0, 1+len(e.Run)+2)
	if cmd := strings.TrimSpace(e.Cmd); cmd != "" {
		actions = append(actions, Action{
			Desc:   inferActionDesc(cmd, "run"),
			Cmd:    cmd,
			Banner: strings.TrimSpace(e.Banner),
		})
	}
	if text := strings.TrimSpace(e.Text); text != "" {
		actions = append(actions, Action{
			Desc: "show",
			Text: text,
		})
	}
	if text := strings.TrimSpace(e.Note); text != "" {
		actions = append(actions, Action{
			Desc: "show",
			Text: text,
		})
	}
	for i, option := range e.Run {
		desc := strings.TrimSpace(option.Desc)
		if desc == "" {
			desc = inferActionDesc(option.Run, fmt.Sprintf("run %d", i+1))
		}
		banner := strings.TrimSpace(option.Banner)
		if banner == "" && i == 0 {
			banner = strings.TrimSpace(e.Banner)
		}
		actions = append(actions, Action{
			Desc:   desc,
			Cmd:    strings.TrimSpace(option.Run),
			Banner: banner,
		})
	}
	if template := strings.TrimSpace(e.Template); template != "" {
		actions = append(actions, Action{
			Desc:     inferActionDesc(template, "template"),
			Template: template,
			Args:     e.Args,
			Banner:   strings.TrimSpace(e.Banner),
		})
	}
	return actions
}
