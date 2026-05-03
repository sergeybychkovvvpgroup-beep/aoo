package ui

import (
	"reflect"
	"testing"

	"aoo/internal/notes"
)

func TestAvailableTagsStartsWithAllAndSortsUniqueTags(t *testing.T) {
	entries := []notes.Entry{
		{Tags: []string{"vyos", "ssh"}},
		{Tags: []string{"ssh", "proxmox"}},
		{Tags: []string{"docs"}},
	}

	got := availableTags(entries)
	want := []string{notes.TypeAll, "docs", "proxmox", "ssh", "vyos"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected tags: got %v want %v", got, want)
	}
}
