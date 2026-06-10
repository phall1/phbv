package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/phall1/phbv/internal/bd"
)

// kanbanView renders one column per status category, with columns and the
// statuses within them discovered from the runtime schema.
func (m Model) kanbanView() string {
	cats := m.statusesByCategory()
	if len(cats) == 0 {
		return dimStyle.Render("no schema categories")
	}

	// Bucket issues by their status's category.
	statusCat := map[string]string{}
	for _, s := range m.schema.Statuses {
		cat := s.Category
		if cat == "" {
			cat = "other"
		}
		statusCat[s.Name] = cat
	}
	buckets := map[string][]bd.Issue{}
	for _, is := range m.issues {
		cat := statusCat[is.Status]
		if cat == "" {
			cat = "other"
		}
		buckets[cat] = append(buckets[cat], is)
	}

	colWidth := (m.width / len(cats)) - 1
	if colWidth < 12 {
		colWidth = 12
	}
	rows := m.height - 6

	cols := make([]string, 0, len(cats))
	for _, cat := range cats {
		cols = append(cols, m.renderColumn(cat, buckets[cat], colWidth, rows))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, cols...)
}

func (m Model) renderColumn(cat string, issues []bd.Issue, width, rows int) string {
	var b strings.Builder
	b.WriteString(columnTitle.Width(width).Render(fmt.Sprintf("%s (%d)", cat, len(issues))))
	b.WriteString("\n")
	shown := 0
	for _, is := range issues {
		if rows > 0 && shown >= rows {
			b.WriteString(dimStyle.Render(fmt.Sprintf("… %d more", len(issues)-shown)))
			break
		}
		icon := m.iconFor(is.Status)
		card := fmt.Sprintf("%s P%d %s", icon, is.Priority, truncate(is.Title, width-6))
		b.WriteString(lipgloss.NewStyle().Width(width).Render(card))
		b.WriteString("\n")
		shown++
	}
	return lipgloss.NewStyle().Width(width).MarginRight(1).Render(b.String())
}
