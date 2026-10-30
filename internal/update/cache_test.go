package update

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCachedNoticeIsReadOnly(t *testing.T) {
	file := filepath.Join(t.TempDir(), "update.json")
	data := []byte(`{"latest":"4.0.0","checked":"2020-01-01T00:00:00Z"}`)
	if err := os.WriteFile(file, data, 0o600); err != nil {
		t.Fatal(err)
	}
	n := Notifier{Current: "3.0.0", Cache: file}
	if n.CachedNewer() != "4.0.0" {
		t.Fatal("missing cached notice")
	}
	after, err := os.ReadFile(file)
	if err != nil || string(after) != string(data) {
		t.Fatal("cached notice changed its file")
	}
	n.Off = true
	if n.CachedNewer() != "" {
		t.Fatal("disabled notifier returned a notice")
	}
}
