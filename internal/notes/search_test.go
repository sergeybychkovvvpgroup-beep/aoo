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

