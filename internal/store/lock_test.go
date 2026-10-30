package store

import "testing"

func TestVaultLock(t *testing.T) {
	v := Vault{Dir: t.TempDir()}
	release, err := v.Lock()
	must(t, err)
	if unlock, err := v.Lock(); err == nil {
		unlock()
		release()
		t.Fatal("a second operation acquired the same vault")
	}
	release()
	unlock, err := v.Lock()
	must(t, err)
	unlock()
}
