package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaultsOnly(t *testing.T) {
	// No XDG override present.
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	km, err := LoadKeyMap()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got := km.Binding("quit").Keys(); len(got) == 0 || got[0] != "q" {
		t.Errorf("default quit binding = %v, want [q ...]", got)
	}
	if got := km.Binding("up").Keys(); len(got) != 2 {
		t.Errorf("default up binding = %v, want 2 keys", got)
	}
}

func TestUserOverrideMergesPerAction(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	cfgDir := filepath.Join(dir, "phbv")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Override ONLY quit; everything else must keep defaults.
	override := "[keys]\nquit = [\"x\"]\n"
	if err := os.WriteFile(filepath.Join(cfgDir, "keys.toml"), []byte(override), 0o644); err != nil {
		t.Fatal(err)
	}

	km, err := LoadKeyMap()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got := km.Binding("quit").Keys(); len(got) != 1 || got[0] != "x" {
		t.Errorf("overridden quit = %v, want [x]", got)
	}
	// Untouched action retains its default.
	if got := km.Binding("refresh").Keys(); len(got) == 0 || got[0] != "r" {
		t.Errorf("refresh should keep default, got %v", got)
	}
}

func TestMalformedUserConfigErrors(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	cfgDir := filepath.Join(dir, "phbv")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "keys.toml"), []byte("this = is = not = toml"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadKeyMap(); err == nil {
		t.Fatal("expected error on malformed user config, got nil")
	}
}
