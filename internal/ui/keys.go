package ui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/phall1/phbv/internal/config"
)

// key_matches reports whether the pressed key triggers the named action under
// the active keymap.
func key_matches(msg tea.KeyMsg, km config.KeyMap, action string) bool {
	return key.Matches(msg, km.Binding(action))
}
