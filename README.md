# pathfinder

A terminal file browser with bookmarks.

```
┌─Bookmarks──┬─/home/user/projects────────────────┬─main.go──────────────────┐
│ ~/projects │ ../                                 │ package main              │
│ ~/docs     │ core/                               │                           │
│            │ ui/                                 │ import (                  │
│            │ go.mod                              │     "log"                 │
│            │ go.sum                              │                           │
│            │ main.go                             │     "github.com/eswai/... │
└────────────┴─────────────────────────────────────┴───────────────────────────┘
```

## Features

- Three-pane layout: bookmarks, file list, preview
- File and directory preview in the right pane
- Persistent bookmarks saved to `~/.config/pathfinder/bookmarks.json`

## Key Bindings

| Key | Action |
|-----|--------|
| `Tab` | Switch focus between bookmarks and file list |
| `↑` / `↓` | Move cursor |
| `→` / `Enter` | Enter directory / switch focus to file list |
| `←` | Navigate to parent directory |
| `a` | Add current directory to bookmarks (when file list is focused) |
| `d` | Delete selected bookmark (when bookmarks are focused) |
| `q` / `Ctrl+C` | Quit |

## Requirements

- Go 1.21+
