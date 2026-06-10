package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/phall1/phbv/internal/bd"
	"github.com/phall1/phbv/internal/config"
)

// testSchema mirrors the beads built-in statuses so views render with real
// icons/categories without shelling out to bd.
func testSchema() bd.Schema {
	return bd.Schema{
		Version: bd.SchemaVersionTested,
		AppVer:  "test",
		Statuses: []bd.Status{
			{Name: "open", Category: "active", Icon: "○"},
			{Name: "in_progress", Category: "wip", Icon: "◐"},
			{Name: "blocked", Category: "wip", Icon: "●"},
			{Name: "closed", Category: "done", Icon: "✓"},
		},
		Types: []bd.IssueType{{Name: "feature"}},
	}
}

func testModel(t *testing.T) Model {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	km, err := config.LoadKeyMap()
	if err != nil {
		t.Fatalf("keymap: %v", err)
	}
	m := New(bd.New(""), testSchema(), km, "phbv", "", "")
	issues := []bd.Issue{
		{ID: "phbv-aaa", Title: "build adapter", Status: "closed", Priority: 0, Type: "feature"},
		{ID: "phbv-bbb", Title: "tui shell", Status: "in_progress", Priority: 0, Type: "feature"},
		{ID: "phbv-ccc", Title: "filter view", Status: "blocked", Priority: 2, Type: "feature",
			Parent: "phbv-bbb", DependencyCount: 1,
			Dependencies: []bd.Dependency{{IssueID: "phbv-ccc", DependsOn: "phbv-bbb", Type: "blocks"}}},
	}
	// Drive through Update so state mirrors real runtime.
	mm, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	mm, _ = mm.(Model).Update(issuesMsg{view: viewList, issues: issues})
	mm, _ = mm.(Model).Update(issuesMsg{view: viewReady, issues: issues[1:2]})
	return mm.(Model)
}

func TestListViewRenders(t *testing.T) {
	m := testModel(t)
	out := m.View()
	for _, want := range []string{"phbv", "phbv-aaa", "build adapter", "phbv-ccc", "[list]"} {
		if !strings.Contains(out, want) {
			t.Errorf("list view missing %q\n---\n%s", want, out)
		}
	}
}

func TestKanbanViewGroupsByCategory(t *testing.T) {
	m := testModel(t)
	// tab twice: list -> ready -> kanban
	mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	mm, _ = mm.(Model).Update(tea.KeyMsg{Type: tea.KeyTab})
	out := mm.(Model).View()
	for _, want := range []string{"[kanban]", "active", "wip", "done"} {
		if !strings.Contains(out, want) {
			t.Errorf("kanban view missing category %q\n---\n%s", want, out)
		}
	}
}

func TestDetailViewShowsDeps(t *testing.T) {
	m := testModel(t)
	// move cursor to the blocked issue (index 2) and open it.
	mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	mm, _ = mm.(Model).Update(tea.KeyMsg{Type: tea.KeyDown})
	mm, _ = mm.(Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	out := mm.(Model).View()
	for _, want := range []string{"[detail]", "phbv-ccc", "filter view", "Dependencies", "phbv-bbb"} {
		if !strings.Contains(out, want) {
			t.Errorf("detail view missing %q\n---\n%s", want, out)
		}
	}
}

func TestFilterActivationFiltersList(t *testing.T) {
	m := testModel(t)
	// activate filter, type "tui" — only "tui shell" should survive.
	mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	for _, r := range "tui" {
		mm, _ = mm.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	out := mm.(Model).View()
	if !strings.Contains(out, "tui shell") {
		t.Errorf("filtered list should contain matching issue\n---\n%s", out)
	}
	if strings.Contains(out, "build adapter") {
		t.Errorf("filtered list should drop non-matching issue\n---\n%s", out)
	}
}

func TestDepTreeViewRenders(t *testing.T) {
	m := testModel(t)
	// cursor to the blocked issue (idx 2), open dep tree with 't'.
	mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	mm, _ = mm.(Model).Update(tea.KeyMsg{Type: tea.KeyDown})
	mm, _ = mm.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	got := mm.(Model)
	if got.view != viewDepTree {
		t.Fatalf("expected viewDepTree, got %v", got.view)
	}
	out := got.View()
	// phbv-ccc depends on phbv-bbb — both must appear in the tree.
	for _, want := range []string{"[deptree]", "phbv-ccc", "phbv-bbb"} {
		if !strings.Contains(out, want) {
			t.Errorf("deptree view missing %q\n---\n%s", want, out)
		}
	}
}

func TestProjectsViewRenders(t *testing.T) {
	m := testModel(t)
	mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	got := mm.(Model)
	if got.view != viewProjects {
		t.Fatalf("expected viewProjects, got %v", got.view)
	}
	// View must render without panicking even when the scan root is empty.
	if out := got.View(); !strings.Contains(out, "[projects]") {
		t.Errorf("projects view missing header\n---\n%s", out)
	}
}

func TestPriorityClamp(t *testing.T) {
	if got := clampPriority(-1); got != 0 {
		t.Errorf("clampPriority(-1) = %d, want 0", got)
	}
	if got := clampPriority(5); got != 4 {
		t.Errorf("clampPriority(5) = %d, want 4", got)
	}
	if got := clampPriority(2); got != 2 {
		t.Errorf("clampPriority(2) = %d, want 2", got)
	}
}

func TestQuitKey(t *testing.T) {
	m := testModel(t)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("expected quit command on 'q'")
	}
	if msg := cmd(); msg == nil {
		t.Fatal("quit command produced nil msg")
	}
}
