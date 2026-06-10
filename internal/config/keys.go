// Package config loads phbv's keymap. The canonical default is checked into
// the repo at keys.default.toml and embedded into the binary, so a built phbv
// is self-contained. A user override at $XDG_CONFIG_HOME/phbv/keys.toml is
// merged over the defaults action-by-action — partial overrides work, you only
// specify what you change.
package config

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/charmbracelet/bubbles/key"
)

//go:embed keys.default.toml
var defaultKeysTOML []byte

// keyFile is the on-disk shape of a keys.toml.
type keyFile struct {
	Keys map[string][]string `toml:"keys"`
}

// KeyMap holds resolved bindings keyed by action name. Bindings carry their
// keys plus help text, so the footer hints come for free.
type KeyMap struct {
	bindings map[string]key.Binding
}

// help text per action, used to build key.Binding help; absent actions still
// bind, they just have no footer label.
var actionHelp = map[string][2]string{
	"quit":       {"q", "quit"},
	"next_view":  {"tab", "next view"},
	"prev_view":  {"⇧tab", "prev view"},
	"up":         {"k", "up"},
	"down":       {"j", "down"},
	"open":       {"enter", "open"},
	"back":       {"esc", "back"},
	"filter":     {"/", "filter"},
	"refresh":    {"r", "refresh"},
	"close":      {"x", "close"},
	"reopen":     {"X", "reopen"},
	"prio_up":    {"+", "prio up"},
	"prio_down":  {"-", "prio down"},
	"deptree":    {"t", "deptree"},
	"projects":   {"p", "projects"},
	"assign":     {"a", "assign"},
	"toggle_all": {"A", "all"},
}

// Binding returns the resolved binding for an action (zero Binding if unknown).
func (k KeyMap) Binding(action string) key.Binding {
	return k.bindings[action]
}

// LoadKeyMap parses the embedded defaults, merges any user override over them,
// and returns resolved bindings. A malformed user file is a hard error — better
// to tell the user their config is broken than to silently ignore it.
func LoadKeyMap() (KeyMap, error) {
	var def keyFile
	if err := toml.Unmarshal(defaultKeysTOML, &def); err != nil {
		return KeyMap{}, fmt.Errorf("embedded default keymap is corrupt (build bug): %w", err)
	}

	if path := userKeysPath(); path != "" {
		if data, err := os.ReadFile(path); err == nil {
			var user keyFile
			if err := toml.Unmarshal(data, &user); err != nil {
				return KeyMap{}, fmt.Errorf("parse %s: %w", path, err)
			}
			for action, keys := range user.Keys {
				def.Keys[action] = keys // merge over: user wins per action
			}
		} else if !os.IsNotExist(err) {
			return KeyMap{}, fmt.Errorf("read %s: %w", path, err)
		}
	}

	return build(def), nil
}

func build(kf keyFile) KeyMap {
	out := KeyMap{bindings: make(map[string]key.Binding, len(kf.Keys))}
	for action, keys := range kf.Keys {
		opts := []key.BindingOpt{key.WithKeys(keys...)}
		if h, ok := actionHelp[action]; ok {
			opts = append(opts, key.WithHelp(h[0], h[1]))
		}
		out.bindings[action] = key.NewBinding(opts...)
	}
	return out
}

// userKeysPath resolves $XDG_CONFIG_HOME/phbv/keys.toml, falling back to
// ~/.config/phbv/keys.toml. Returns "" if no home can be determined.
func userKeysPath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "phbv", "keys.toml")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "phbv", "keys.toml")
}
