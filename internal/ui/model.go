package ui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/phall1/phbv/internal/bd"
	"github.com/phall1/phbv/internal/config"
)

type view int

const (
	viewList view = iota
	viewReady
	viewKanban
	viewDetail
	viewDepTree
	viewProjects
)

// cycleViews is the number of leading views that tab/shift+tab rotates through
// (list -> ready -> kanban). viewDetail, viewDepTree, and viewProjects are modal
// and entered/exited by their own keys, not the tab cycle.
const cycleViews = 3

func (v view) title() string {
	switch v {
	case viewList:
		return "list"
	case viewReady:
		return "ready"
	case viewKanban:
		return "kanban"
	case viewDetail:
		return "detail"
	case viewDepTree:
		return "deptree"
	case viewProjects:
		return "projects"
	}
	return "?"
}

// fetchTimeout bounds a single bd invocation so a wedged subprocess can't hang
// the UI forever.
const fetchTimeout = 10 * time.Second

// Model is the root Bubble Tea model. It owns a shared issue cache and routes
// rendering to the active view.
type Model struct {
	client  *bd.Client
	schema  bd.Schema
	keys    config.KeyMap
	project string

	view     view
	prevView view // to return from detail/deptree/projects

	issues   []bd.Issue // cache for list/kanban
	ready    []bd.Issue // cache for ready view
	cursor   int
	selected *bd.Issue

	filter      filterModel
	tree        bd.DepNode
	projects    []Project
	picker      projectPicker
	projectRoot string // parent dir scanned for sibling .beads workspaces
	beadsDir    string // the active .beads dir being watched
	assignee    string // resolved git user.email / $USER, used for `a` assign
	showAll     bool   // when true, list includes closed issues (bd --all)

	width, height int
	loading       bool
	err           error
}

// New builds the root model. The schema handshake has already run (so we can
// render the drift banner and data-driven columns immediately). beadsDir is the
// resolved .beads directory (watched for live refresh); projectRoot is the
// parent directory scanned for sibling workspaces by the project picker.
func New(client *bd.Client, schema bd.Schema, keys config.KeyMap, project, beadsDir, projectRoot string) Model {
	return Model{
		client:      client,
		schema:      schema,
		keys:        keys,
		project:     project,
		view:        viewList,
		loading:     true,
		filter:      newFilter(),
		beadsDir:    beadsDir,
		projectRoot: projectRoot,
		assignee:    resolveAssignee(),
	}
}

// resolveAssignee determines who `a` (assign) targets: the git user.email if
// configured, falling back to $USER. Read once at startup so assign never
// shells out to git per keystroke.
func resolveAssignee() string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "git", "config", "user.email").Output()
	if err == nil {
		if email := strings.TrimSpace(string(out)); email != "" {
			return email
		}
	}
	return strings.TrimSpace(os.Getenv("USER"))
}

// --- messages ---

type issuesMsg struct {
	view   view
	issues []bd.Issue
}
type errMsg struct{ err error }

// mutationDoneMsg signals a mutation finished; carries any error so the model
// can surface it, then triggers a refetch.
type mutationDoneMsg struct{ err error }

// projectSwitchedMsg signals a project swap completed (handshake + initial
// fetch happened off the UI thread).
type projectSwitchedMsg struct {
	schema bd.Schema
	err    error
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.fetchList(), m.fetchReady(), WatchBeadsCmd(m.beadsDir))
}

func (m Model) fetchList() tea.Cmd {
	c := m.client
	// bd hides closed issues by default; --all includes them so the list view
	// isn't misleadingly empty once work is done.
	var extra []string
	if m.showAll {
		extra = []string{"--all"}
	}
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
		defer cancel()
		issues, err := c.List(ctx, extra...)
		if err != nil {
			return errMsg{err}
		}
		return issuesMsg{view: viewList, issues: issues}
	}
}

func (m Model) fetchReady() tea.Cmd {
	c := m.client
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
		defer cancel()
		issues, err := c.Ready(ctx)
		if err != nil {
			return errMsg{err}
		}
		return issuesMsg{view: viewReady, issues: issues}
	}
}

