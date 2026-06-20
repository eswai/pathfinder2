package core

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Bookmarks struct {
	Paths []string `json:"paths"`
	path  string
}

func LoadBookmarks() *Bookmarks {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = filepath.Join(HomeDir(), ".config")
	}
	p := filepath.Join(configDir, "pathfinder", "bookmarks.json")
	b := &Bookmarks{path: p}

	data, err := os.ReadFile(p)
	if err == nil {
		_ = json.Unmarshal(data, b)
	}
	if len(b.Paths) == 0 {
		b.Paths = []string{HomeDir()}
	}
	return b
}

func (b *Bookmarks) Save() error {
	if err := os.MkdirAll(filepath.Dir(b.path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(b.path, data, 0644)
}

func (b *Bookmarks) Add(path string) {
	for _, p := range b.Paths {
		if p == path {
			return
		}
	}
	b.Paths = append(b.Paths, path)
	_ = b.Save()
}

func (b *Bookmarks) Delete(idx int) {
	if idx < 0 || idx >= len(b.Paths) {
		return
	}
	b.Paths = append(b.Paths[:idx], b.Paths[idx+1:]...)
	_ = b.Save()
}
