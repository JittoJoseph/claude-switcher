package store

import (
	"fmt"
	"os"
	"path/filepath"
)

// Lock serializes mutations across processes that share this vault. The OS releases
// the lock on exit or crash; the persistent lock file must not be removed.
func (v Vault) Lock() (func(), error) {
	if v.Dir == "" {
		return nil, fmt.Errorf("lock vault: empty path")
	}
	if err := os.MkdirAll(v.Dir, 0o700); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(filepath.Join(v.Dir, "_lock"), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	if err := lockFile(f); err != nil {
		f.Close()
		return nil, fmt.Errorf("another switcher operation holds the vault, or locking failed: %w", err)
	}
	return func() { f.Close() }, nil
}