// refetch refreshes both caches. Used after a mutation or a file-change event.
func (m Model) refetch() tea.Cmd {
	return tea.Batch(m.fetchList(), m.fetchReady())
}

// rawList returns the unfiltered slice backing the current view.
func (m *Model) rawList() []bd.Issue {
	if m.view == viewReady {
		return m.ready
	}
	return m.issues
}

// activeList returns the slice backing the current view's cursor, applying the
// fuzzy filter when active in a list/ready view.
func (m *Model) activeList() []bd.Issue {
	list := m.rawList()
	if m.filter.Active() && (m.view == viewList || m.view == viewReady) {
		return FuzzyFilter(list, m.filter.Query())
	}
	return list
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case errMsg:
		m.err = msg.err
		m.loading = false
		return m, nil

	case issuesMsg:
		m.loading = false
		if msg.view == viewReady {
			m.ready = msg.issues
		} else {
			m.issues = msg.issues
		}
		if m.cursor >= len(m.activeList()) {
			m.cursor = 0
		}
		return m, nil

	case mutationDoneMsg:
		if msg.err != nil {
			m.err = msg.err
		}
		return m, m.refetch()

	case projectSwitchedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.schema = msg.schema
		m.err = nil
		m.issues = nil
		m.ready = nil
		m.cursor = 0
		m.view = viewList
		return m, tea.Batch(m.refetch(), WatchBeadsCmd(m.beadsDir))

	case FileChangedMsg:
		return m, tea.Batch(m.refetch(), WatchBeadsCmd(m.beadsDir))

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// While the filter is active, route typing to it. esc/back deactivates;
	// everything else feeds the textinput so the user can type a query.
	if m.filter.Active() {
		if key_matches(msg, m.keys, "back") {
			m.filter.Deactivate()
			m.cursor = 0
			return m, nil
		}
		var cmd tea.Cmd
		m.filter, cmd = m.filter.Update(msg)
		if m.cursor >= len(m.activeList()) {
			m.cursor = 0
		}
		return m, cmd
	}

	// Project picker owns navigation/selection while in viewProjects.
	if m.view == viewProjects {
		return m.handleProjectsKey(msg)
	}

	switch {
	case key_matches(msg, m.keys, "quit"):
		return m, tea.Quit

	case key_matches(msg, m.keys, "refresh"):
		m.loading = true
		return m, m.refetch()

	case key_matches(msg, m.keys, "toggle_all"):
		m.showAll = !m.showAll
		m.cursor = 0
		return m, m.fetchList()

	case key_matches(msg, m.keys, "filter"):
		if m.view == viewList || m.view == viewReady {
			m.filter.Activate()
			m.cursor = 0
		}
		return m, nil

	case key_matches(msg, m.keys, "next_view"):
		if isCyclable(m.view) {
			m.view = (m.view + 1) % cycleViews
			m.cursor = 0
		}
		return m, nil

	case key_matches(msg, m.keys, "prev_view"):
		if isCyclable(m.view) {
			m.view = (m.view + cycleViews - 1) % cycleViews
			m.cursor = 0
		}
		return m, nil

	case key_matches(msg, m.keys, "down"):
		if n := len(m.activeList()); n > 0 && isListLike(m.view) {
			m.cursor = (m.cursor + 1) % n
		}
		return m, nil

	case key_matches(msg, m.keys, "up"):
		if n := len(m.activeList()); n > 0 && isListLike(m.view) {
			m.cursor = (m.cursor - 1 + n) % n
		}
		return m, nil

	case key_matches(msg, m.keys, "open"):
		if isListLike(m.view) {
			list := m.activeList()
			if m.cursor < len(list) {
				sel := list[m.cursor]
				m.selected = &sel
				m.prevView = m.view
				m.view = viewDetail
			}
		}
		return m, nil

	case key_matches(msg, m.keys, "back"):
		if m.view == viewDetail || m.view == viewDepTree {
			m.view = m.prevView
		}
		return m, nil

	case key_matches(msg, m.keys, "deptree"):
		if id := m.selectedID(); id != "" {
			m.tree = bd.BuildDepTree(m.issues, id)
			m.prevView = m.view
			m.view = viewDepTree
		}
		return m, nil

	case key_matches(msg, m.keys, "projects"):
		m.projects, _ = DiscoverProjects(m.projectRoot)
		m.picker = newProjectPicker(m.projects, m.beadsDir)
		m.prevView = m.view
		m.view = viewProjects
		return m, nil

	case key_matches(msg, m.keys, "close"):
		if id := m.selectedID(); id != "" {
			return m, m.mutate(func(c *bd.Client, ctx context.Context) error {
				return c.Close(ctx, id)
			})
		}
		return m, nil

	case key_matches(msg, m.keys, "reopen"):
		if id := m.selectedID(); id != "" {
			return m, m.mutate(func(c *bd.Client, ctx context.Context) error {
				return c.Reopen(ctx, id)
			})
		}
		return m, nil

	case key_matches(msg, m.keys, "prio_up"):
		if is := m.selectedIssue(); is != nil {
			p := clampPriority(is.Priority - 1)
			id := is.ID
			return m, m.mutate(func(c *bd.Client, ctx context.Context) error {
				return c.SetPriority(ctx, id, p)
			})
		}
		return m, nil

	case key_matches(msg, m.keys, "prio_down"):
		if is := m.selectedIssue(); is != nil {
			p := clampPriority(is.Priority + 1)
			id := is.ID
			return m, m.mutate(func(c *bd.Client, ctx context.Context) error {
				return c.SetPriority(ctx, id, p)
			})
		}
		return m, nil

	case key_matches(msg, m.keys, "assign"):
		if id := m.selectedID(); id != "" {
			who := m.assignee
			return m, m.mutate(func(c *bd.Client, ctx context.Context) error {
				return c.Assign(ctx, id, who)
			})
		}
		return m, nil
	}
	return m, nil
}

