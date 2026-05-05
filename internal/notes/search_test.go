package notes

import (
	"strings"
	"testing"
)

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

	results := Filter(entries, "chash router", SearchModeFlat)
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
			Actions:    []Action{{Desc: "run", Template: "sudo nmap -O {{host}}"}},
			SourceFile: "command_templates.yaml",
		},
		{
			Desc:       "mount cifs share",
			Actions:    []Action{{Desc: "run", Template: "sudo mount -t cifs -o username={{username}} //{{server}}/{{share}} {{mountpoint}}"}},
			SourceFile: "command_templates.yaml",
		},
		{
			Desc:       "ssh to host",
			Actions:    []Action{{Desc: "run", Template: "ssh {{user}}@{{host}}"}},
			SourceFile: "command_templates.yaml",
		},
	}

	results := Filter(entries, "nmap", SearchModeFlat)
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
			Actions:    []Action{{Desc: "ssh", Cmd: "ssh root@10.111.0.3"}},
			SourceFile: "netrack.yaml",
		},
		{
			Desc:       "IPMI netrack-pve-beria-03",
			Actions:    []Action{{Desc: "show", Text: "https://10.111.200.9"}},
			SourceFile: "netrack.yaml",
		},
		{
			Desc:       "ssh pve-beria-16-external-beeline",
			Actions:    []Action{{Desc: "ssh", Cmd: "ssh root@10.115.0.16"}},
			SourceFile: "beria-himki.yaml",
		},
	}

	results := Filter(entries, "beria-3", SearchModeFlat)
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

func TestEntryActionSummarizesTemplate(t *testing.T) {
	entry := Entry{
		Desc:    "ssh to host",
		Actions: []Action{{Desc: "run", Template: "ssh {{user}}@{{host}}"}},
	}

	if got := entry.Action(); got != "TEMPLATE" {
		t.Fatalf("expected template action summary, got %s", got)
	}
}

func TestFilterBuildsPreviewFromMatchedShowText(t *testing.T) {
	entries := []Entry{
		{
			Desc: "router config",
			Actions: []Action{{
				Desc: "show",
				Text: "alpha\nbeta\nset firewall group address-group NATUM address 178.154.247.75\ngamma",
			}},
		},
	}

	preview := BuildPreview(entries[0], "natum")
	if preview.Section != "show" {
		t.Fatalf("expected show preview, got %s", preview.Section)
	}
	if len(preview.Occurrences) == 0 {
		t.Fatal("expected preview occurrences")
	}
	if !strings.Contains(preview.Text, "NATUM") {
		t.Fatalf("expected preview text to preserve original content, got %q", preview.Text)
	}
}

func TestFilterPreviewHandlesUnicodeWithoutPanic(t *testing.T) {
	entries := []Entry{
		{
			Desc: "маршрутизатор",
			Actions: []Action{{
				Desc: "show",
				Text: "первая строка\nвторая строка\nмаршрутизатор vyos\nтретья строка",
			}},
		},
	}

	preview := BuildPreview(entries[0], "марш")
	if len(preview.Occurrences) == 0 {
		t.Fatal("expected unicode preview occurrences")
	}
}

func TestFilterPreviewPrefersCandidateMatchingMoreQueryTerms(t *testing.T) {
	entries := []Entry{
		{
			Desc: "static-mapping",
			Actions: []Action{{
				Desc: "show",
				Text: "set protocols static route 1.1.1.1/32 next-hop 10.0.0.1\nset protocols static route 2.2.2.2/32 next-hop 10.0.0.2",
			}},
		},
	}

	preview := BuildPreview(entries[0], "static-mappi")
	if preview.Section != "title" {
		t.Fatalf("expected title preview, got %s", preview.Section)
	}
}

func TestFilterPreviewBuildsMultipleSnippets(t *testing.T) {
	entries := []Entry{
		{
			Desc: "router routes",
			Actions: []Action{{
				Desc: "show",
				Text: "set protocols static route 1.1.1.1/32\nset protocols static route 2.2.2.2/32",
			}},
		},
	}

	preview := BuildPreview(entries[0], "static")
	if len(preview.Snippets) < 2 {
		t.Fatalf("expected multiple preview snippets, got %d", len(preview.Snippets))
	}
}

