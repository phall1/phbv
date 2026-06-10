package ui

import "github.com/charmbracelet/lipgloss"

// Minimal palette. Kept in one place so the look can evolve without hunting
// through view code.
var (
	colSubtle   = lipgloss.Color("240")
	colAccent   = lipgloss.Color("69")
	colWarn     = lipgloss.Color("214")
	colSelBg    = lipgloss.Color("236")
	colPriority = lipgloss.Color("203")

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colAccent)

	driftStyle = lipgloss.NewStyle().
			Foreground(colWarn).
			Bold(true)

	countsStyle = lipgloss.NewStyle().
			Foreground(colSubtle)

	footerStyle = lipgloss.NewStyle().
			Foreground(colSubtle).
			BorderTop(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(colSubtle)

	selectedRow = lipgloss.NewStyle().
			Background(colSelBg).
			Bold(true)

	priorityStyle = lipgloss.NewStyle().Foreground(colPriority)

	dimStyle = lipgloss.NewStyle().Foreground(colSubtle)

	columnTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colAccent).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(colSubtle)
)
