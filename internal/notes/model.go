package notes

import (
	"fmt"
	"path/filepath"
	"strings"
)

type Arg struct {
	Name        string `yaml:"name"`
	Prompt      string `yaml:"prompt"`
	Default     string `yaml:"default"`
	Example     string `yaml:"example"`
	Description string `yaml:"description"`
}

type Entry struct {
	Desc       string   `yaml:"desc"`
	Run        string   `yaml:"run"`
	Template   string   `yaml:"template"`
	Args       []Arg    `yaml:"args"`
	Note       string   `yaml:"note"`
	Banner     string   `yaml:"banner"`
	Check      string   `yaml:"check"`
	CheckError string   `yaml:"check_error"`
	Tags       []string `yaml:"tags"`
	SourcePath string   `yaml:"-"`
	SourceFile string   `yaml:"-"`
	index      int
}

func (e Entry) IsRun() bool {
	return strings.TrimSpace(e.Run) != ""
}

func (e Entry) IsTemplate() bool {
	return strings.TrimSpace(e.Template) != ""
}

func (e Entry) Action() string {
	if e.IsTemplate() {
		return "CMD"
	}
	if e.IsRun() {
		return "RUN"
	}
	return "NOTE"
}

func (e Entry) DisplayValue() string {
	value := e.Note
	if e.IsTemplate() {
		value = e.Template
	} else if e.IsRun() {
		value = e.Run
	}
	return oneLine(value, 90)
}

func (e Entry) Title() string {
	base := strings.TrimSuffix(e.SourceFile, filepath.Ext(e.SourceFile))
	return fmt.Sprintf("%s [%s]", e.Desc, base)
}

func (e Entry) SearchFields() []string {
	fields := []string{e.Desc, e.Run, e.Template, e.Note, e.Banner, e.Check, e.CheckError, e.SourceFile}
	for _, arg := range e.Args {
		fields = append(fields, arg.Name, arg.Prompt, arg.Default, arg.Example, arg.Description)
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
