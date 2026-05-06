package ui

import (
	"strings"
	"testing"

	"aoo/internal/notes"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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
			Detail: "static-map",
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
	lastLines := strings.Join(lines[maxInt(0, len(lines)-4):], "\n")
	if !strings.Contains(lastLines, "esc quit") {
		t.Fatalf("expected help block near bottom, got %q", lastLines)
	}
}

func TestViewHidesResultsOnEmptyQueryByDefault(t *testing.T) {
	model := PickerModel{
		input: textInputWithValue(""),
		matches: []notes.Match{
			{Entry: notes.Entry{Desc: "router notes"}, Label: "router notes", Detail: "dhcp"},
		},
		height:  8,
		width:   80,
		theme:   Theme{SelectedMark: ">", RowFG: "7", SelectedFG: "15", SelectedBG: "0"},
		options: Options{Layout: "bottom"},
	}

	view := model.View()
	if strings.Contains(view, "router notes") {
		t.Fatalf("expected no visible notes on empty query by default, got %q", view)
	}
}

func TestViewShowsResultsOnEmptyQueryWhenEnabled(t *testing.T) {
	model := PickerModel{
		input: textInputWithValue(""),
		matches: []notes.Match{
			{Entry: notes.Entry{Desc: "router notes"}, Label: "router notes", Detail: "dhcp"},
		},
		height:  8,
		width:   80,
		theme:   Theme{SelectedMark: ">", RowFG: "7", SelectedFG: "15", SelectedBG: "0"},
		options: Options{Layout: "bottom", ShowListOnStart: true},
	}

	view := model.View()
	if !strings.Contains(view, "router notes") {
		t.Fatalf("expected visible notes on empty query, got %q", view)
	}
}

func TestBottomLayoutRendersFirstResultAtBottomOfResultBlock(t *testing.T) {
	model := PickerModel{
		input: textInputWithValue("vyos"),
		matches: []notes.Match{
			{Entry: notes.Entry{Desc: "first"}, Label: "first", Detail: "one"},
			{Entry: notes.Entry{Desc: "second"}, Label: "second", Detail: "two"},
			{Entry: notes.Entry{Desc: "third"}, Label: "third", Detail: "three"},
		},
		height:  10,
		width:   80,
		theme:   Theme{SelectedMark: ">", RowFG: "7", SelectedFG: "15", SelectedBG: "0"},
		options: Options{Layout: "bottom"},
	}

	lines := model.resultLines(
		80,
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
	)
	if len(lines) != 6 {
		t.Fatalf("expected 6 result lines, got %d: %#v", len(lines), lines)
	}
	if !strings.Contains(lines[0], "third") {
		t.Fatalf("expected top visible line to contain last match, got %q", lines[0])
	}
	if !strings.Contains(lines[4], "first") {
		t.Fatalf("expected bottom visible label line to contain first match, got %q", lines[4])
	}
}

func TestResultLinesDoNotRenderLegacyTypeBadges(t *testing.T) {
	model := PickerModel{
		input: textInputWithValue("vyos"),
		matches: []notes.Match{
			{Entry: notes.Entry{Desc: "vyos chashnikovo"}, Label: "vyos chashnikovo", Detail: "ssh vyos-volga"},
		},
		height:  8,
		width:   80,
		theme:   Theme{SelectedMark: ">", RowFG: "7", SelectedFG: "15", SelectedBG: "0"},
		options: Options{Layout: "bottom"},
	}

	lines := model.resultLines(
		80,
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
	)
	if len(lines) == 0 {
		t.Fatal("expected result lines")
	}
	if strings.Contains(lines[0], "[cmd]") || strings.Contains(lines[0], "[txt]") || strings.Contains(lines[0], "[tpl]") {
		t.Fatalf("expected result line without legacy type badges, got %q", lines[0])
	}
}

