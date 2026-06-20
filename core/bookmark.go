package core

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Bookmark struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

type Bookmarks struct {
	Items []Bookmark `json:"items"`
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
	if len(b.Items) == 0 {
		home := HomeDir()
		b.Items = []Bookmark{{Path: home, Name: filepath.Base(home)}}
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
	for _, item := range b.Items {
		if item.Path == path {
			return
		}
	}
	b.Items = append(b.Items, Bookmark{Path: path, Name: filepath.Base(path)})
	_ = b.Save()
}

func (b *Bookmarks) Delete(idx int) {
	if idx < 0 || idx >= len(b.Items) {
		return
	}
	b.Items = append(b.Items[:idx], b.Items[idx+1:]...)
	_ = b.Save()
}

func (b *Bookmarks) Rename(idx int, name string) {
	if idx < 0 || idx >= len(b.Items) || name == "" {
		return
	}
	b.Items[idx].Name = name
	_ = b.Save()
}

func (b *Bookmarks) Move(idx, delta int) {
	dst := idx + delta
	if dst < 0 || dst >= len(b.Items) {
		return
	}
	b.Items[idx], b.Items[dst] = b.Items[dst], b.Items[idx]
	_ = b.Save()
}
