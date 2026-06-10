// Command phbv is a terminal viewer for beads (bd) issue databases. It talks to
// the bd CLI exclusively (never the underlying Dolt store), so it stays robust
// across beads storage changes. See internal/bd for the design contract.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/phall1/phbv/internal/bd"
	"github.com/phall1/phbv/internal/config"
	"github.com/phall1/phbv/internal/ui"
)

// gitSHA is set via -ldflags at build time (see Makefile); "dev" for `go run`.
var gitSHA = "dev"

func main() {
	dir := flag.String("dir", "", "path to a .beads directory (default: $BEADS_DIR or cwd resolution)")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println("phbv", gitSHA)
		return
	}

	beadsDir := *dir
	if beadsDir == "" {
		beadsDir = os.Getenv("BEADS_DIR")
	}

	keys, err := config.LoadKeyMap()
	if err != nil {
		fmt.Fprintln(os.Stderr, "phbv: "+err.Error())
		os.Exit(1)
	}

	client := bd.New(beadsDir)

	// Handshake up front so the UI has the schema (drift banner, status icons,
	// kanban columns) before the first frame.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	schema, err := client.Handshake(ctx)
	cancel()
	if err != nil {
		fmt.Fprintln(os.Stderr, "phbv: cannot reach bd: "+err.Error())
		fmt.Fprintln(os.Stderr, "  is bd installed and is --dir/BEADS_DIR a valid .beads workspace?")
		os.Exit(1)
	}

	prog := tea.NewProgram(
		ui.New(client, schema, keys, projectName(beadsDir), beadsDir, projectRoot(beadsDir)),
		tea.WithAltScreen(),
	)
	if _, err := prog.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "phbv: "+err.Error())
		os.Exit(1)
	}
}

// projectName derives a friendly label from the beads dir (the repo dir name),
// falling back to "cwd" when bd is resolving the workspace itself.
func projectName(beadsDir string) string {
	if beadsDir == "" {
		return "cwd"
	}
	clean := filepath.Clean(beadsDir)
	if filepath.Base(clean) == ".beads" {
		clean = filepath.Dir(clean)
	}
	return filepath.Base(clean)
}

// projectRoot derives the directory to scan for sibling beads workspaces: the
// parent of the resolved workspace dir. If beadsDir is ".beads" under some dir,
// the workspace is that dir and root is its parent; if beadsDir is the workspace
// itself, root is its parent. When unresolved (empty), fall back to the parent
// of the cwd so the picker still has somewhere to look.
func projectRoot(beadsDir string) string {
	workspace := beadsDir
	if workspace == "" {
		if cwd, err := os.Getwd(); err == nil {
			workspace = cwd
		} else {
			return ""
		}
	}
	clean := filepath.Clean(workspace)
	if filepath.Base(clean) == ".beads" {
		clean = filepath.Dir(clean)
	}
	return filepath.Dir(clean)
}