func TestBottomLayoutInvertsArrowNavigation(t *testing.T) {
	model := PickerModel{
		input: textInputWithValue("vyos"),
		matches: []notes.Match{
			{Entry: notes.Entry{Desc: "first"}, Label: "first", Detail: "one"},
			{Entry: notes.Entry{Desc: "second"}, Label: "second", Detail: "two"},
			{Entry: notes.Entry{Desc: "third"}, Label: "third", Detail: "three"},
		},
		cursor:  0,
		theme:   Theme{SelectedMark: ">", RowFG: "7", SelectedFG: "15", SelectedBG: "0"},
		options: Options{Layout: "bottom"},
	}

	nextModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyUp})
	updated, ok := nextModel.(PickerModel)
	if !ok {
		t.Fatalf("expected PickerModel after key update")
	}
	if updated.cursor != 1 {
		t.Fatalf("expected up key to move selection visually upward in bottom layout, got cursor %d", updated.cursor)
	}

	nextModel, _ = updated.Update(tea.KeyMsg{Type: tea.KeyDown})
	updated, ok = nextModel.(PickerModel)
	if !ok {
		t.Fatalf("expected PickerModel after key update")
	}
	if updated.cursor != 0 {
		t.Fatalf("expected down key to move selection visually downward in bottom layout, got cursor %d", updated.cursor)
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

func TestStatusLineShowsEnterRunHint(t *testing.T) {
	entry := notes.Entry{
		Desc: "ssh office-notebook",
		Actions: []notes.Action{{
			Desc: "ssh",
			Cmd:  "ssh user@host",
		}},
	}

	model := PickerModel{
		entries: []notes.Entry{entry},
		matches: []notes.Match{{
			Entry:  entry,
			Label:  entry.Desc,
			Detail: entry.DisplayValue(),
		}},
	}

	if got := model.enterHintText(entry); got != "enter: run command" {
		t.Fatalf("unexpected enter hint: %q", got)
	}
}

func TestStatusLineShowsEnterShowHint(t *testing.T) {
	entry := notes.Entry{
		Desc: "router notes",
		Actions: []notes.Action{{
			Desc: "show",
			Text: "dhcp notes",
		}},
	}

	model := PickerModel{
		entries: []notes.Entry{entry},
		matches: []notes.Match{{
			Entry:  entry,
			Label:  entry.Desc,
			Detail: entry.DisplayValue(),
		}},
	}

	if got := model.enterHintText(entry); got != "enter: print note" {
		t.Fatalf("unexpected enter hint: %q", got)
	}
}

func TestStatusLineShowsEnterRawFileHint(t *testing.T) {
	entry := notes.Entry{
		Desc:       "netplan prod",
		SourceFile: "netplan-prod.yaml",
		SourceKind: notes.SourceKindRaw,
		Actions: []notes.Action{{
			Desc: "yaml",
			Text: "network:\n  version: 2\n",
		}},
	}

	model := PickerModel{
		entries: []notes.Entry{entry},
		matches: []notes.Match{{
			Entry:  entry,
			Label:  entry.Desc,
			Detail: entry.DisplayValue(),
		}},
	}

	if got := model.enterHintText(entry); got != "enter: print yaml" {
		t.Fatalf("unexpected raw hint: %q", got)
	}
}

func TestStatusLineShowsEnterActionsHint(t *testing.T) {
	entry := notes.Entry{
		Desc: "vyos chashnikovo",
		Actions: []notes.Action{
			{Desc: "ssh", Cmd: "ssh user@host"},
			{Desc: "show", Text: "dhcp"},
			{Desc: "https", Cmd: "ssh -L 8443:127.0.0.1:443 user@host"},
		},
	}

	model := PickerModel{
		entries: []notes.Entry{entry},
		matches: []notes.Match{{
			Entry:  entry,
			Label:  entry.Desc,
			Detail: entry.DisplayValue(),
		}},
	}

	if got := model.enterHintText(entry); got != "enter: select action" {
		t.Fatalf("unexpected enter hint: %q", got)
	}
}

func TestPickerHelpTextUsesStaticOpenLabel(t *testing.T) {
	if got := pickerHelpText(); !strings.Contains(got, "enter open") {
		t.Fatalf("expected static enter label in help text, got %q", got)
	}
}

func TestResultLinesRenderBadgeLineBelowSelectedEntry(t *testing.T) {
	entry := notes.Entry{
		Desc: "vyos chashnikovo",
		Actions: []notes.Action{
			{Desc: "ssh", Cmd: "ssh user@host"},
			{Desc: "show", Text: "dhcp"},
		},
	}

	model := PickerModel{
		input: textInputWithValue("vyos"),
		matches: []notes.Match{{
			Entry:  entry,
			Label:  entry.Desc,
			Detail: entry.DisplayValue(),
		}},
		cursor:  0,
		width:   80,
		theme:   Theme{SelectedMark: ">", RowFG: "7", SelectedFG: "15", SelectedBG: "0", HelpFG: "8"},
		options: Options{ShowMatchContext: true},
	}

	lines := model.resultLines(80, lipgloss.NewStyle(), lipgloss.NewStyle(), lipgloss.NewStyle(), lipgloss.NewStyle())
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines with inline hint, got %d: %#v", len(lines), lines)
	}
	if !strings.Contains(lines[1], "enter: select action") {
		t.Fatalf("expected detail line with enter action, got %q", lines[1])
	}
}