func TestFilterPreviewDoesNotCapSnippetCountAtTwelve(t *testing.T) {
	entry := Entry{
		Desc: "router dhcp",
		Actions: []Action{{
			Desc: "show",
			Text: strings.Join([]string{
				"static-map-01",
				"static-map-02",
				"static-map-03",
				"static-map-04",
				"static-map-05",
				"static-map-06",
				"static-map-07",
				"static-map-08",
				"static-map-09",
				"static-map-10",
				"static-map-11",
				"static-map-12",
				"static-map-13",
				"static-map-14",
			}, "\n"),
		}},
	}

	preview := BuildPreview(entry, "static-map")
	if len(preview.Snippets) != 14 {
		t.Fatalf("expected 14 preview snippets, got %d", len(preview.Snippets))
	}
}

func TestFilterPreviewPrioritizesLinesMatchingAllTerms(t *testing.T) {
	entry := Entry{
		Desc: "router cha",
		Actions: []Action{{
			Desc: "show",
			Text: "set protocols static route 0.0.0.0/0 next-hop 1.1.1.1\n" +
				"set protocols static route 8.8.8.0/24 next-hop 1.1.1.1\n" +
				"set service dhcp-server subnet 10.120.0.0/16 static-mapping vyos-lan ip-address '10.120.0.5'\n",
		}},
	}

	preview := BuildPreview(entry, "static-mapping")
	if len(preview.Snippets) == 0 {
		t.Fatal("expected preview snippets")
	}
	snippet := preview.Snippets[0]
	runes := []rune(snippet.Text)
	start := snippet.Occurrence.Start - 10
	if start < 0 {
		start = 0
	}
	end := snippet.Occurrence.End + 20
	if end > len(runes) {
		end = len(runes)
	}
	got := string(runes[start:end])
	if !strings.Contains(got, "static-mapping") {
		t.Fatalf("expected first snippet to contain static-mapping, got %q", got)
	}
}

func TestFilterMatchesDottedNumericFragmentAsSingleTerm(t *testing.T) {
	entries := []Entry{
		{
			Desc: "vorsino pve proxmox",
			Actions: []Action{{
				Desc: "jump",
				Text: "proxmox 192.168.53.53\nxray gateway 192.168.53.33",
			}},
			SourceFile: "vorsino-kim-ips.yaml",
		},
		{
			Desc: "router cha",
			Actions: []Action{{
				Desc: "show",
				Text: "set protocols static route 10.101.53.0/24 next-hop 10.115.250.201\nset protocols static route 10.101.13.2/32 next-hop 10.115.250.201",
			}},
			SourceFile: "router-b.yaml",
		},
	}

	results := Filter(entries, "53.3", SearchModeFlat)
	if len(results) == 0 {
		t.Fatal("expected matches")
	}
	if got := results[0].Entry.SourceFile; got != "vorsino-kim-ips.yaml" {
		t.Fatalf("expected vorsino-kim-ips.yaml first, got %s", got)
	}
}

func TestFilterEntryFirstKeepsEntriesSeparate(t *testing.T) {
	entries := []Entry{
		{
			Desc:       "ssh vyos-volga",
			Actions:    []Action{{Desc: "ssh", Cmd: "ssh vyos@volga"}},
			SourceFile: "vyos-chashnikovo.yaml",
			SourcePath: "/notes/vyos-chashnikovo.yaml",
			SourceLine: 3,
			index:      1,
		},
		{
			Desc:       "show dhcp subnets",
			Actions:    []Action{{Desc: "show", Cmd: "show dhcp subnets"}},
			SourceFile: "vyos-chashnikovo.yaml",
			SourcePath: "/notes/vyos-chashnikovo.yaml",
			SourceLine: 8,
			index:      2,
		},
	}
	for i := range entries {
		entries[i].searchData = weightedFields(entries[i])
	}

	results := Filter(entries, "chashnikovo", SearchModeEntryFirst)
	if len(results) != 2 {
		t.Fatalf("expected 2 entry results, got %d", len(results))
	}
	if results[0].Entry.IsGroup() || results[1].Entry.IsGroup() {
		t.Fatal("expected plain entry results, not grouped entries")
	}
}

