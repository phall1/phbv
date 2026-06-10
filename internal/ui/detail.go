package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
)

// detailView renders the selected issue: header fields, a glamour-rendered
// description, and its dependency edges.
func (m Model) detailView() string {
	is := m.selected
	if is == nil {
		return dimStyle.Render("no issue selected")
	}
	var b strings.Builder

	b.WriteString(headerStyle.Render(is.ID))
	b.WriteString("  ")
	b.WriteString(fmt.Sprintf("%s %s", m.iconFor(is.Status), is.Status))
	b.WriteString("  ")
	b.WriteString(priorityStyle.Render(fmt.Sprintf("P%d", is.Priority)))
	b.WriteString("  ")
	b.WriteString(dimStyle.Render(is.Type))
	b.WriteString("\n\n")

	b.WriteString(lipglossBold(is.Title))
	b.WriteString("\n")

	meta := []string{}
	if is.Owner != "" {
		meta = append(meta, "owner: "+shortOwner(is.Owner))
	}
	if is.Parent != "" {
		meta = append(meta, "parent: "+is.Parent)
	}
	meta = append(meta, fmt.Sprintf("deps: %d", is.DependencyCount))
	meta = append(meta, fmt.Sprintf("dependents: %d", is.DependentCount))
	meta = append(meta, fmt.Sprintf("comments: %d", is.CommentCount))
	b.WriteString(dimStyle.Render(strings.Join(meta, "  ·  ")))
	b.WriteString("\n\n")

	if is.Description != "" {
		w := m.width - 2
		if w < 20 {
			w = 20
		}
		r, err := glamour.NewTermRenderer(glamour.WithAutoStyle(), glamour.WithWordWrap(w))
		if err == nil {
			if out, rerr := r.Render(is.Description); rerr == nil {
				b.WriteString(out)
			} else {
				b.WriteString(is.Description)
			}
		} else {
			b.WriteString(is.Description)
		}
	}

	if len(is.Dependencies) > 0 {
		b.WriteString("\n")
		b.WriteString(lipglossBold("Dependencies"))
		b.WriteString("\n")
		for _, d := range is.Dependencies {
			b.WriteString(fmt.Sprintf("  %s → %s (%s)\n", d.IssueID, d.DependsOn, d.Type))
		}
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("esc back"))
	return b.String()
}

func lipglossBold(s string) string {
	return headerStyle.Render(s)
}

// shortOwner trims the noisy "12345+user@users.noreply.github.com" form down to
// just the handle for display.
func shortOwner(owner string) string {
	if i := strings.Index(owner, "+"); i >= 0 {
		owner = owner[i+1:]
	}
	if i := strings.Index(owner, "@"); i >= 0 {
		owner = owner[:i]
	}
	return owner
}
