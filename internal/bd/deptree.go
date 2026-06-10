package bd

// DepNode is one node in a dependency tree built client-side from already-
// fetched issues. The tree models the "blocks" direction: a node's children are
// the issues it depends on (its outgoing DependsOn edges).
type DepNode struct {
	Issue    Issue
	Children []DepNode
	// Truncated is true if this node was cut to avoid a cycle (already visited).
	Truncated bool
}

// BuildDepTree returns the tree rooted at rootID, resolving DependsOn edges
// against the provided issues. Missing issues become a DepNode with a
// zero-value Issue whose ID is set. Cycles are broken via a visited set: the
// repeat node is marked Truncated and not recursed into.
func BuildDepTree(issues []Issue, rootID string) DepNode {
	index := make(map[string]Issue, len(issues))
	for _, iss := range issues {
		index[iss.ID] = iss
	}
	return buildNode(index, rootID, map[string]bool{})
}

// buildNode resolves a single issue ID into a DepNode, recursing through its
// DependsOn edges. visited tracks the IDs on the current path so cycles
// terminate; it is mutated on entry and restored on exit so siblings sharing a
// dependency are each expanded fully (only true cycles truncate).
func buildNode(index map[string]Issue, id string, visited map[string]bool) DepNode {
	iss, ok := index[id]
	if !ok {
		// Missing issue: surface the dangling edge with just its ID so the
		// view can render "<id> (unknown)".
		return DepNode{Issue: Issue{ID: id}}
	}
	if visited[id] {
		return DepNode{Issue: iss, Truncated: true}
	}

	visited[id] = true
	defer delete(visited, id)

	node := DepNode{Issue: iss}
	for _, dep := range iss.Dependencies {
		if dep.DependsOn == "" {
			continue
		}
		node.Children = append(node.Children, buildNode(index, dep.DependsOn, visited))
	}
	return node
}
