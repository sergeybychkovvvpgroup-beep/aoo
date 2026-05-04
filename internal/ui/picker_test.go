package ui

import (
	"strings"
	"testing"

	"aoo/internal/notes"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

func TestBottomLayoutPlacesInputAtBottom(t *testing.T) {
	model := PickerModel{
		input: textInputWithValue("static"),
		entries: []notes.Entry{{
			Desc: "router dhcp",
			Actions: []notes.Action{{
				Desc: "show",
				Text: "static-map",
			}},
		}},
		matches: []notes.Match{{
			Entry: notes.Entry{
				Desc: "router dhcp",
				Actions: []notes.Action{{
					Desc: "show",
					Text: "static-map",
				}},
			},
			Label:  "router dhcp",
			Detail: "SHOW | static-map",
		}},
		height:  8,
		width:   80,
		theme:   Theme{SelectedMark: ">", RowFG: "7", SelectedFG: "15", SelectedBG: "0"},
		options: Options{Layout: "bottom"},
	}

	view := model.View()
	lines := strings.Split(view, "\n")
	if len(lines) < 2 {
		t.Fatalf("expected multiline view, got %q", view)
	}
	if !strings.Contains(lines[len(lines)-1], "notes>") {
		t.Fatalf("expected prompt on last line, got %q", lines[len(lines)-1])
	}
}

func TestStatusLineShowsOnlyMatchesAndTotal(t *testing.T) {
	model := PickerModel{
		entries: make([]notes.Entry, 64),
		matches: make([]notes.Match, 2),
	}

	if got := model.statusLine(); got != "2/64" {
		t.Fatalf("unexpected status line: %q", got)
	}
}

func TestStatusLineShowsActiveHitIndexWhenMultiplePreviewHitsExist(t *testing.T) {
	entry := notes.Entry{
		Desc: "router cha",
		Actions: []notes.Action{{
			Desc: "show",
			Text: "static one\nstatic two\nstatic three\nstatic four\nstatic five\n",
		}},
	}

	model := PickerModel{
		entries: []notes.Entry{entry},
		matches: []notes.Match{{
			Entry:  entry,
			Label:  entry.Desc,
			Detail: entry.Action() + " | " + entry.DisplayValue(),
		}},
		previewHit: 1,
	}
	model.preview = notes.BuildPreview(entry, "static")

	if got := model.statusLine(); got != "1/1  hit 2/5" {
		t.Fatalf("unexpected status line with hit index: %q", got)
	}
}

func TestStatusLineHidesHitIndexWhenPreviewLineIsDisabled(t *testing.T) {
	entry := notes.Entry{
		Desc: "router cha",
		Actions: []notes.Action{{
			Desc: "show",
			Text: "static one\nstatic two\nstatic three\n",
		}},
	}

	model := PickerModel{
		entries: []notes.Entry{entry},
		matches: []notes.Match{{
			Entry:  entry,
			Label:  entry.Desc,
			Detail: entry.Action() + " | " + entry.DisplayValue(),
		}},
		previewHit: 1,
		options:    Options{ShowPreview: false},
	}
	model.preview = notes.BuildPreview(entry, "static")

	if got := model.statusLine(); got != "1/1" {
		t.Fatalf("expected no hit index without preview line, got %q", got)
	}
}

func TestResultLinesDoNotRenderExtraInlinePreviewBlockForSelectedEntry(t *testing.T) {
	entry := notes.Entry{
		Desc: "aoo-help",
		Actions: []notes.Action{{
			Desc: "show",
			Text: "aoo = terminal notes + command launcher.\n\nStart:\n  aoo\n  aoo --query ssh\n",
		}},
	}

	model := PickerModel{
		matches: []notes.Match{{
			Entry:  entry,
			Label:  entry.Desc,
			Detail: entry.Action() + " | " + entry.DisplayValue(),
		}},
		cursor:  0,
		theme:   Theme{SelectedMark: ">", RowFG: "7", SelectedFG: "15", SelectedBG: "0"},
		options: Options{ShowPreview: true},
	}
	model.preview = notes.BuildPreview(entry, "start")

	lines := model.resultLines(
		80,
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
	)
	if len(lines) != 2 {
		t.Fatalf("expected selected entry to render as 2 lines, got %d: %#v", len(lines), lines)
	}
}

func TestResultLinesDoNotRenderExtraInlinePreviewBlockForCmdOnlyEntry(t *testing.T) {
	entry := notes.Entry{
		Desc: "git add-commit-push",
		Actions: []notes.Action{{
			Desc: "run",
			Cmd:  `git add . && git commit -m "periodic commit" && git push -u origin main`,
		}},
	}

	model := PickerModel{
		matches: []notes.Match{{
			Entry:  entry,
			Label:  entry.Desc,
			Detail: entry.Action() + " | " + entry.DisplayValue(),
		}},
		cursor:  0,
		theme:   Theme{SelectedMark: ">", RowFG: "7", SelectedFG: "15", SelectedBG: "0"},
		options: Options{ShowPreview: true},
	}
	model.preview = notes.BuildPreview(entry, "git")

	lines := model.resultLines(
		100,
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
	)
	if len(lines) != 2 {
		t.Fatalf("expected cmd-only entry to render as 2 lines, got %d: %#v", len(lines), lines)
	}
}

func TestSelectedPreviewLineUsesActivePreviewHit(t *testing.T) {
	entry := notes.Entry{
		Desc: "router cha",
		Actions: []notes.Action{{
			Desc: "show",
			Text: "set protocols static route 0.0.0.0/0 next-hop 1.1.1.1\n" +
				"set service dhcp-server subnet 10.120.0.0/16 static-mapping vyos-lan ip-address '10.120.0.5'\n",
		}},
	}

	model := PickerModel{
		matches: []notes.Match{{
			Entry:  entry,
			Label:  entry.Desc,
			Detail: entry.Action() + " | " + entry.DisplayValue(),
		}},
		cursor:     0,
		previewHit: 1,
		theme:      Theme{SelectedMark: ">", RowFG: "7", SelectedFG: "15", SelectedBG: "0"},
		options:    Options{ShowPreview: true},
	}
	model.preview = notes.BuildPreview(entry, "static")

	lines := model.resultLines(
		100,
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
	)
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %#v", len(lines), lines)
	}
	if !strings.Contains(lines[1], "static-mapping") {
		t.Fatalf("expected preview line for active hit, got %q", lines[1])
	}
}

func textInputWithValue(value string) textinput.Model {
	input := textinput.New()
	input.Prompt = "notes> "
	input.SetValue(value)
	return input
}
