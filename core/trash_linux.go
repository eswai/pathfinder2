package core

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// MoveToTrash moves the given path to the XDG Trash directory.
// Follows the FreeDesktop.org Trash specification.
func MoveToTrash(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	trashDir := xdgTrashDir()
	filesDir := filepath.Join(trashDir, "files")
	infoDir := filepath.Join(trashDir, "info")
	for _, d := range []string{filesDir, infoDir} {
		if err := os.MkdirAll(d, 0700); err != nil {
			return err
		}
	}

	name := filepath.Base(abs)
	dest := filepath.Join(filesDir, name)
	// Avoid name collisions.
	if _, err := os.Lstat(dest); err == nil {
		suffix := time.Now().Format("20060102-150405")
		name = fmt.Sprintf("%s_%s", name, suffix)
		dest = filepath.Join(filesDir, name)
	}

	deletionDate := time.Now().Format("2006-01-02T15:04:05")
	infoContent := fmt.Sprintf("[Trash Info]\nPath=%s\nDeletionDate=%s\n", abs, deletionDate)
	infoPath := filepath.Join(infoDir, name+".trashinfo")
	if err := os.WriteFile(infoPath, []byte(infoContent), 0600); err != nil {
		return err
	}

	if err := os.Rename(abs, dest); err != nil {
		_ = os.Remove(infoPath)
		return err
	}
	return nil
}

func xdgTrashDir() string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "Trash")
	}
	return filepath.Join(HomeDir(), ".local", "share", "Trash")
}
