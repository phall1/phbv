package bd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
)

// Client runs the bd binary as a subprocess. It is safe for concurrent use:
// each call spawns its own process and shares no mutable state.
type Client struct {
	// Bin is the bd executable (default "bd", resolved via PATH).
	Bin string
	// Dir, if set, scopes commands to a specific beads workspace by exporting
	// BEADS_DIR. Empty means bd resolves the workspace from the cwd as usual.
	Dir string
}

// New returns a Client for the given beads directory (may be "").
func New(dir string) *Client {
	return &Client{Bin: "bd", Dir: dir}
}

// run executes `bd <args...> --json` and returns stdout. bd writes advisory
// noise (e.g. the "beads.role not configured" warning) to stderr, so stdout is
// clean JSON; we capture stderr only to enrich error messages.
func (c *Client) run(ctx context.Context, args ...string) ([]byte, error) {
	bin := c.Bin
	if bin == "" {
		bin = "bd"
	}
	cmd := exec.CommandContext(ctx, bin, args...)
	if c.Dir != "" {
		cmd.Env = append(cmd.Environ(), "BEADS_DIR="+c.Dir)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("bd %v: %w: %s", args, err, stderr.String())
	}
	return stdout.Bytes(), nil
}

// List returns issues matching the given bd list flags (e.g. "--status",
// "open"). Pass no extra args for all issues.
func (c *Client) List(ctx context.Context, extra ...string) ([]Issue, error) {
	args := append([]string{"list", "--json"}, extra...)
	out, err := c.run(ctx, args...)
	if err != nil {
		return nil, err
	}
	return decodeIssues(out)
}

// Ready returns issues that are unblocked and available to work on — the
// signature beads concept.
func (c *Client) Ready(ctx context.Context, extra ...string) ([]Issue, error) {
	args := append([]string{"ready", "--json"}, extra...)
	out, err := c.run(ctx, args...)
	if err != nil {
		return nil, err
	}
	return decodeIssues(out)
}

// decodeIssues parses a JSON array of issues. Tolerant by design: unknown
// fields are ignored and a null/empty payload yields an empty slice.
func decodeIssues(data []byte) ([]Issue, error) {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil, nil
	}
	var wire []wireIssue
	if err := json.Unmarshal(data, &wire); err != nil {
		return nil, fmt.Errorf("decode issues: %w", err)
	}
	out := make([]Issue, 0, len(wire))
	for _, w := range wire {
		out = append(out, w.toDomain())
	}
	return out, nil
}
