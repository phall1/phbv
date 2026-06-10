package bd

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// These golden tests pin the bd JSON boundary. The fixtures in testdata/ were
// captured from a live bd (schema_version 1). If `brew upgrade beads` reshapes
// the JSON, these go red and point straight at this package — converting a
// runtime "the TUI bugged out" into a CI signal.
//
// Regenerate fixtures with:
//   BEADS_DIR=/path/to/.beads bd statuses --json > testdata/statuses.json
//   ... (types, version, list) ...

func readGolden(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read golden %s: %v", name, err)
	}
	return b
}

func TestDecodeIssuesGolden(t *testing.T) {
	issues, err := decodeIssues(readGolden(t, "list.json"))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(issues) == 0 {
		t.Fatal("expected at least one issue from golden list.json")
	}
	first := issues[0]
	if first.ID == "" {
		t.Error("issue ID empty — id field mapping broke")
	}
	if first.Title == "" {
		t.Error("issue Title empty — title field mapping broke")
	}
	if first.Status == "" {
		t.Error("issue Status empty — status field mapping broke")
	}
	// The agent-browser fixture's first issue has a parent-child dependency.
	if first.Parent == "" {
		t.Error("expected non-empty Parent on golden first issue")
	}
}

func TestDecodeIssuesTolerant(t *testing.T) {
	// Empty and null payloads must yield no error and no issues.
	for _, in := range []string{"", "   ", "[]", "null"} {
		got, err := decodeIssues([]byte(in))
		if err != nil {
			t.Errorf("decodeIssues(%q): unexpected error %v", in, err)
		}
		if len(got) != 0 {
			t.Errorf("decodeIssues(%q): expected empty, got %d", in, len(got))
		}
	}
	// Unknown fields from a hypothetical newer bd must be ignored, not error.
	withExtra := `[{"id":"x-1","title":"t","status":"open","brand_new_field":42}]`
	got, err := decodeIssues([]byte(withExtra))
	if err != nil {
		t.Fatalf("unknown-field payload errored: %v", err)
	}
	if len(got) != 1 || got[0].ID != "x-1" {
		t.Fatalf("unknown-field payload mis-decoded: %+v", got)
	}
}

func TestHandshakeDecodeGolden(t *testing.T) {
	var st wireStatuses
	if err := unmarshalTrim(readGolden(t, "statuses.json"), &st); err != nil {
		t.Fatalf("statuses: %v", err)
	}
	if len(st.BuiltInStatuses) == 0 {
		t.Fatal("no built-in statuses decoded")
	}
	if st.SchemaVersion != SchemaVersionTested {
		t.Errorf("golden schema_version = %d, tested against %d", st.SchemaVersion, SchemaVersionTested)
	}
	// Icons drive kanban rendering; assert they survived decode.
	var sawIcon bool
	for _, s := range st.BuiltInStatuses {
		if s.Name == "" || s.Category == "" {
			t.Errorf("status missing name/category: %+v", s)
		}
		if s.Icon != "" {
			sawIcon = true
		}
	}
	if !sawIcon {
		t.Error("expected at least one status icon in golden data")
	}

	var ts wireTypes
	if err := unmarshalTrim(readGolden(t, "types.json"), &ts); err != nil {
		t.Fatalf("types: %v", err)
	}
	if len(ts.CoreTypes) == 0 {
		t.Fatal("no core types decoded")
	}

	var ver wireVersion
	if err := unmarshalTrim(readGolden(t, "version.json"), &ver); err != nil {
		t.Fatalf("version: %v", err)
	}
	if ver.SchemaVersion != SchemaVersionTested {
		t.Errorf("golden version schema_version = %d, want %d", ver.SchemaVersion, SchemaVersionTested)
	}
}

// TestLiveBdSmoke exercises the actually-installed bd. It skips cleanly when bd
// is absent (CI without beads) or BEADS_DIR is unset, so it never flakes — but
// when both are present it catches drift in the real binary before the UI does.
func TestLiveBdSmoke(t *testing.T) {
	if _, err := exec.LookPath("bd"); err != nil {
		t.Skip("bd not on PATH; skipping live smoke test")
	}
	dir := os.Getenv("PHBV_TEST_BEADS_DIR")
	if dir == "" {
		t.Skip("PHBV_TEST_BEADS_DIR unset; skipping live smoke test")
	}
	c := New(dir)
	ctx := context.Background()

	s, err := c.Handshake(ctx)
	if err != nil {
		t.Fatalf("live handshake: %v", err)
	}
	if len(s.Statuses) == 0 {
		t.Error("live handshake returned no statuses")
	}
	if s.Drifted() {
		t.Logf("NOTE: live bd schema_version %d differs from tested %d (non-fatal)", s.Version, SchemaVersionTested)
	}
	if _, err := c.List(ctx, "--limit", "5"); err != nil {
		t.Fatalf("live list: %v", err)
	}
}
