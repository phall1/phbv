package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sahilm/fuzzy"

	"github.com/phall1/phbv/internal/bd"
)

// filterModel is the incremental fuzzy-filter overlay. It wraps a
// bubbles/textinput and an active flag; when active, the model routes key
// presses to it and uses Query() to narrow the visible list. The actual
// ranking lives in FuzzyFilter so it is testable without a tea program.
type filterModel struct {
	input  textinput.Model
	active bool
}

// newFilter builds an inactive filter with a "/"-style prompt. The textinput
// is not focused until Activate is called.
func newFilter() filterModel {
	ti := textinput.New()
	ti.Prompt = "/"
	ti.Placeholder = "filter…"
	return filterModel{input: ti}
}

// Update routes the message to the textinput (typing, backspace, etc.) only
// while the filter is active; an inactive filter swallows nothing.
func (f filterModel) Update(msg tea.Msg) (filterModel, tea.Cmd) {
	if !f.active {
		return f, nil
	}
	var cmd tea.Cmd
	f.input, cmd = f.input.Update(msg)
	return f, cmd
}

// View renders the prompt line (e.g. "/tui") when active, or "" otherwise so
// callers can unconditionally prepend it.
func (f filterModel) View() string {
	if !f.active {
		return ""
	}
	return f.input.View()
}

// Active reports whether the filter is capturing input.
func (f filterModel) Active() bool { return f.active }

// Query is the current filter text.
func (f filterModel) Query() string { return f.input.Value() }

// Activate focuses the textinput and starts capturing input.
func (f *filterModel) Activate() {
	f.active = true
	f.input.Focus()
}

// Deactivate blurs the textinput, stops capturing, and clears the query so the
// next activation starts fresh.
func (f *filterModel) Deactivate() {
	f.active = false
	f.input.Blur()
	f.input.SetValue("")
}

// FuzzyFilter ranks issues against query by fuzzy-matching the string
// "<id> <title>", best matches first and dropping non-matches. An empty (or
// whitespace-only) query returns issues unchanged. Matching is
// case-insensitive: both the query and the haystacks are folded to lower case
// before ranking (the v0.1.2 fuzzy package has no Fold variant).
func FuzzyFilter(issues []bd.Issue, query string) []bd.Issue {
	if strings.TrimSpace(query) == "" {
		return issues
	}
	haystacks := make([]string, len(issues))
	for i, is := range issues {
		haystacks[i] = strings.ToLower(is.ID + " " + is.Title)
	}
	matches := fuzzy.Find(strings.ToLower(query), haystacks)
	out := make([]bd.Issue, 0, len(matches))
	for _, m := range matches {
		out = append(out, issues[m.Index])
	}
	return out
}