func TestFilterHybridUsesEntryResultsWithoutPrefix(t *testing.T) {
	entries := []Entry{
		{
			Desc:       "ssh vyos-volga",
			Actions:    []Action{{Desc: "ssh", Cmd: "ssh vyos@volga"}},
			SourceFile: "vyos-chashnikovo.yaml",
			SourcePath: "/notes/vyos-chashnikovo.yaml",
			index:      1,
		},
		{
			Desc:       "show dhcp subnets",
			Actions:    []Action{{Desc: "show", Cmd: "show dhcp subnets"}},
			SourceFile: "vyos-chashnikovo.yaml",
			SourcePath: "/notes/vyos-chashnikovo.yaml",
			index:      2,
		},
	}
	for i := range entries {
		entries[i].searchData = weightedFields(entries[i])
	}

	results := Filter(entries, "chashnikovo", SearchModeHybrid)
	if len(results) != 2 {
		t.Fatalf("expected 2 entry results, got %d", len(results))
	}
	if results[0].Entry.IsGroup() || results[1].Entry.IsGroup() {
		t.Fatal("expected hybrid search without prefix to return entries, not groups")
	}
}

func TestFilterHybridUsesFlatResultsWithPrefix(t *testing.T) {
	entries := []Entry{
		{
			Desc:       "ssh vyos-volga",
			Actions:    []Action{{Desc: "ssh", Cmd: "ssh vyos@volga"}},
			SourceFile: "vyos-chashnikovo.yaml",
			SourcePath: "/notes/vyos-chashnikovo.yaml",
			index:      1,
		},
		{
			Desc:       "show dhcp subnets",
			Actions:    []Action{{Desc: "show", Cmd: "show dhcp subnets"}},
			SourceFile: "vyos-chashnikovo.yaml",
			SourcePath: "/notes/vyos-chashnikovo.yaml",
			index:      2,
		},
	}
	for i := range entries {
		entries[i].searchData = weightedFields(entries[i])
	}

	results := Filter(entries, ":ssh", SearchModeHybrid)
	if len(results) != 1 {
		t.Fatalf("expected 1 flat result, got %d", len(results))
	}
	if results[0].Entry.IsGroup() {
		t.Fatal("expected prefix search to return flat entry, not group")
	}
	if got := results[0].Entry.Desc; got != "ssh vyos-volga" {
		t.Fatalf("unexpected first flat result: %s", got)
	}
}

func TestFilterCommandOnlyExcludesShowOnlyEntries(t *testing.T) {
	entries := []Entry{
		{
			Desc:       "ssh vyos-volga",
			Actions:    []Action{{Desc: "ssh", Cmd: "ssh vyos@volga"}},
			SourceFile: "vyos-chashnikovo.yaml",
			SourcePath: "/notes/vyos-chashnikovo.yaml",
			index:      1,
		},
		{
			Desc:       "router notes",
			Actions:    []Action{{Desc: "show", Text: "dhcp relay and static routes"}},
			SourceFile: "vyos-chashnikovo.yaml",
			SourcePath: "/notes/vyos-chashnikovo.yaml",
			index:      2,
		},
	}
	for i := range entries {
		entries[i].searchData = weightedFields(entries[i])
	}

	results := Filter(entries, ":vyos", SearchModeHybrid)
	if len(results) != 1 {
		t.Fatalf("expected only command entries in prefix search, got %d", len(results))
	}
	if got := results[0].Entry.Desc; got != "ssh vyos-volga" {
		t.Fatalf("unexpected command-only result: %s", got)
	}
}

func TestFilterCommandOnlyShowsCommandAsPrimaryLine(t *testing.T) {
	entries := []Entry{
		{
			Desc:       "ssh office-notebook",
			Actions:    []Action{{Desc: "ssh", Cmd: "ssh -t sergey\\ bychkov@192.168.41.49 wsl"}},
			SourceFile: "office-notebook.yaml",
			SourcePath: "/notes/office-notebook.yaml",
			index:      1,
		},
	}
	for i := range entries {
		entries[i].searchData = weightedFields(entries[i])
	}

	results := Filter(entries, ":ssh", SearchModeHybrid)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !strings.Contains(results[0].Label, "ssh -t sergey\\ bychkov@192.168.41.49 wsl") {
		t.Fatalf("expected command as primary label, got %q", results[0].Label)
	}
	if results[0].Detail != "ssh" {
		t.Fatalf("expected action desc as secondary line, got %q", results[0].Detail)
	}
}

func TestFilterFindsLegacyTopLevelDescViaActionDescCompatibility(t *testing.T) {
	result := LoadBytes("snippets.yaml", []byte(`
- desc: netplan example
  text: |
    network:
      version: 2
      ethernets:
        ens18:
          addresses:
            - 192.168.42.88/22
`))
	if len(result.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", result.Errors)
	}

	results := Filter(result.Entries, "netpla", SearchModeHybrid)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if got := results[0].Entry.SourceFile; got != "snippets.yaml" {
		t.Fatalf("unexpected result source file: %s", got)
	}
}
