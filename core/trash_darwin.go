package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// MoveToTrash moves the given path to the macOS Trash (~/.Trash).
// It uses osascript when available for proper Finder integration; otherwise
// falls back to a direct move into ~/.Trash.
func MoveToTrash(path string) error {
	// Try osascript first (no dependency on external tools).
	script := fmt.Sprintf(`tell application "Finder" to delete POSIX file %q`, path)
	cmd := exec.Command("osascript", "-e", script)
	if err := cmd.Run(); err == nil {
		return nil
	}

	// Fallback: move into ~/.Trash directly.
	trashDir := filepath.Join(HomeDir(), ".Trash")
	if err := os.MkdirAll(trashDir, 0700); err != nil {
		return err
	}
	name := filepath.Base(path)
	dest := filepath.Join(trashDir, name)
	// Avoid collisions.
	if _, err := os.Lstat(dest); err == nil {
		suffix := time.Now().Format("20060102-150405")
		dest = filepath.Join(trashDir, fmt.Sprintf("%s_%s", name, suffix))
	}
	return os.Rename(path, dest)
}
