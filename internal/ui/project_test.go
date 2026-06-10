package ui

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDiscoverProjects builds a temp dir with two beads workspaces (plus a
// noise dir without .beads) and asserts both are found, sorted by name, with
// absolute Dir paths pointing at the .beads subdir.
func TestDiscoverProjects(t *testing.T) {
	root := t.TempDir()
	// Create out of order to prove sorting, and a noise dir to prove filtering.
	for _, name := range []string{"zebra", "alpha", "no-beads"} {
		dir := filepath.Join(root, name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"zebra", "alpha"} {
		if err := os.MkdirAll(filepath.Join(root, name, ".beads"), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	projects, err := DiscoverProjects(root)
	if err != nil {
		t.Fatalf("DiscoverProjects: %v", err)
	}
	if len(projects) != 2 {
		t.Fatalf("want 2 projects, got %d: %+v", len(projects), projects)
	}
	if projects[0].Name != "alpha" || projects[1].Name != "zebra" {
		t.Fatalf("want sorted [alpha zebra], got [%s %s]", projects[0].Name, projects[1].Name)
	}
	wantDir := filepath.Join(root, "alpha", ".beads")
	if !filepath.IsAbs(projects[0].Dir) {
		t.Errorf("Dir not absolute: %s", projects[0].Dir)
	}
	if projects[0].Dir != wantDir {
		// wantDir from t.TempDir() is already absolute, so this is a direct compare.
		t.Errorf("Dir = %s, want %s", projects[0].Dir, wantDir)
	}
}

// TestDiscoverProjectsMissingRoot ensures a nonexistent root degrades to an
// empty slice and a nil error rather than failing.
func TestDiscoverProjectsMissingRoot(t *testing.T) {
	projects, err := DiscoverProjects(filepath.Join(t.TempDir(), "does-not-exist"))
	if err != nil {
		t.Fatalf("want nil err for missing root, got %v", err)
	}
	if len(projects) != 0 {
		t.Fatalf("want empty slice, got %+v", projects)
	}
}
