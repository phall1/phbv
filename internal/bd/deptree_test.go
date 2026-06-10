package bd

import "testing"

// depth returns the number of nodes on the longest root-to-leaf path.
func depth(n DepNode) int {
	max := 0
	for _, c := range n.Children {
		if d := depth(c); d > max {
			max = d
		}
	}
	return max + 1
}

// dep is a small helper to build a DependsOn edge.
func dep(from, on string) Dependency {
	return Dependency{IssueID: from, DependsOn: on, Type: "blocks"}
}

func TestBuildDepTreeLinearChain(t *testing.T) {
	// A -> B -> C
	issues := []Issue{
		{ID: "A", Title: "a", Dependencies: []Dependency{dep("A", "B")}},
		{ID: "B", Title: "b", Dependencies: []Dependency{dep("B", "C")}},
		{ID: "C", Title: "c"},
	}
	root := BuildDepTree(issues, "A")

	if root.Issue.ID != "A" {
		t.Fatalf("root ID = %q, want A", root.Issue.ID)
	}
	if got := depth(root); got != 3 {
		t.Fatalf("chain depth = %d, want 3", got)
	}
	if len(root.Children) != 1 || root.Children[0].Issue.ID != "B" {
		t.Fatalf("A child = %+v, want single B", root.Children)
	}
	b := root.Children[0]
	if len(b.Children) != 1 || b.Children[0].Issue.ID != "C" {
		t.Fatalf("B child = %+v, want single C", b.Children)
	}
}

func TestBuildDepTreeCycleTerminates(t *testing.T) {
	// A -> B -> A : must terminate, marking the repeat node Truncated.
	issues := []Issue{
		{ID: "A", Title: "a", Dependencies: []Dependency{dep("A", "B")}},
		{ID: "B", Title: "b", Dependencies: []Dependency{dep("B", "A")}},
	}
	root := BuildDepTree(issues, "A") // would hang if cycles weren't broken

	if len(root.Children) != 1 {
		t.Fatalf("A children = %d, want 1", len(root.Children))
	}
	b := root.Children[0]
	if b.Issue.ID != "B" {
		t.Fatalf("A child ID = %q, want B", b.Issue.ID)
	}
	if len(b.Children) != 1 {
		t.Fatalf("B children = %d, want 1 (the truncated A)", len(b.Children))
	}
	repeat := b.Children[0]
	if repeat.Issue.ID != "A" {
		t.Fatalf("B child ID = %q, want A", repeat.Issue.ID)
	}
	if !repeat.Truncated {
		t.Fatalf("repeated node A should be Truncated")
	}
	if len(repeat.Children) != 0 {
		t.Fatalf("truncated node must not recurse, got %d children", len(repeat.Children))
	}
}

func TestBuildDepTreeMissingIssue(t *testing.T) {
	// A depends on Z, which is not in the issue set.
	issues := []Issue{
		{ID: "A", Title: "a", Dependencies: []Dependency{dep("A", "Z")}},
	}
	root := BuildDepTree(issues, "A")
	if len(root.Children) != 1 {
		t.Fatalf("A children = %d, want 1", len(root.Children))
	}
	missing := root.Children[0]
	if missing.Issue.ID != "Z" {
		t.Fatalf("missing node ID = %q, want Z", missing.Issue.ID)
	}
	if missing.Issue.Title != "" {
		t.Fatalf("missing node Title = %q, want empty", missing.Issue.Title)
	}
}
