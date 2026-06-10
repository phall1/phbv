package ui

import (
	"testing"

	"github.com/phall1/phbv/internal/bd"
)

func TestFuzzyFilter(t *testing.T) {
	issues := []bd.Issue{
		{ID: "bd-1", Title: "tui shell"},
		{ID: "bd-2", Title: "database migration"},
	}

	t.Run("empty query is identity", func(t *testing.T) {
		got := FuzzyFilter(issues, "")
		if len(got) != len(issues) {
			t.Fatalf("empty query: got %d issues, want %d", len(got), len(issues))
		}
		for i := range issues {
			if got[i].ID != issues[i].ID {
				t.Errorf("empty query reordered: got[%d]=%s want %s", i, got[i].ID, issues[i].ID)
			}
		}
	})

	t.Run("query surfaces match first", func(t *testing.T) {
		got := FuzzyFilter(issues, "tui")
		if len(got) == 0 {
			t.Fatal(`query "tui" matched nothing`)
		}
		if got[0].Title != "tui shell" {
			t.Errorf(`query "tui" ranked %q first, want "tui shell"`, got[0].Title)
		}
	})

	t.Run("case-insensitive", func(t *testing.T) {
		got := FuzzyFilter(issues, "TUI")
		if len(got) == 0 || got[0].Title != "tui shell" {
			t.Errorf(`query "TUI" did not surface "tui shell"`)
		}
	})
}
