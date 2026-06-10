package bd

import (
	"context"
	"testing"
)

// SetPriority must reject out-of-range priorities before ever shelling out, so
// this runs without a live bd binary. We use a Client whose Bin points at a
// command that would fail loudly if it were ever executed, proving the
// validation short-circuits.
func TestSetPriorityRejectsOutOfRange(t *testing.T) {
	c := &Client{Bin: "/nonexistent/bd-should-not-run"}

	for _, p := range []int{-1, 5, 100} {
		if err := c.SetPriority(context.Background(), "bd-1", p); err == nil {
			t.Errorf("SetPriority(p=%d): expected validation error, got nil", p)
		}
	}
}
