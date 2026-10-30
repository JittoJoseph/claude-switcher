package store

import (
	"golang.org/x/sys/windows"
	"os"
	"path/filepath"
	"testing"
)

// A config.json that cannot be removed must leave every other login item in place.
func TestClearLoginStopsAtLockedConfig(t *testing.T) {
	data := filepath.Join(t.TempDir(), "Claude")
	fillLogin(t, data, "work")
	before := login(t, data)
	f, err := os.Open(filepath.Join(data, "config.json")) // on Windows os.Open withholds delete sharing
	must(t, err)
	defer f.Close()
	if ClearLogin(data, items) == nil {
		t.Fatal("removing a locked config.json should fail")
	}
	if login(t, data) != before {
		t.Error("items after config.json were removed")
	}
}

func TestSaveLoginCopyFailureKeepsSavedProfile(t *testing.T) {
	v := Vault{Dir: t.TempDir()}
	data := t.TempDir()
	fillLogin(t, data, "before")
	must(t, v.SaveLogin("work", data, items))
	saved := hash(t, filepath.Join(v.Dir, "work"))
	fillLogin(t, data, "after")
	// Opening with no sharing simulates Electron/antivirus locking a later item.
	ptr, err := windows.UTF16PtrFromString(filepath.Join(data, "Network", "Cookies"))
	must(t, err)
	h, err := windows.CreateFile(ptr, windows.GENERIC_READ, 0, nil, windows.OPEN_EXISTING, 0, 0)
	must(t, err)
	defer windows.CloseHandle(h)
	if v.SaveLogin("work", data, items) == nil {
		t.Fatal("locked source should fail")
	}
	if hash(t, filepath.Join(v.Dir, "work")) != saved {
		t.Fatal("failed copy damaged saved profile")
	}
}

func TestRestoreLoginRollsBackOnLockedItem(t *testing.T) {
	v := Vault{Dir: t.TempDir()}
	data := t.TempDir()
	fillLogin(t, data, "saved")
	must(t, v.SaveLogin("work", data, items))
	fillLogin(t, data, "live")
	before := login(t, data)
	f, err := os.Open(filepath.Join(data, "Preferences"))
	must(t, err)
	defer f.Close()
	if v.RestoreLogin("work", data, items) == nil {
		t.Fatal("locked destination should fail")
	}
	if login(t, data) != before {
		t.Fatal("failed restore did not put back original login")
	}
}

func TestClearLoginRollsBackOnLaterLockedItem(t *testing.T) {
	data := t.TempDir()
	fillLogin(t, data, "live")
	before := login(t, data)
	f, err := os.Open(filepath.Join(data, "Preferences"))
	must(t, err)
	defer f.Close()
	if ClearLogin(data, items) == nil {
		t.Fatal("locked item should fail")
	}
	if login(t, data) != before {
		t.Fatal("failed clear lost the original login")
	}
}
