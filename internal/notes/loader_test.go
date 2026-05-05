package notes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadDirSkipsHiddenYAML(t *testing.T) {
	workingDir := t.TempDir()

	noteFile := filepath.Join(workingDir, "notes.yaml")
	if err := os.WriteFile(noteFile, []byte("- desc: router\n  note: |\n    ok\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	hiddenFile := filepath.Join(workingDir, ".goreleaser.yaml")
	if err := os.WriteFile(hiddenFile, []byte("project_name: term-notes\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	result := LoadDir(workingDir)
	if len(result.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", result.Errors)
	}

	if len(result.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result.Entries))
	}
}

func TestLoadBytesAcceptsActions(t *testing.T) {
	result := LoadBytes("notes.yaml", []byte(`
- desc: show interfaces
  actions:
    - desc: show
      text: |
        show interfaces
    - desc: ssh router
      cmd: ssh root@router
- desc: router note
  actions:
    - desc: ssh
      cmd: ssh root@router
`))

	if len(result.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", result.Errors)
	}
	if len(result.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result.Entries))
	}
	if !result.Entries[0].HasShow() || !result.Entries[0].HasCmd() {
		t.Fatalf("expected first entry to have both show and cmd actions")
	}
	if !result.Entries[1].HasCmd() {
		t.Fatalf("expected second entry to have cmd action")
	}
}

func TestLoadBytesRejectsInvalidMode(t *testing.T) {
	result := LoadBytes("notes.yaml", []byte(`
- desc: invalid
  mode: template
  template: echo hi
`))

	if len(result.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(result.Errors))
	}
	if !strings.Contains(result.Errors[0].Error(), "invalid mode") {
		t.Fatalf("expected invalid mode error, got %v", result.Errors[0])
	}
}

func TestLoadBytesSupportsLegacyFields(t *testing.T) {
	result := LoadBytes("notes.yaml", []byte(`
- desc: mixed show
  note: hello
  run: echo hi
`))

	if len(result.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", result.Errors)
	}
	if !result.Entries[0].HasShow() || !result.Entries[0].HasCmd() {
		t.Fatalf("expected legacy fields to normalize into actions")
	}
}

func TestLoadBytesAcceptsLiteFields(t *testing.T) {
	result := LoadBytes("notes.yaml", []byte(`
- desc: router ssh
  text: main access
  action: ssh root@router
`))

	if len(result.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", result.Errors)
	}
	entry := result.Entries[0]
	if !entry.HasShow() || !entry.HasCmd() {
		t.Fatalf("expected lite fields to normalize into actions")
	}
	if action := entry.QuickAction(); action == nil || !action.IsCmd() {
		t.Fatalf("expected quick action to prefer cmd, got %#v", action)
	}
	if got := entry.ActionsList()[0].Desc; got != "" {
		t.Fatalf("expected canonical action field to avoid duplicated desc, got %q", got)
	}
}

func TestLoadBytesAcceptsMultipleCommandActions(t *testing.T) {
	result := LoadBytes("notes.yaml", []byte(`
- desc: himki
  actions:
    - desc: ssh vyos
      cmd: ssh vyos@himki
    - desc: ssh tunnel
      cmd: ssh -L 8443:10.0.0.1:443 vyos@himki
      banner: https://127.0.0.1:8443
`))

	if len(result.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", result.Errors)
	}
	if len(result.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result.Entries))
	}
	entry := result.Entries[0]
	if got := entry.RunCount(); got != 2 {
		t.Fatalf("expected 2 cmd actions, got %d", got)
	}
	if got := entry.RunOptions()[1].Desc; got != "ssh tunnel" {
		t.Fatalf("unexpected second command desc: %q", got)
	}
	if got := entry.PrimaryRun(); got != "ssh vyos@himki" {
		t.Fatalf("unexpected primary command: %q", got)
	}
}
