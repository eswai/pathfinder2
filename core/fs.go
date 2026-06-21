package core

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Entry struct {
	Name  string
	IsDir bool
}

func ListDir(dir string) ([]Entry, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var dirs, files []Entry
	for _, e := range entries {
		isDir := e.IsDir()
		if e.Type()&os.ModeSymlink != 0 {
			if info, err := os.Stat(filepath.Join(dir, e.Name())); err == nil {
				isDir = info.IsDir()
			}
		}
		entry := Entry{Name: e.Name(), IsDir: isDir}
		if isDir {
			dirs = append(dirs, entry)
		} else {
			files = append(files, entry)
		}
	}
	sort.Slice(dirs, func(i, j int) bool { return dirs[i].Name < dirs[j].Name })
	sort.Slice(files, func(i, j int) bool { return files[i].Name < files[j].Name })

	result := append(dirs, files...)
	return result, nil
}

func IsBinary(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil {
		return false
	}
	for _, b := range buf[:n] {
		if b == 0 {
			return true
		}
	}
	return false
}

func IsZip(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".zip")
}

func ReadZipPreview(path string, maxLines int) ([]string, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	var lines []string
	for _, f := range r.File {
		if len(lines) >= maxLines {
			break
		}
		size := fmt.Sprintf("%d", f.UncompressedSize64)
		lines = append(lines, fmt.Sprintf("%-40s %s", f.Name, size))
	}
	return lines, nil
}

func ReadPreview(path string, maxLines int) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var lines []string
	start := 0
	for i, b := range data {
		if b == '\n' {
			lines = append(lines, string(data[start:i]))
			start = i + 1
			if len(lines) >= maxLines {
				break
			}
		}
	}
	if start < len(data) && len(lines) < maxLines {
		lines = append(lines, string(data[start:]))
	}
	return lines, nil
}

func Abs(path string) (string, error) {
	return filepath.Abs(path)
}

// MoveEntry moves src into dstDir, using rename where possible.
func MoveEntry(src, dstDir string) error {
	dst := filepath.Join(dstDir, filepath.Base(src))
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	// Cross-device: copy then remove.
	if err := copyEntry(src, dst); err != nil {
		return err
	}
	return os.RemoveAll(src)
}

// CopyEntry recursively copies src into dstDir.
func CopyEntry(src, dstDir string) error {
	return copyEntry(src, filepath.Join(dstDir, filepath.Base(src)))
}

func copyEntry(src, dst string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		if err := os.MkdirAll(dst, info.Mode()); err != nil {
			return err
		}
		entries, err := os.ReadDir(src)
		if err != nil {
			return err
		}
		for _, e := range entries {
			if err := copyEntry(filepath.Join(src, e.Name()), filepath.Join(dst, e.Name())); err != nil {
				return err
			}
		}
		return nil
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func HomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/"
	}
	return home
}
