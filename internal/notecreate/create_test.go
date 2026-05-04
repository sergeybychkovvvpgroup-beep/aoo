package notecreate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateCommandDraft(t *testing.T) {
	root := t.TempDir()

	draft, err := Create(root, KindCommand, "router jump")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasSuffix(draft.Path, "router-jump.yaml") {
		t.Fatalf("unexpected draft path: %s", draft.Path)
	}
	if draft.Line != 5 {
		t.Fatalf("expected edit line 5, got %d", draft.Line)
	}

	raw, err := os.ReadFile(draft.Path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	if !strings.Contains(text, "desc: router jump") {
		t.Fatalf("expected desc in draft, got %q", text)
	}
	if !strings.Contains(text, "cmd: |") {
		t.Fatalf("expected cmd scaffold, got %q", text)
	}
	if strings.Contains(text, "actions:") {
		t.Fatalf("expected flat scaffold without actions, got %q", text)
	}
}

func TestCreateUsesUniqueFilename(t *testing.T) {
	root := t.TempDir()
	first := filepath.Join(root, "router.yaml")
	if err := os.WriteFile(first, []byte("desc: router\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	draft, err := Create(root, KindNote, "router")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasSuffix(draft.Path, "router-2.yaml") {
		t.Fatalf("unexpected unique path: %s", draft.Path)
	}
}
