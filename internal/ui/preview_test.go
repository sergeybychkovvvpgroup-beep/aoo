package ui

import (
	"strings"
	"testing"

	"aoo/internal/notes"
)

func TestPreviewExcerptCentersOnMatchedLine(t *testing.T) {
	snippet := notes.PreviewSnippet{
		Section: "show",
		Text: strings.Join([]string{
			"line one",
			"line two",
			"set firewall group address-group NATUM address 178.154.247.75",
			"line four",
			"line five",
		}, "\n"),
		Occurrence: notes.Occurrence{Start: 28, End: 33},
	}

	lines := previewExcerpt(snippet, 32, 4)
	if len(lines) == 0 {
		t.Fatal("expected excerpt lines")
	}

	found := false
	for _, line := range lines {
		if strings.Contains(line.text, "NATU") && len(line.hits) > 0 {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected excerpt to include the matched line, got %#v", lines)
	}
}

func TestPreviewExcerptKeepsHorizontalWindowAroundMatch(t *testing.T) {
	snippet := notes.PreviewSnippet{
		Section: "show",
		Text: strings.Join([]string{
			"set service dhcp-server shared-network-name LAN-VOIP-OLD subnet 192.168.11.0/24 static-mapping voip-new ip-address '192.168.11.10'",
			"set service dhcp-server shared-network-name LAN-VOIP-OLD subnet 192.168.11.0/24 static-mapping voip-new mac-address '06:A6:73:64:BB:70'",
		}, "\n"),
		Occurrence: notes.Occurrence{Start: 84, End: 98},
	}

	lines := previewExcerpt(snippet, 48, 4)
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 lines, got %d", len(lines))
	}
	if !strings.Contains(lines[0].text, "ip-address") {
		t.Fatalf("expected first line to show tail with ip-address, got %q", lines[0].text)
	}
	if !strings.Contains(lines[1].text, "mac-address") {
		t.Fatalf("expected second line to show next full related line, got %q", lines[1].text)
	}
}

func TestPreviewSingleLineShowsMatchedLineInsteadOfEllipsis(t *testing.T) {
	preview := notes.PreviewMatch{
		Section: "show",
		Snippets: []notes.PreviewSnippet{{
			Section: "show",
			Text: strings.Join([]string{
				"set protocols static route 0.0.0.0/0 next-hop 1.1.1.1",
				"set service dhcp-server subnet 10.120.0.0/16 static-mapping vyos-lan ip-address '10.120.0.5'",
			}, "\n"),
			Occurrence: notes.Occurrence{Start: 94, End: 108},
		}},
	}

	lines := previewPaneLines(preview, 60, 1, 0, Theme{})
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	if strings.TrimSpace(lines[0]) == "" || strings.Contains(lines[0], "вЂ¦") && strings.TrimSpace(lines[0]) == "вЂ¦" {
		t.Fatalf("expected matched content instead of plain ellipsis, got %q", lines[0])
	}
}