func TestResultLinesRenderBadgeLineForNonSelectedEntry(t *testing.T) {
	entryA := notes.Entry{
		Desc: "vyos chashnikovo",
		Actions: []notes.Action{
			{Desc: "ssh", Cmd: "ssh user@host"},
			{Desc: "show", Text: "dhcp"},
		},
	}
	entryB := notes.Entry{
		Desc: "router notes",
		Actions: []notes.Action{{
			Desc: "show",
			Text: "dhcp notes",
		}},
	}

	model := PickerModel{
		input: textInputWithValue("v"),
		matches: []notes.Match{
			{Entry: entryA, Label: entryA.Desc, Detail: entryA.DisplayValue()},
			{Entry: entryB, Label: entryB.Desc, Detail: entryB.DisplayValue()},
		},
		cursor:  0,
		width:   80,
		theme:   Theme{SelectedMark: ">", RowFG: "7", SelectedFG: "15", SelectedBG: "0", HelpFG: "8"},
		options: Options{ShowMatchContext: false},
	}

	lines := model.resultLines(80, lipgloss.NewStyle(), lipgloss.NewStyle(), lipgloss.NewStyle(), lipgloss.NewStyle())
	if len(lines) != 4 {
		t.Fatalf("expected 4 lines with hints for all results, got %d: %#v", len(lines), lines)
	}
	if !strings.Contains(lines[1], "enter: select action") {
		t.Fatalf("expected first hint line, got %q", lines[1])
	}
	if !strings.Contains(lines[3], "enter: print note") {
		t.Fatalf("expected second hint line, got %q", lines[3])
	}
}

