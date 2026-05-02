package bundled

import (
	_ "embed"

	"aoo/internal/notes"
)

//go:embed service.yaml
var serviceYAML []byte

//go:embed command_templates.yaml
var templatesYAML []byte

func Load() notes.LoadResult {
	service := notes.LoadBytes("builtin/service.yaml", serviceYAML)
	templates := notes.LoadBytes("builtin/command_templates.yaml", templatesYAML)

	return notes.LoadResult{
		Entries: append(service.Entries, templates.Entries...),
		Errors:  append(service.Errors, templates.Errors...),
	}
}
