package ui

import (
	"strings"
	"testing"

	"github.com/phall1/phbv/internal/bd"
)

// TestRenderDepTreeTwoLevel checks that a 2-level tree renders both ids and that
// the child line is indented beneath its parent.
func TestRenderDepTreeTwoLevel(t *testing.T) {
	root := bd.DepNode{
		Issue: bd.Issue{ID: "phbv-1", Title: "root issue", Status: "open"},
		Children: []bd.DepNode{
			{Issue: bd.Issue{ID: "phbv-2", Title: "child issue", Status: "open"}},
		},
	}

	out := RenderDepTree(root, func(string) string { return "○" }, 80)

	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), out)
	}
	if !strings.Contains(lines[0], "phbv-1") {
		t.Errorf("parent line missing id: %q", lines[0])
	}
	if !strings.Contains(lines[1], "phbv-2") {
		t.Errorf("child line missing id: %q", lines[1])
	}
	// The child must be visually indented under the parent: the connector
	// (└─/├─) precedes the child's id, whereas the parent's id leads its line.
	parentIdx := strings.Index(lines[0], "phbv-1")
	childIdx := strings.Index(lines[1], "phbv-2")
	if childIdx <= parentIdx {
		t.Errorf("child not indented: parent id at %d, child id at %d (%q / %q)",
			parentIdx, childIdx, lines[0], lines[1])
	}
	if !strings.Contains(lines[1], "└─") && !strings.Contains(lines[1], "├─") {
		t.Errorf("child line lacks tree connector: %q", lines[1])
	}
}

// TestRenderDepTreeTruncatedAndUnknown checks the cycle marker and unknown-node
// rendering.
func TestRenderDepTreeTruncatedAndUnknown(t *testing.T) {
	root := bd.DepNode{
		Issue: bd.Issue{ID: "phbv-1", Title: "root", Status: "open"},
		Children: []bd.DepNode{
			{Issue: bd.Issue{ID: "phbv-1", Title: "root", Status: "open"}, Truncated: true},
			{Issue: bd.Issue{ID: "phbv-9"}}, // missing issue: empty Title
		},
	}

	out := RenderDepTree(root, func(string) string { return "○" }, 80)

	if !strings.Contains(out, "(…cycle)") {
		t.Errorf("expected cycle marker in output: %q", out)
	}
	if !strings.Contains(out, "phbv-9 (unknown)") {
		t.Errorf("expected unknown marker for missing issue: %q", out)
	}
}
