package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/phall1/phbv/internal/bd"
)

// listView renders a flat list of issues, one row each, with the cursor row
// highlighted. Used by both the list and ready views.
func (m Model) listView(issues []bd.Issue) string {
	if len(issues) == 0 {
		return dimStyle.Render("no issues")
	}
	var b strings.Builder
	// Leave room for header (3 lines) + footer (2 lines).
	maxRows := m.height - 6
	if maxRows < 1 {
		maxRows = len(issues)
	}
	start := 0
	if m.cursor >= maxRows {
		start = m.cursor - maxRows + 1
	}
	end := start + maxRows
	if end > len(issues) {
		end = len(issues)
	}
	for i := start; i < end; i++ {
		b.WriteString(m.renderRow(issues[i], i == m.cursor))
		b.WriteString("\n")
	}
	if end < len(issues) {
		b.WriteString(dimStyle.Render(fmt.Sprintf("  … %d more", len(issues)-end)))
	}
	return b.String()
}

func (m Model) renderRow(is bd.Issue, selected bool) string {
	icon := m.iconFor(is.Status)
	pri := priorityStyle.Render(fmt.Sprintf("P%d", is.Priority))
	id := dimStyle.Render(fmt.Sprintf("%-22s", truncate(is.ID, 22)))
	title := truncate(is.Title, max(10, m.width-34))
	row := fmt.Sprintf(" %s %s %s %s", icon, id, pri, title)
	if selected {
		return selectedRow.Width(m.width).Render(row)
	}
	return row
}

func truncate(s string, n int) string {
	if n <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= n {
		return s
	}
	if n <= 1 {
		return "…"
	}
	return s[:n-1] + "…"
}