func TestBadgeTextCanBeDisabledIndependently(t *testing.T) {
	entry := notes.Entry{
		Desc: "router notes",
		Actions: []notes.Action{
			{Desc: "show", Text: "dhcp notes"},
			{Desc: "ssh", Cmd: "ssh user@host"},
		},
	}

	model := PickerModel{
		input: textInputWithValue("router"),
		matches: []notes.Match{{
			Entry:  entry,
			Label:  entry.Desc,
			Detail: entry.DisplayValue(),
		}},
		cursor:  0,
		width:   80,
		theme:   Theme{SelectedMark: ">", RowFG: "7", SelectedFG: "15", SelectedBG: "0", HelpFG: "8"},
		options: Options{ShowMatchContext: false},
	}

	got := model.enterHintText(entry)
	if got != "enter: select action" {
		t.Fatalf("expected stable enter hint, got %q", got)
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
		input:   textInputWithValue("static"),
		entries: []notes.Entry{entry},
		matches: []notes.Match{{
			Entry:  entry,
			Label:  entry.Desc,
			Detail: entry.DisplayValue(),
		}},
		previewHit: 1,
		options:    Options{ShowMatchContext: true},
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
			Detail: entry.DisplayValue(),
		}},
		previewHit: 1,
		options:    Options{ShowMatchContext: false},
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
		input: textInputWithValue("start"),
		matches: []notes.Match{{
			Entry:  entry,
			Label:  entry.Desc,
			Detail: entry.DisplayValue(),
		}},
		cursor:  0,
		theme:   Theme{SelectedMark: ">", RowFG: "7", SelectedFG: "15", SelectedBG: "0"},
		options: Options{ShowMatchContext: true},
	}
	model.preview = notes.BuildPreview(entry, "start")

	lines := model.resultLines(
		80,
		lipgloss.NewStyle(),
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
		input: textInputWithValue("git"),
		matches: []notes.Match{{
			Entry:  entry,
			Label:  entry.Desc,
			Detail: entry.DisplayValue(),
		}},
		cursor:  0,
		theme:   Theme{SelectedMark: ">", RowFG: "7", SelectedFG: "15", SelectedBG: "0"},
		options: Options{ShowMatchContext: true},
	}
	model.preview = notes.BuildPreview(entry, "git")

	lines := model.resultLines(
		100,
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
	)
	if len(lines) != 2 {
		t.Fatalf("expected cmd-only entry to render as 2 lines, got %d: %#v", len(lines), lines)
	}
	if !strings.Contains(lines[1], "git add . && git commit") {
		t.Fatalf("expected selected cmd preview to stay on command text, got %q", lines[1])
	}
	if strings.Contains(lines[1], entry.Desc) {
		t.Fatalf("expected selected cmd preview not to switch to description, got %q", lines[1])
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
		input: textInputWithValue("static"),
		matches: []notes.Match{{
			Entry:  entry,
			Label:  entry.Desc,
			Detail: entry.DisplayValue(),
		}},
		cursor:     0,
		previewHit: 1,
		theme:      Theme{SelectedMark: ">", RowFG: "7", SelectedFG: "15", SelectedBG: "0"},
		options:    Options{ShowMatchContext: true},
	}
	model.preview = notes.BuildPreview(entry, "static")

	lines := model.resultLines(
		100,
		lipgloss.NewStyle(),
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

func TestSelectedCommandPreviewDoesNotDuplicatePrefix(t *testing.T) {
	entry := notes.Entry{
		Desc: "aoo notes git add commit push",
		Actions: []notes.Action{{
			Desc: "run",
			Cmd:  `git -C ~/.local/share/aoo/notes add . && git -C ~/.local/share/aoo/notes commit -m "update" && git -C ~/.local/share/aoo/notes push`,
		}},
	}

	model := PickerModel{
		input: textInputWithValue("push"),
		matches: []notes.Match{{
			Entry:  entry,
			Label:  entry.Desc,
			Detail: entry.DisplayValue(),
		}},
		cursor:  0,
		theme:   Theme{SelectedMark: ">", RowFG: "7", SelectedFG: "15", SelectedBG: "0"},
		options: Options{ShowMatchContext: true},
	}
	model.preview = notes.BuildPreview(entry, "push")

	lines := model.resultLines(
		120,
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
	)
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %#v", len(lines), lines)
	}
	if strings.Contains(lines[1], "RUN |") {
		t.Fatalf("expected selected preview without duplicated RUN prefix, got %q", lines[1])
	}
}

func textInputWithValue(value string) textinput.Model {
	input := textinput.New()
	input.Prompt = "notes> "
	input.SetValue(value)
	return input
}
