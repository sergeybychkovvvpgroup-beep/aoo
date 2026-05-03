package ui

import (
	"testing"

	"aoo/internal/notes"
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
