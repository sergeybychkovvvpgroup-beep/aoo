package templatecmd

import "testing"

func TestRenderQuotesValues(t *testing.T) {
	cmd, err := Render(`find . -type f -name {{pattern}}`, map[string]string{
		"pattern": "*.yaml",
	})
	if err != nil {
		t.Fatal(err)
	}
	if cmd != `find . -type f -name '*.yaml'` {
		t.Fatalf("unexpected command: %s", cmd)
	}
}

func TestRenderReportsMissingValues(t *testing.T) {
	_, err := Render(`sudo nmap -O {{host}}`, map[string]string{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRenderSupportsBuiltins(t *testing.T) {
	cmd, err := Render(`git -C {{aoo_app_dir}} status`, map[string]string{
		"aoo_app_dir": "/srv/aoo",
	})
	if err != nil {
		t.Fatal(err)
	}
	if cmd != `git -C '/srv/aoo' status` {
		t.Fatalf("unexpected command: %s", cmd)
	}
}
