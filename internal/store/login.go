package store

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// SaveLogin copies Desktop's login items from dataDir into the profile: files are copied,
// folders replaced whole, and items Desktop does not have are removed from the profile.
func (v Vault) SaveLogin(name, dataDir string, items []string) error {
	if v.Dir == "" || dataDir == "" || name == "" {
		return errors.New("save login: empty path")
	}
	if !ValidName(name) {
		return errors.New("save login: invalid profile name")
	}
	st, err := os.Lstat(filepath.Join(dataDir, "config.json"))
	if err != nil {
		return err
	}
	if !st.Mode().IsRegular() {
		return errors.New("save login: config.json is not a regular file")
	}
	dest := filepath.Join(v.Dir, name)
	if err := os.MkdirAll(v.Dir, 0o700); err != nil {
		return err
	}
	// Build the complete snapshot before touching the last working saved login.
	stage, err := os.MkdirTemp(v.Dir, "_login-stage-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	if err := replaceAll(dataDir, stage, items); err != nil {
		return err
	}
	// Preserve profile metadata until the caller updates the account identifier.
	if err := replace(filepath.Join(dest, "_account"), filepath.Join(stage, "_account")); err != nil {
		return err
	}
	backup := stage + ".previous"
	existed := false
	if err := os.Rename(dest, backup); err == nil {
		existed = true
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if err := os.Rename(stage, dest); err != nil {
		if existed {
			err = errors.Join(err, os.Rename(backup, dest))
		}
		return err
	}
	if existed {
		return os.RemoveAll(backup)
	}
	return nil
}

// RestoreLogin makes dataDir signed in as the profile: every login item is replaced by the
// profile's copy, and items the profile lacks are removed, so nothing of the previous account
// lingers.
func (v Vault) RestoreLogin(name, dataDir string, items []string) error {
	if v.Dir == "" || dataDir == "" || name == "" {
		return errors.New("restore login: empty path")
	}
	if !ValidName(name) {
		return errors.New("restore login: invalid profile name")
	}
	source := filepath.Join(v.Dir, name)
	st, err := os.Lstat(filepath.Join(source, "config.json"))
	if err != nil {
		return err
	}
	if !st.Mode().IsRegular() {
		return errors.New("restore login: config.json is not a regular file")
	}
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return err
	}
	stage, err := os.MkdirTemp(dataDir, "_login-restore-*")
	if err != nil {
		return err
	}
	// Copy first: an unreadable saved login must not change Desktop's current login.
	if err := replaceAll(source, stage, items); err != nil {
		os.RemoveAll(stage)
		return err
	}
	defer os.RemoveAll(stage)
	return commitLogin(stage, dataDir, items)
}

// ClearLogin removes login items only after each item can move into a recovery
// folder. A locked item restores all earlier items, including config.json.
func ClearLogin(dataDir string, items []string) error {
	if dataDir == "" {
		return errors.New("clear login: empty path")
	}
	if _, err := os.Stat(dataDir); errors.Is(err, fs.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}
	stage, err := os.MkdirTemp(dataDir, "_login-clear-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	return commitLogin(stage, dataDir, items)
}

// commitLogin applies a prepared snapshot and restores the original on errors.
func commitLogin(stage, dataDir string, items []string) error {
	backup, err := os.MkdirTemp(dataDir, "_login-rollback-*")
	if err != nil {
		return err
	}
	type changedItem struct {
		name    string
		existed bool
	}
	var changed []changedItem
	for _, item := range items {
		dest := filepath.Join(dataDir, item)
		old := filepath.Join(backup, item)
		existed := false
		if err = os.Rename(dest, old); err == nil {
			existed = true
		} else if !errors.Is(err, fs.ErrNotExist) {
			break
		}
		changed = append(changed, changedItem{item, existed})
		err = os.Rename(filepath.Join(stage, item), dest)
		if errors.Is(err, fs.ErrNotExist) {
			err = nil
		}
		if err != nil {
			break
		}
	}
	if err != nil {
		var rollbackErr error
		for i := len(changed) - 1; i >= 0; i-- {
			item := changed[i]
			dest := filepath.Join(dataDir, item.name)
			rerr := os.RemoveAll(dest)
			if rerr == nil && item.existed {
				rerr = os.Rename(filepath.Join(backup, item.name), dest)
			}
			rollbackErr = errors.Join(rollbackErr, rerr)
		}
		if rollbackErr != nil {
			return errors.Join(err, fmt.Errorf("login rollback failed; recovery files remain in %s: %w", backup, rollbackErr))
		}
		return errors.Join(err, os.RemoveAll(backup))
	}
	return os.RemoveAll(backup)
}

func replaceAll(from, to string, items []string) error {
	for _, it := range items {
		if err := replace(filepath.Join(from, it), filepath.Join(to, it)); err != nil {
			return err
		}
	}
	return nil
}

// replace makes dst a copy of src, or removes dst when src does not exist.
func replace(src, dst string) error {
	st, err := os.Lstat(src)
	if errors.Is(err, fs.ErrNotExist) {
		return os.RemoveAll(dst)
	}
	if err != nil {
		return err
	}
	if st.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("login item is a symbolic link: %s", src)
	}
	if !st.IsDir() {
		if !st.Mode().IsRegular() {
			return fmt.Errorf("login item is not a regular file: %s", src)
		}
		return copyFile(src, dst)
	}
	if err := os.RemoveAll(dst); err != nil {
		return err
	}
	return copyDir(src, dst)
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		switch {
		case d.IsDir():
			return os.MkdirAll(target, 0o700)
		case d.Type().IsRegular():
			return copyFile(path, target)
		}
		return fmt.Errorf("unsupported login item: %s", path)
	})
}

// copyFile copies one file, readable only by the user: saved logins hold credentials.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o700); err != nil {
		return err
	}
	out, err := os.CreateTemp(filepath.Dir(dst), ".login-copy-*")
	if err != nil {
		return err
	}
	defer os.Remove(out.Name())
	_, err = io.Copy(out, in)
	if cerr := out.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return err
	}
	return os.Rename(out.Name(), dst)
}
