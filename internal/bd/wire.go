package bd

import "time"

// This file holds the raw JSON shapes emitted by `bd ... --json`. They are
// private to this package and exist only to be decoded and mapped into the
// domain types in types.go. Nothing outside internal/bd should ever see these.
//
// All fields are optional by intent: a missing field decodes to its zero value
// and the mapping functions tolerate it. We deliberately do NOT use
// DisallowUnknownFields — extra fields from a newer bd are ignored, not errors.

type wireIssue struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	Status          string    `json:"status"`
	Priority        int       `json:"priority"`
	IssueType       string    `json:"issue_type"`
	Owner           string    `json:"owner"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	CreatedBy       string    `json:"created_by"`
	Parent          string    `json:"parent"`
	Dependencies    []wireDep `json:"dependencies"`
	DependencyCount int       `json:"dependency_count"`
	DependentCount  int       `json:"dependent_count"`
	CommentCount    int       `json:"comment_count"`
}

type wireDep struct {
	IssueID   string    `json:"issue_id"`
	DependsOn string    `json:"depends_on_id"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy string    `json:"created_by"`
}

type wireStatuses struct {
	BuiltInStatuses []wireStatus `json:"built_in_statuses"`
	CustomStatuses  []wireStatus `json:"custom_statuses"`
	SchemaVersion   int          `json:"schema_version"`
}

type wireStatus struct {
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type wireTypes struct {
	CoreTypes   []wireType `json:"core_types"`
	CustomTypes []wireType `json:"custom_types"`
}

type wireType struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type wireVersion struct {
	Version       string `json:"version"`
	Branch        string `json:"branch"`
	Build         string `json:"build"`
	SchemaVersion int    `json:"schema_version"`
}

func (w wireIssue) toDomain() Issue {
	deps := make([]Dependency, 0, len(w.Dependencies))
	for _, d := range w.Dependencies {
		deps = append(deps, Dependency{
			IssueID:   d.IssueID,
			DependsOn: d.DependsOn,
			Type:      d.Type,
			CreatedAt: d.CreatedAt,
			CreatedBy: d.CreatedBy,
		})
	}
	return Issue{
		ID:              w.ID,
		Title:           w.Title,
		Description:     w.Description,
		Status:          w.Status,
		Priority:        w.Priority,
		Type:            w.IssueType,
		Owner:           w.Owner,
		CreatedAt:       w.CreatedAt,
		UpdatedAt:       w.UpdatedAt,
		CreatedBy:       w.CreatedBy,
		Parent:          w.Parent,
		Dependencies:    deps,
		DependencyCount: w.DependencyCount,
		DependentCount:  w.DependentCount,
		CommentCount:    w.CommentCount,
	}
}
