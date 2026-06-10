package ui

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Project is a discovered beads workspace: a sibling directory containing a
// ".beads" subdir. Dir is the absolute path to that ".beads" dir, ready to
// hand to bd.New.
type Project struct {
	Name string
	Dir  string
}

// DiscoverProjects scans the immediate subdirectories of root for ones that
// contain a ".beads" directory, returning them sorted by Name. A root that
// does not exist (or any read error) yields an empty slice and a nil error —
// the picker degrades to "no projects" rather than surfacing an error.
func DiscoverProjects(root string) ([]Project, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, nil
	}
	var projects []Project
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		beads := filepath.Join(root, e.Name(), ".beads")
		info, err := os.Stat(beads)
		if err != nil || !info.IsDir() {
			continue
		}
		abs, err := filepath.Abs(beads)
		if err != nil {
			abs = beads
		}
		projects = append(projects, Project{Name: e.Name(), Dir: abs})
	}
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Name < projects[j].Name
	})
	return projects, nil
}

// projectPicker is the modal list of discovered projects. The cursor starts on
// the project matching currentDir (if present) so re-opening the picker lands
// on the active workspace.
type projectPicker struct {
	items  []Project
	cursor int
}

// newProjectPicker builds a picker over projects, positioning the cursor on the
// project whose Dir equals currentDir when one matches.
func newProjectPicker(projects []Project, currentDir string) projectPicker {
	cursor := 0
	for i, p := range projects {
		if p.Dir == currentDir {
			cursor = i
			break
		}
	}
	return projectPicker{items: projects, cursor: cursor}
}

// Update moves the cursor on j/k and arrow keys, clamping to the item range.
func (p projectPicker) Update(msg tea.Msg) (projectPicker, tea.Cmd) {
	if len(p.items) == 0 {
		return p, nil
	}
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "down", "j":
			if p.cursor < len(p.items)-1 {
				p.cursor++
			}
		case "up", "k":
			if p.cursor > 0 {
				p.cursor--
			}
		}
	}
	return p, nil
}

// View renders the project list, one row each, with the cursor row highlighted.
func (p projectPicker) View() string {
	if len(p.items) == 0 {
		return dimStyle.Render("no sibling beads workspaces found")
	}
	var b strings.Builder
	b.WriteString(headerStyle.Render("projects"))
	b.WriteString("\n\n")
	for i, item := range p.items {
		row := "  " + item.Name
		if i == p.cursor {
			row = "> " + item.Name
			b.WriteString(selectedRow.Render(row))
		} else {
			b.WriteString(row)
		}
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("    " + item.Dir))
		b.WriteString("\n")
	}
	return b.String()
}

// Selected returns the project under the cursor, or a zero Project when the
// list is empty.
func (p projectPicker) Selected() Project {
	if len(p.items) == 0 || p.cursor < 0 || p.cursor >= len(p.items) {
		return Project{}
	}
	return p.items[p.cursor]
}
