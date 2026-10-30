package cli

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/JittoJoseph/claude-switcher/v3/internal/update"
)

func TestShortCommandUsesCachedNoticeWithoutNetwork(t *testing.T) {
	f := newFixture(t)
	for _, key := range []string{"CLAUDE_SWITCHER_NO_UPDATE_CHECK", "DO_NOT_TRACK", "CI"} {
		t.Setenv(key, "")
	}
	f.put(filepath.Join(f.vault, "cache", "update.json"), `{"latest":"9.0.0","checked":"2020-01-01T00:00:00Z"}`)
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		http.Error(w, "not needed", http.StatusServiceUnavailable)
	}))
	defer server.Close()
	previous := update.Releases
	update.Releases = server.URL
	defer func() { update.Releases = previous }()
	var out, errOut bytes.Buffer
	a := newApp(Build{Version: "3.0.0"}, []string{"list", "--vault-dir", f.vault}, strings.NewReader(""), &out, &errOut)
	a.ttyErr = true
	cmd, _, err := a.root().Find([]string{"list"})
	if err != nil {
		t.Fatal(err)
	}
	cmd.SetContext(context.Background())
	if err := a.prepare(cmd, nil); err != nil {
		t.Fatal(err)
	}
	a.notifyUpdate()
	if requests.Load() != 0 {
		t.Fatal("a short command contacted the network")
	}
	if !strings.Contains(errOut.String(), "9.0.0") {
		t.Fatal("cached update notice was lost")
	}
}

func TestCLIAutoThemeUsesEnvironment(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	f := newFixture(t)
	a := newApp(Build{}, []string{"--vault-dir", f.vault}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	for _, c := range []struct {
		value string
		dark  bool
	}{{"15;0", true}, {"0;15", false}, {"0;7", false}, {"", true}} {
		t.Setenv("COLORFGBG", c.value)
		if got := a.darkBackground(); got != c.dark {
			t.Fatalf("%q: dark=%v", c.value, got)
		}
	}
}
