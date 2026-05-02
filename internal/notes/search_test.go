package notes

import "testing"

func TestFilterMatchesMultiWordAcrossFields(t *testing.T) {
	entries := []Entry{
		{
			Desc:       "internal warehouse router vyos",
			SourceFile: "chashnikovo-wh2.yaml",
			Tags:       []string{"warehouse"},
		},
		{
			Desc:       "office notebook",
			SourceFile: "notes_home.yaml",
		},
	}

	results := Filter(entries, "chash router")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if got := results[0].Entry.SourceFile; got != "chashnikovo-wh2.yaml" {
		t.Fatalf("unexpected match: %s", got)
	}
}

func TestFilterDoesNotReturnLooseSubsequenceNoise(t *testing.T) {
	entries := []Entry{
		{
			Desc:       "nmap os detect host",
			Template:   "sudo nmap -O {{host}}",
			SourceFile: "command_templates.yaml",
		},
		{
			Desc:       "mount cifs share",
			Template:   "sudo mount -t cifs -o username={{username}} //{{server}}/{{share}} {{mountpoint}}",
			SourceFile: "command_templates.yaml",
		},
		{
			Desc:       "ssh to host",
			Template:   "ssh {{user}}@{{host}}",
			SourceFile: "command_templates.yaml",
		},
	}

	results := Filter(entries, "nmap")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if got := results[0].Entry.Desc; got != "nmap os detect host" {
		t.Fatalf("unexpected match: %s", got)
	}
}

func TestFilterMatchesNumericTokensWithLeadingZeros(t *testing.T) {
	entries := []Entry{
		{
			Desc:       "ssh pve-beria-03",
			Run:        "ssh root@10.111.0.3",
			SourceFile: "netrack.yaml",
		},
		{
			Desc:       "IPMI netrack-pve-beria-03",
			Note:       "https://10.111.200.9",
			SourceFile: "netrack.yaml",
		},
		{
			Desc:       "ssh pve-beria-16-external-beeline",
			Run:        "ssh root@10.115.0.16",
			SourceFile: "beria-himki.yaml",
		},
	}

	results := Filter(entries, "beria-3")
	if len(results) < 2 {
		t.Fatalf("expected at least 2 results, got %d", len(results))
	}

	if results[0].Entry.Desc != "ssh pve-beria-03" {
		t.Fatalf("unexpected first result: %s", results[0].Entry.Desc)
	}
	if results[1].Entry.Desc != "IPMI netrack-pve-beria-03" {
		t.Fatalf("unexpected second result: %s", results[1].Entry.Desc)
	}
}

func TestEntryTypeTreatsTemplateAsRun(t *testing.T) {
	entry := Entry{
		Desc:     "ssh to host",
		Template: "ssh {{user}}@{{host}}",
	}

	if got := entry.Type(); got != TypeRun {
		t.Fatalf("expected template entry type %s, got %s", TypeRun, got)
	}
}
