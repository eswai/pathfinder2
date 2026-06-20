package core

import (
	"os"
	"path/filepath"
	"sort"
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
		entry := Entry{Name: e.Name(), IsDir: e.IsDir()}
		if e.IsDir() {
			dirs = append(dirs, entry)
		} else {
			files = append(files, entry)
		}
	}
	sort.Slice(dirs, func(i, j int) bool { return dirs[i].Name < dirs[j].Name })
	sort.Slice(files, func(i, j int) bool { return files[i].Name < files[j].Name })

	result := []Entry{{Name: "..", IsDir: true}}
	result = append(result, dirs...)
	result = append(result, files...)
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

func HomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/"
	}
	return home
}
