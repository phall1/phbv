package bd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

// SchemaVersionTested is the bd schema_version this build was developed and
// golden-tested against. Per the detect-and-warn policy, a mismatch produces a
// non-fatal banner (see (Schema).Drifted) — we never refuse to launch.
const SchemaVersionTested = 1

// Handshake interrogates the installed bd binary for everything the UI needs to
// render without hardcoding beads' vocabulary: valid statuses (with icons and
// categories for kanban columns), issue types, and the schema/app version.
//
// This is the runtime source-of-truth discovery. Call it once at startup.
func (c *Client) Handshake(ctx context.Context) (Schema, error) {
	var s Schema

	vout, err := c.run(ctx, "version", "--json")
	if err != nil {
		return s, fmt.Errorf("handshake version: %w", err)
	}
	var ver wireVersion
	if err := unmarshalTrim(vout, &ver); err != nil {
		return s, fmt.Errorf("handshake version decode: %w", err)
	}
	s.Version = ver.SchemaVersion
	s.AppVer = ver.Version

	sout, err := c.run(ctx, "statuses", "--json")
	if err != nil {
		return s, fmt.Errorf("handshake statuses: %w", err)
	}
	var st wireStatuses
	if err := unmarshalTrim(sout, &st); err != nil {
		return s, fmt.Errorf("handshake statuses decode: %w", err)
	}
	for _, w := range append(st.BuiltInStatuses, st.CustomStatuses...) {
		s.Statuses = append(s.Statuses, Status{
			Name:        w.Name,
			Category:    w.Category,
			Description: w.Description,
			Icon:        w.Icon,
		})
	}

	tout, err := c.run(ctx, "types", "--json")
	if err != nil {
		return s, fmt.Errorf("handshake types: %w", err)
	}
	var ts wireTypes
	if err := unmarshalTrim(tout, &ts); err != nil {
		return s, fmt.Errorf("handshake types decode: %w", err)
	}
	for _, w := range append(ts.CoreTypes, ts.CustomTypes...) {
		s.Types = append(s.Types, IssueType{Name: w.Name, Description: w.Description})
	}

	return s, nil
}

// Drifted reports whether the installed bd's schema_version differs from the
// version this build was tested against. The UI shows a non-fatal banner; it
// does not block.
func (s Schema) Drifted() bool {
	return s.Version != SchemaVersionTested
}

// HasStatus reports whether name is a status the installed bd knows about.
func (s Schema) HasStatus(name string) bool {
	for _, st := range s.Statuses {
		if st.Name == name {
			return true
		}
	}
	return false
}

func unmarshalTrim(data []byte, v any) error {
	return json.Unmarshal(bytes.TrimSpace(data), v)
}