// handleProjectsKey routes keys in the project picker modal: back exits,
// open/enter selects (swaps client, re-handshakes, refetches), arrows/jk move.
func (m Model) handleProjectsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key_matches(msg, m.keys, "back"):
		m.view = m.prevView
		return m, nil
	case key_matches(msg, m.keys, "quit"):
		return m, tea.Quit
	case key_matches(msg, m.keys, "open"):
		sel := m.picker.Selected()
		if sel.Dir == "" {
			m.view = m.prevView
			return m, nil
		}
		m.client = bd.New(sel.Dir)
		m.beadsDir = sel.Dir
		m.project = sel.Name
		m.loading = true
		return m, m.switchProject()
	default:
		var cmd tea.Cmd
		m.picker, cmd = m.picker.Update(msg)
		return m, cmd
	}
}

// switchProject runs the handshake for the freshly-swapped client off the UI
// thread, returning a projectSwitchedMsg.
func (m Model) switchProject() tea.Cmd {
	c := m.client
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
		defer cancel()
		schema, err := c.Handshake(ctx)
		return projectSwitchedMsg{schema: schema, err: err}
	}
}

// mutate runs a write against the active client off the UI thread and triggers
// a refetch via mutationDoneMsg.
func (m Model) mutate(fn func(c *bd.Client, ctx context.Context) error) tea.Cmd {
	c := m.client
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
		defer cancel()
		return mutationDoneMsg{err: fn(c, ctx)}
	}
}

// selectedIssue returns the issue the cursor (or detail view) currently targets.
func (m *Model) selectedIssue() *bd.Issue {
	if m.view == viewDetail || m.view == viewDepTree {
		return m.selected
	}
	list := m.activeList()
	if m.cursor >= 0 && m.cursor < len(list) {
		is := list[m.cursor]
		return &is
	}
	return nil
}

func (m *Model) selectedID() string {
	if is := m.selectedIssue(); is != nil {
		return is.ID
	}
	return ""
}

func clampPriority(p int) int {
	if p < 0 {
		return 0
	}
	if p > 4 {
		return 4
	}
	return p
}

// isCyclable reports whether tab/shift+tab should rotate from this view.
func isCyclable(v view) bool {
	return v == viewList || v == viewReady || v == viewKanban
}

