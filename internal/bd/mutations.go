package bd

import (
	"context"
	"fmt"
)

// Mutations are write commands against the bd CLI. Unlike List/Ready, they
// produce no JSON we care about — we run the subcommand and surface its error.
// We never pass --json here: bd's mutation output is human-oriented and we only
// need success/failure. The model refreshes by re-fetching after a mutation.

// Close transitions an issue to closed (`bd close <id>`).
func (c *Client) Close(ctx context.Context, id string) error {
	_, err := c.run(ctx, "close", id)
	return err
}

// Reopen transitions a closed issue back to open (`bd reopen <id>`).
func (c *Client) Reopen(ctx context.Context, id string) error {
	_, err := c.run(ctx, "reopen", id)
	return err
}

// SetPriority sets an issue's priority (`bd priority <id> <p>`). p must be in
// the range 0..4 (0 = highest); anything else is rejected without shelling out.
func (c *Client) SetPriority(ctx context.Context, id string, p int) error {
	if p < 0 || p > 4 {
		return fmt.Errorf("priority %d out of range (must be 0..4)", p)
	}
	_, err := c.run(ctx, "priority", id, fmt.Sprintf("%d", p))
	return err
}

// Assign sets an issue's owner (`bd assign <id> <who>`). An empty who unassigns.
func (c *Client) Assign(ctx context.Context, id, who string) error {
	_, err := c.run(ctx, "assign", id, who)
	return err
}
