package ops

import (
	"github.com/JittoJoseph/claude-switcher/v3/internal/platform"
	"os"
	"path/filepath"
	"testing"

	"github.com/JittoJoseph/claude-switcher/v3/internal/i18n"
)

func TestDoctorFindsMissingCoworkDisk(t *testing.T) {
	f := newFixture(t)
	bundle := filepath.Join(f.data, "vm_bundles", "claudevm.bundle")
	must(t, os.MkdirAll(bundle, 0o700))
	if c, ok := find(f.Doctor(), i18n.T("ops.doctor.cowork")); !ok || c.Level != Warn {
		t.Fatal("a Cowork bundle without sessiondata.vhdx should warn")
	}
	f.put(filepath.Join(bundle, "sessiondata.vhdx"), "")
	if _, ok := find(f.Doctor(), i18n.T("ops.doctor.cowork")); ok {
		t.Fatal("a complete bundle should not warn")
	}
}

func TestDoctorLeftoverLogin(t *testing.T) {
	f := newFixture(t)
	f.Install.Kind = platform.MSIX
	roaming := t.TempDir()
	t.Setenv("APPDATA", roaming)
	old := filepath.Join(roaming, "Claude")
	f.put(filepath.Join(old, "claude_desktop_config.json"), "{}")
	title := i18n.T("ops.doctor.leftover")
	if _, ok := find(f.Doctor(), title); ok {
		t.Fatal("a settings-only folder is not an old login")
	}
	f.put(filepath.Join(old, "config.json"), "{}")
	if _, ok := find(f.Doctor(), title); !ok {
		t.Fatal("legacy overlay should warn about an old login")
	}
	f.Install.DataDir = filepath.Join(t.TempDir(), "Claude-Data", "claude_pzs8sxrjxfjjc")
	if _, ok := find(f.Doctor(), title); ok {
		t.Fatal("migrated data does not use the legacy overlay")
	}
	f.Install.DataDir = old
	if _, ok := find(f.Doctor(), title); ok {
		t.Fatal("the active folder is not its own leftover")
	}
}
