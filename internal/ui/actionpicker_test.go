package ui

import (
	"strings"
	"testing"

	"aoo/internal/notes"
	"github.com/charmbracelet/lipgloss"
)

func TestEntryActionsPreferShowFirst(t *testing.T) {
	entry := notes.Entry{
		Desc: "himki",
		Actions: []notes.Action{
			{Desc: "show", Text: "Host notes"},
			{Desc: "ssh vyos", Cmd: "ssh vyos@himki"},
			{Desc: "ssh tunnel", Cmd: "ssh -L 8443:10.0.0.1:443 vyos@himki"},
		},
	}

	actions := entryActions(entry)
	if len(actions) != 3 {
		t.Fatalf("expected 3 actions, got %d", len(actions))
	}
	if actions[0].Kind != ActionRead {
		t.Fatalf("expected first action to be show, got %s", actions[0].Kind)
	}
	if actions[0].Label != "show note" {
		t.Fatalf("expected first action label to be show note, got %q", actions[0].Label)
	}
	if actions[1].Kind != ActionRun {
		t.Fatalf("expected second action to be run, got %s", actions[1].Kind)
	}
	if actions[1].Label != "ssh vyos" {
		t.Fatalf("unexpected run label: %q", actions[1].Label)
	}
}

func TestEntryActionsLabelRawFileWithExtension(t *testing.T) {
	entry := notes.Entry{
		Desc:       "netplan prod",
		SourceFile: "netplan-prod.yaml",
		SourceKind: notes.SourceKindRaw,
		Actions: []notes.Action{{
			Desc: "yaml",
			Text: "network:\n  version: 2\n",
		}},
	}

	actions := entryActions(entry)
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	if actions[0].Label != "show yaml" {
		t.Fatalf("expected raw action label to include extension, got %q", actions[0].Label)
	}
}

func TestActionHintTextShowsCommandAndTextActions(t *testing.T) {
	model := ActionPickerModel{
		options: Options{},
	}

	commandLabel := model.enterHintText(EntryAction{
		Kind:  ActionRun,
		Label: "ssh vyos",
	})
	if !strings.Contains(commandLabel, "enter: run command") {
		t.Fatalf("expected command hint, got %q", commandLabel)
	}

	textLabel := model.enterHintText(EntryAction{
		Kind:  ActionRead,
		Label: "show note",
	})
	if !strings.Contains(textLabel, "enter: print note") {
		t.Fatalf("expected text hint, got %q", textLabel)
	}
}

func TestActionDetailLineShowsHint(t *testing.T) {
	model := ActionPickerModel{
		theme:   Theme{HelpFG: "8"},
		options: Options{},
	}
	got := model.detailLine("ssh vyos@himki", "enter: run command", 80, lipgloss.NewStyle(), lipgloss.NewStyle())
	if !strings.Contains(got, "enter: run command") {
		t.Fatalf("expected inline hint in detail line, got %q", got)
	}
}

func TestActionPickerRendersHintsForAllVisibleActions(t *testing.T) {
	entry := notes.Entry{
		Desc: "himki",
		Actions: []notes.Action{
			{Desc: "show", Text: "Host notes"},
			{Desc: "ssh vyos", Cmd: "ssh vyos@himki"},
		},
	}

	model := ActionPickerModel{
		entry:   entry,
		actions: entryActions(entry),
		width:   80,
		height:  10,
		theme:   Theme{SelectedMark: ">", RowFG: "7", SelectedFG: "15", SelectedBG: "0", HelpFG: "8"},
		options: Options{},
	}

	view := model.View()
	if !strings.Contains(view, "enter: print note") {
		t.Fatalf("expected text hint in action picker view, got %q", view)
	}
	if !strings.Contains(view, "enter: run command") {
		t.Fatalf("expected command hint in action picker view, got %q", view)
	}
}

func TestActionPickerFocusModeHidesFooter(t *testing.T) {
	entry := notes.Entry{
		Desc: "himki",
		Actions: []notes.Action{
			{Desc: "show", Text: "Host notes"},
		},
	}

	model := ActionPickerModel{
		entry:   entry,
		actions: entryActions(entry),
		width:   80,
		height:  10,
		theme:   Theme{SelectedMark: ">", RowFG: "7", SelectedFG: "15", SelectedBG: "0", HelpFG: "8"},
		options: Options{FocusMode: true},
	}

	view := model.View()
	if strings.Contains(view, "enter choose") {
		t.Fatalf("expected focus mode to hide action footer, got %q", view)
	}
}

func TestActionDetailLineShrinksHintOnNarrowWidth(t *testing.T) {
	model := ActionPickerModel{}
	line := model.detailLine(
		"ssh -L 8443:10.0.0.1:443 vyos@himki",
		"enter: run command",
		18,
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
	)
	if lipgloss.Width(line) > 18 {
		t.Fatalf("expected detail line to fit narrow width, got %d: %q", lipgloss.Width(line), line)
	}
}
