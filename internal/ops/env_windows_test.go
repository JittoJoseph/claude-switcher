package ops

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/JittoJoseph/claude-switcher/v3/internal/platform"
)

func TestKnownDataOverrideKeepsDesktopController(t *testing.T) {
	local := t.TempDir()
	t.Setenv("LOCALAPPDATA", local)
	t.Setenv("APPDATA", t.TempDir())
	data := filepath.Join(local, "Claude-Data", "claude_pzs8sxrjxfjjc")
	must(t, os.MkdirAll(data, 0o700))
	env, err := Open(Options{DataDir: data, VaultDir: t.TempDir(), ProjectsDir: t.TempDir()})
	must(t, err)
	if env.Install.Kind != platform.MSIX || env.Install.AppID != "Claude_pzs8sxrjxfjjc!Claude" {
		t.Fatalf("known override lost its process checks or launch identity: %+v", env.Install)
	}
}
