package ui

import (
	"strings"

	"github.com/phall1/phbv/internal/bd"
)

// RenderDepTree renders a dependency tree as an indented ASCII tree. It is pure
// (string in/out, no tea): the model builds the tree with bd.BuildDepTree and
// passes its schema icon lookup so glyphs stay consistent with the other views.
//
// Each line is `<icon> <id>  <title>`. Children are indented under their parent
// with ├─/└─ connectors. Cycle-truncated nodes get a trailing "(…cycle)" and
// missing-issue nodes (empty Title) render as "<id> (unknown)".
func RenderDepTree(root bd.DepNode, iconFor func(status string) string, width int) string {
	var b strings.Builder
	renderNode(&b, root, iconFor, "", true, true)
	return b.String()
}

// renderNode appends one node line plus its subtree. prefix is the accumulated
// indentation for descendants; isRoot suppresses the connector on the top node;
// isLast picks └─ vs ├─ and extends the prefix accordingly.
func renderNode(b *strings.Builder, n bd.DepNode, iconFor func(status string) string, prefix string, isLast, isRoot bool) {
	b.WriteString(prefix)

	childPrefix := prefix
	if !isRoot {
		if isLast {
			b.WriteString("└─ ")
			childPrefix += "   "
		} else {
			b.WriteString("├─ ")
			childPrefix += "│  "
		}
	}

	b.WriteString(nodeLabel(n, iconFor))
	b.WriteString("\n")

	for i, c := range n.Children {
		renderNode(b, c, iconFor, childPrefix, i == len(n.Children)-1, false)
	}
}

// nodeLabel formats a single node's text: missing issues show "(unknown)",
// truncated (cycle) nodes get a trailing marker, everything else shows the
// status icon, id, and title.
func nodeLabel(n bd.DepNode, iconFor func(status string) string) string {
	if n.Issue.Title == "" {
		label := n.Issue.ID + " (unknown)"
		if n.Truncated {
			label += " (…cycle)"
		}
		return label
	}

	icon := ""
	if iconFor != nil {
		icon = iconFor(n.Issue.Status)
	}
	var sb strings.Builder
	if icon != "" {
		sb.WriteString(icon)
		sb.WriteString(" ")
	}
	sb.WriteString(n.Issue.ID)
	sb.WriteString("  ")
	sb.WriteString(n.Issue.Title)
	if n.Truncated {
		sb.WriteString(" (…cycle)")
	}
	return sb.String()
}
