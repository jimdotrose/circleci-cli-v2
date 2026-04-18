package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

// AssertGolden compares got against testdata/golden/<name> relative to the
// test's working directory (always the package directory for Go tests).
//
// Regenerate all golden files with:
//
//	UPDATE_GOLDEN=1 go test ./...
func AssertGolden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", "golden", name)

	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("golden: mkdir: %v", err)
		}
		if err := os.WriteFile(path, []byte(got), 0644); err != nil {
			t.Fatalf("golden: write %s: %v", path, err)
		}
		t.Logf("golden: updated %s", path)
		return
	}

	want, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		t.Fatalf("golden file %s does not exist — run UPDATE_GOLDEN=1 go test ./... to create it", path)
	}
	if err != nil {
		t.Fatalf("golden: read %s: %v", path, err)
	}

	if got != string(want) {
		t.Errorf("golden mismatch for %s\n--- want ---\n%s\n--- got ---\n%s", name, string(want), got)
	}
}