// isListLike reports whether the cursor/open/mutation keys operate on a row list.
func isListLike(v view) bool {
	return v == viewList || v == viewReady
}

func (m Model) View() string {
	if m.width == 0 {
		return "loading…"
	}
	var b strings.Builder
	b.WriteString(m.header())
	b.WriteString("\n")
	b.WriteString(m.body())
	b.WriteString("\n")
	b.WriteString(m.footer())
	return b.String()
}

func (m Model) header() string {
	left := headerStyle.Render(fmt.Sprintf("phbv ─ %s", m.project))
	scope := ""
	if m.showAll {
		scope = " (all)"
	}
	mid := dimStyle.Render(fmt.Sprintf("  [%s]%s", m.view.title(), scope))
	right := ""
	if m.schema.Drifted() {
		right = "  " + driftStyle.Render(fmt.Sprintf("⚠ bd schema %d (tested %d)", m.schema.Version, bd.SchemaVersionTested))
	}
	head := left + mid + right
	return head + "\n" + countsStyle.Render(m.statusCounts())
}

// statusCounts renders per-status tallies using icons from the runtime schema,
// so new beads statuses appear here with no code change.
func (m Model) statusCounts() string {
	counts := map[string]int{}
	for _, is := range m.issues {
		counts[is.Status]++
	}
	var parts []string
	for _, s := range m.schema.Statuses {
		if n := counts[s.Name]; n > 0 {
			parts = append(parts, fmt.Sprintf("%s %s %d", s.Icon, s.Name, n))
		}
	}
	return strings.Join(parts, "   ")
}

func (m Model) body() string {
	if m.err != nil {
		return driftStyle.Render("error: " + m.err.Error())
	}
	if m.loading {
		return dimStyle.Render("loading…")
	}
	switch m.view {
	case viewDetail:
		return m.detailView()
	case viewKanban:
		return m.kanbanView()
	case viewDepTree:
		return m.depTreeView()
	case viewProjects:
		return m.picker.View()
	default:
		return m.listBody()
	}
}

// listBody renders the (optionally filtered) row list, prepending the filter
// prompt line when the filter is active.
func (m Model) listBody() string {
	if m.filter.Active() {
		return m.filter.View() + "\n" + m.listView(m.activeList())
	}
	return m.listView(m.activeList())
}

// depTreeView renders the dependency tree modal with a header line.
func (m Model) depTreeView() string {
	root := m.tree.Issue.ID
	header := headerStyle.Render("dependency tree: " + root)
	return header + "\n\n" + RenderDepTree(m.tree, m.iconFor, m.width)
}

func (m Model) footer() string {
	hints := []string{
		"tab views", "/ filter", "enter open",
		"x close", "X reopen", "+/- prio", "a assign",
		"t deptree", "p projects", "A all", "r refresh", "q quit",
	}
	return footerStyle.Width(m.width).Render(strings.Join(hints, " · "))
}

// iconFor returns the schema icon for a status, or a fallback dot.
func (m Model) iconFor(status string) string {
	for _, s := range m.schema.Statuses {
		if s.Name == status {
			if s.Icon != "" {
				return s.Icon
			}
			break
		}
	}
	return "·"
}

// statusesByCategory groups schema statuses into kanban columns in a stable
// order. Categories are discovered from the schema, not hardcoded.
func (m Model) statusesByCategory() []string {
	seen := map[string]bool{}
	var order []string
	for _, s := range m.schema.Statuses {
		cat := s.Category
		if cat == "" {
			cat = "other"
		}
		if !seen[cat] {
			seen[cat] = true
			order = append(order, cat)
		}
	}
	sort.SliceStable(order, func(i, j int) bool {
		return categoryRank(order[i]) < categoryRank(order[j])
	})
	return order
}

// categoryRank gives kanban columns a sensible left-to-right flow while still
// tolerating unknown categories (which sort to the end).
func categoryRank(cat string) int {
	switch cat {
	case "active":
		return 0
	case "wip":
		return 1
	case "done":
		return 2
	case "frozen":
		return 3
	default:
		return 99
	}
}
