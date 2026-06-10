// Package bd is the anti-corruption layer between the phbv TUI and the `bd`
// (beads) CLI. It is the ONLY package that knows bd's JSON shape exists.
//
// Design contract (see also client.go):
//   - bd's CLI --json output is treated as the source of truth. We never touch
//     the underlying Dolt database, because storage churns and the CLI is the
//     only interface beads keeps backward-compatible.
//   - Statuses, types, and priorities are discovered at runtime via the
//     handshake (schema.go), never hardcoded as Go enums. New beads statuses
//     show up in the UI without a code change.
//   - Decoding is tolerant: unknown fields are ignored, missing fields take
//     zero values. Shape drift degrades to a blank cell, never a panic.
//   - The boundary is pinned by golden tests (client_test.go). When beads
//     reshapes its JSON, a test goes red pointing at this package — instead of
//     the TUI silently bugging out, the failure mode that killed prior viewers.
package bd

import "time"

// Issue is the viewer's domain model for a single bead. It is intentionally a
// subset of bd's JSON: we map only what the UI renders, so new bd fields are
// inert until we choose to surface them.
type Issue struct {
	ID          string
	Title       string
	Description string
	Status      string // open string, validated against Schema.Statuses
	Priority    int    // 0 = highest
	Type        string // open string, validated against Schema.Types
	Owner       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CreatedBy   string

	// Parent is the parent bead ID for epic/subtask hierarchy ("" if none).
	Parent string
	// Dependencies are this issue's outgoing edges (what it depends on).
	Dependencies []Dependency

	DependencyCount int
	DependentCount  int
	CommentCount    int
}

// Dependency is one edge in the dependency graph.
type Dependency struct {
	IssueID   string
	DependsOn string
	Type      string // blocks, related, parent-child, discovered-from, ...
	CreatedAt time.Time
	CreatedBy string
}

// Status describes a valid issue status, discovered at runtime. Category groups
// statuses for kanban columns and sorting (active, wip, done, frozen, ...).
type Status struct {
	Name        string
	Category    string
	Description string
	Icon        string
}

// IssueType is a valid issue type, discovered at runtime.
type IssueType struct {
	Name        string
	Description string
}

// Schema is the runtime handshake result: everything the UI needs to render
// without baking beads' vocabulary into Go enums.
type Schema struct {
	Version  int // bd's schema_version; we warn (never block) when it drifts.
	AppVer   string
	Statuses []Status
	Types    []IssueType
}
