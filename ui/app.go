package ui

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/eswai/pathfinder2/core"
	"github.com/gdamore/tcell/v2"
)

type focus int

const (
	focusBookmarks focus = iota
	focusFilelist
	focusBuffer
)

type bmState struct {
	curDir   string
	flCursor int
	flScroll int
}

type App struct {
	screen       tcell.Screen
	bookmarks    *core.Bookmarks
	bmCursor     int // bookmark list cursor
	fileList     []core.Entry
	curDir       string
	flCursor     int // filelist cursor
	flScroll     int // filelist scroll offset
	bmScroll     int // bookmark scroll offset
	focused      focus
	bmStateMap   map[string]bmState // per-bookmark saved state, keyed by bookmark path
	buffer       []string           // staged paths for move/copy
	bufCursor    int                // buffer pane cursor
}

func NewApp() (*App, error) {
	s, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	if err := s.Init(); err != nil {
		return nil, err
	}
	s.EnableMouse()
	s.SetStyle(tcell.StyleDefault.Background(tcell.NewRGBColor(0x2E, 0x34, 0x40)).Foreground(tcell.NewRGBColor(0xD8, 0xDE, 0xE9)))
	s.Clear()

	bm := core.LoadBookmarks()
	startDir := core.HomeDir()
	if len(bm.Paths) > 0 {
		startDir = bm.Paths[0]
	}

	app := &App{
		screen:     s,
		bookmarks:  bm,
		curDir:     startDir,
		focused:    focusBookmarks,
		bmStateMap: make(map[string]bmState),
	}
	app.reloadDir()
	return app, nil
}

func (app *App) reloadDir() {
	entries, err := core.ListDir(app.curDir)
	if err != nil {
		entries = []core.Entry{{Name: "..", IsDir: true}}
	}
	app.fileList = entries
	if app.flCursor >= len(app.fileList) {
		app.flCursor = len(app.fileList) - 1
	}
	if app.flCursor < 0 {
		app.flCursor = 0
	}
}

func (app *App) Run() {
	defer app.screen.Fini()
	for {
		app.draw()
		ev := app.screen.PollEvent()
		switch e := ev.(type) {
		case *tcell.EventResize:
			app.screen.Sync()
		case *tcell.EventKey:
			if !app.handleKey(e) {
				return
			}
		}
	}
}

func (app *App) handleKey(ev *tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyCtrlC:
		return false
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'q':
			return false
		case 'a':
			if app.focused == focusFilelist {
				app.bookmarks.Add(app.curDir)
			}
		case 'd':
			if app.focused == focusBookmarks {
				app.bookmarks.Delete(app.bmCursor)
				if app.bmCursor >= len(app.bookmarks.Paths) && app.bmCursor > 0 {
					app.bmCursor--
				}
				app.clampBmScroll()
			}
		case 't':
			if app.focused == focusFilelist {
				app.trashSelected()
			}
		case 'b':
			if app.focused == focusFilelist {
				app.bufferSelected()
			}
		case 'm':
			if app.focused == focusFilelist {
				app.executeMove()
			}
		case 'c':
			if app.focused == focusFilelist {
				app.executeCopy()
			}
		case 'z':
			if app.focused == focusBuffer {
				app.removeFromBuffer()
			}
		}
	case tcell.KeyTab:
		switch app.focused {
		case focusBookmarks:
			app.focused = focusFilelist
		case focusFilelist:
			if len(app.buffer) > 0 {
				app.focused = focusBuffer
			} else {
				app.focused = focusBookmarks
			}
		case focusBuffer:
			app.focused = focusBookmarks
		}
	case tcell.KeyUp:
		app.moveUp()
	case tcell.KeyDown:
		app.moveDown()
	case tcell.KeyRight, tcell.KeyEnter:
		app.enterOrSelect()
	case tcell.KeyLeft:
		if app.focused == focusFilelist {
			app.navigateUp()
		}
	case tcell.KeyPgUp:
		app.pageUp()
	case tcell.KeyPgDn:
		app.pageDown()
	case tcell.KeyHome:
		app.moveTop()
	case tcell.KeyEnd:
		app.moveBottom()
	}
	return true
}

func (app *App) pageUp() {
	_, h := app.screen.Size()
	pageSize := h - 2
	if pageSize < 1 {
		pageSize = 1
	}
	if app.focused == focusFilelist {
		app.flCursor -= pageSize
		if app.flCursor < 0 {
			app.flCursor = 0
		}
		app.clampFlScroll()
	} else {
		app.saveCurrentBmState()
		app.bmCursor -= pageSize
		if app.bmCursor < 0 {
			app.bmCursor = 0
		}
		app.clampBmScroll()
		app.syncFromBookmark()
	}
}

func (app *App) pageDown() {
	_, h := app.screen.Size()
	pageSize := h - 2
	if pageSize < 1 {
		pageSize = 1
	}
	if app.focused == focusFilelist {
		app.flCursor += pageSize
		if app.flCursor >= len(app.fileList) {
			app.flCursor = len(app.fileList) - 1
		}
		app.clampFlScroll()
	} else {
		app.saveCurrentBmState()
		app.bmCursor += pageSize
		if app.bmCursor >= len(app.bookmarks.Paths) {
			app.bmCursor = len(app.bookmarks.Paths) - 1
		}
		app.clampBmScroll()
		app.syncFromBookmark()
	}
}

func (app *App) moveTop() {
	if app.focused == focusFilelist {
		app.flCursor = 0
		app.clampFlScroll()
	} else {
		app.saveCurrentBmState()
		app.bmCursor = 0
		app.clampBmScroll()
		app.syncFromBookmark()
	}
}

func (app *App) moveBottom() {
	if app.focused == focusFilelist {
		if len(app.fileList) > 0 {
			app.flCursor = len(app.fileList) - 1
		}
		app.clampFlScroll()
	} else {
		app.saveCurrentBmState()
		if len(app.bookmarks.Paths) > 0 {
			app.bmCursor = len(app.bookmarks.Paths) - 1
		}
		app.clampBmScroll()
		app.syncFromBookmark()
	}
}

func (app *App) moveUp() {
	switch app.focused {
	case focusFilelist:
		if app.flCursor > 0 {
			app.flCursor--
		}
		app.clampFlScroll()
	case focusBuffer:
		if app.bufCursor > 0 {
			app.bufCursor--
		}
	default:
		if app.bmCursor > 0 {
			app.saveCurrentBmState()
			app.bmCursor--
		}
		app.clampBmScroll()
		app.syncFromBookmark()
	}
}

func (app *App) moveDown() {
	switch app.focused {
	case focusFilelist:
		if app.flCursor < len(app.fileList)-1 {
			app.flCursor++
		}
		app.clampFlScroll()
	case focusBuffer:
		if app.bufCursor < len(app.buffer)-1 {
			app.bufCursor++
		}
	default:
		if app.bmCursor < len(app.bookmarks.Paths)-1 {
			app.saveCurrentBmState()
			app.bmCursor++
		}
		app.clampBmScroll()
		app.syncFromBookmark()
	}
}

// saveCurrentBmState persists the center-pane position for the current bookmark.
func (app *App) saveCurrentBmState() {
	if app.bmCursor < len(app.bookmarks.Paths) {
		key := app.bookmarks.Paths[app.bmCursor]
		app.bmStateMap[key] = bmState{
			curDir:   app.curDir,
			flCursor: app.flCursor,
			flScroll: app.flScroll,
		}
	}
}

// syncFromBookmark restores (or initialises) the center-pane for the current bookmark.
func (app *App) syncFromBookmark() {
	if app.bmCursor >= len(app.bookmarks.Paths) {
		return
	}
	key := app.bookmarks.Paths[app.bmCursor]
	if saved, ok := app.bmStateMap[key]; ok {
		app.curDir = saved.curDir
		app.flCursor = saved.flCursor
		app.flScroll = saved.flScroll
	} else {
		app.curDir = key
		app.flCursor = 0
		app.flScroll = 0
	}
	app.reloadDir()
}

func (app *App) enterOrSelect() {
	if app.focused == focusBookmarks {
		app.focused = focusFilelist
		return
	}
	if app.flCursor >= len(app.fileList) {
		return
	}
	e := app.fileList[app.flCursor]
	if !e.IsDir {
		return
	}
	if e.Name == ".." {
		app.navigateUp()
	} else {
		app.curDir = filepath.Join(app.curDir, e.Name)
		app.flCursor = 0
		app.flScroll = 0
		app.reloadDir()
	}
}

func (app *App) navigateUp() {
	parent := filepath.Dir(app.curDir)
	if parent == app.curDir {
		return
	}
	// Do not navigate above the currently selected bookmark's directory.
	if app.bmCursor < len(app.bookmarks.Paths) {
		bmPath := app.bookmarks.Paths[app.bmCursor]
		if !strings.HasPrefix(app.curDir+"/", bmPath+"/") || app.curDir == bmPath {
			return
		}
	}
	app.curDir = parent
	app.flCursor = 0
	app.flScroll = 0
	app.reloadDir()
}

func (app *App) trashSelected() {
	if app.flCursor >= len(app.fileList) {
		return
	}
	e := app.fileList[app.flCursor]
	if e.Name == ".." {
		return
	}
	target := filepath.Join(app.curDir, e.Name)
	if err := core.MoveToTrash(target); err != nil {
		return
	}
	app.reloadDir()
	if app.flCursor >= len(app.fileList) && app.flCursor > 0 {
		app.flCursor--
	}
	app.clampFlScroll()
}

func (app *App) bufferSelected() {
	if app.flCursor >= len(app.fileList) {
		return
	}
	e := app.fileList[app.flCursor]
	if e.Name == ".." {
		return
	}
	path := filepath.Join(app.curDir, e.Name)
	for _, p := range app.buffer {
		if p == path {
			return // already in buffer
		}
	}
	app.buffer = append(app.buffer, path)
}

func (app *App) removeFromBuffer() {
	if len(app.buffer) == 0 {
		return
	}
	app.buffer = append(app.buffer[:app.bufCursor], app.buffer[app.bufCursor+1:]...)
	if app.bufCursor >= len(app.buffer) && app.bufCursor > 0 {
		app.bufCursor--
	}
	if len(app.buffer) == 0 && app.focused == focusBuffer {
		app.focused = focusFilelist
	}
}

func (app *App) executeMove() {
	if len(app.buffer) == 0 {
		return
	}
	for _, src := range app.buffer {
		core.MoveEntry(src, app.curDir)
	}
	app.buffer = app.buffer[:0]
	app.bufCursor = 0
	app.reloadDir()
}

func (app *App) executeCopy() {
	if len(app.buffer) == 0 {
		return
	}
	for _, src := range app.buffer {
		core.CopyEntry(src, app.curDir)
	}
	app.buffer = app.buffer[:0]
	app.bufCursor = 0
	app.reloadDir()
}

func (app *App) clampFlScroll() {
	_, h := app.screen.Size()
	innerH := h - 2 // inside border
	if innerH <= 0 {
		return
	}
	if app.flCursor < app.flScroll {
		app.flScroll = app.flCursor
	}
	if app.flCursor >= app.flScroll+innerH {
		app.flScroll = app.flCursor - innerH + 1
	}
}

func (app *App) clampBmScroll() {
	_, h := app.screen.Size()
	innerH := h - 2
	if innerH <= 0 {
		return
	}
	if app.bmCursor < app.bmScroll {
		app.bmScroll = app.bmCursor
	}
	if app.bmCursor >= app.bmScroll+innerH {
		app.bmScroll = app.bmCursor - innerH + 1
	}
}

// ── Nord color palette ────────────────────────────────────────────────────────

var (
	// Polar Night
	nord0 = tcell.NewRGBColor(0x2E, 0x34, 0x40) // background
	nord1 = tcell.NewRGBColor(0x3B, 0x42, 0x52) // selection bg (unfocused)
	nord3 = tcell.NewRGBColor(0x81, 0xA1, 0xC1) // dim border / dim text
	// Snow Storm
	nord4 = tcell.NewRGBColor(0xD8, 0xDE, 0xE9) // primary text
	nord6 = tcell.NewRGBColor(0xEC, 0xEF, 0xF4) // bright text (selected)
	// Frost
	nord7  = tcell.NewRGBColor(0xD8, 0xDE, 0xE9) // focused border / teal
	nord8  = tcell.NewRGBColor(0x88, 0xC0, 0xD0) // title
	nord9  = tcell.NewRGBColor(0x81, 0xA1, 0xC1) // directory
	nord10 = tcell.NewRGBColor(0x5E, 0x81, 0xAC) // selection bg (focused)
)

// ── styles ────────────────────────────────────────────────────────────────────

var (
	stDefault     = tcell.StyleDefault.Background(nord0).Foreground(nord4)
	stFocusBorder = tcell.StyleDefault.Background(nord0).Foreground(nord7)
	stDimBorder   = tcell.StyleDefault.Background(nord0).Foreground(nord3)
	stTitle       = tcell.StyleDefault.Background(nord0).Foreground(nord8).Bold(true)
	stSelected    = tcell.StyleDefault.Background(nord10).Foreground(nord6).Bold(true)
	stDimSel      = tcell.StyleDefault.Background(nord1).Foreground(nord4)
	stDir         = tcell.StyleDefault.Background(nord0).Foreground(nord9).Bold(true)
	stDimDir      = tcell.StyleDefault.Background(nord0).Foreground(nord9)
)

// ── draw ──────────────────────────────────────────────────────────────────────

type paneSpec struct {
	x0, y0, x1, y1 int
	focused         bool
	draw            func(x0, y0, x1, y1 int, focused bool)
	border          func(focused bool) tcell.Style
	title           func() string
}

func (app *App) draw() {
	s := app.screen
	s.Clear()
	w, h := s.Size()

	leftW := w / 5
	midW := leftW + (w-leftW)*2/3

	bmFocused := app.focused == focusBookmarks
	flFocused := app.focused == focusFilelist
	bufFocused := app.focused == focusBuffer

	// When buffer is non-empty, split the right column into preview (top) + buffer (bottom).
	bufPaneH := 0
	if len(app.buffer) > 0 {
		bufPaneH = len(app.buffer) + 2 // border top+bottom
		maxBufH := (h - 1) / 3
		if bufPaneH > maxBufH {
			bufPaneH = maxBufH
		}
		if bufPaneH < 4 {
			bufPaneH = 4
		}
	}
	previewY1 := h - 1 - bufPaneH

	panes := []paneSpec{
		{
			x0: 0, y0: 0, x1: leftW, y1: h - 1,
			focused: bmFocused,
			draw:    app.drawBookmarks,
			border: func(f bool) tcell.Style {
				if f {
					return stFocusBorder
				}
				return stDimBorder
			},
			title: func() string { return "Bookmarks" },
		},
		{
			x0: leftW, y0: 0, x1: midW, y1: h - 1,
			focused: flFocused,
			draw:    app.drawFilelist,
			border: func(f bool) tcell.Style {
				if f {
					return stFocusBorder
				}
				return stDimBorder
			},
			title: func() string { return app.curDir },
		},
		{
			x0: midW, y0: 0, x1: w - 1, y1: previewY1,
			focused: false,
			draw:    func(x0, y0, x1, y1 int, _ bool) { app.drawPreview(x0, y0, x1, y1) },
			border:  func(_ bool) tcell.Style { return stDimBorder },
			title:   app.previewTitle,
		},
	}

	if len(app.buffer) > 0 {
		panes = append(panes, paneSpec{
			x0: midW, y0: h - 1 - bufPaneH, x1: w - 1, y1: h - 1,
			focused: bufFocused,
			draw:    app.drawBuffer,
			border: func(f bool) tcell.Style {
				if f {
					return stFocusBorder
				}
				return stDimBorder
			},
			title: func() string { return "Buffer" },
		})
	}

	// Draw non-focused panes first, focused pane last so its border wins at shared edges.
	for _, p := range panes {
		if !p.focused {
			drawBox(s, p.x0, p.y0, p.x1, p.y1, p.border(false), stTitle, p.title())
			p.draw(p.x0+1, p.y0+1, p.x1, p.y1, false)
		}
	}
	for _, p := range panes {
		if p.focused {
			drawBox(s, p.x0, p.y0, p.x1, p.y1, p.border(true), stTitle, p.title())
			p.draw(p.x0+1, p.y0+1, p.x1, p.y1, true)
		}
	}

	s.Show()
}

func (app *App) drawBookmarks(x0, y0, x1, y1 int, focused bool) {
	innerH := y1 - y0
	for i := 0; i < innerH; i++ {
		idx := app.bmScroll + i
		if idx >= len(app.bookmarks.Paths) {
			break
		}
		p := app.bookmarks.Paths[idx]
		var st tcell.Style
		switch {
		case idx == app.bmCursor && focused:
			st = stSelected
		case idx == app.bmCursor:
			st = stDimSel
		default:
			st = stDefault
		}
		drawText(app.screen, x0, y0+i, x1, p, st)
	}
}

func (app *App) drawFilelist(x0, y0, x1, y1 int, focused bool) {
	innerH := y1 - y0
	for i := 0; i < innerH; i++ {
		idx := app.flScroll + i
		if idx >= len(app.fileList) {
			break
		}
		e := app.fileList[idx]
		name := e.Name
		if e.IsDir {
			name += "/"
		}
		var st tcell.Style
		switch {
		case idx == app.flCursor && focused:
			st = stSelected
		case idx == app.flCursor:
			st = stDimSel
		case e.IsDir && focused:
			st = stDir
		case e.IsDir:
			st = stDimDir
		default:
			st = stDefault
		}
		drawText(app.screen, x0, y0+i, x1, name, st)
	}
}

func (app *App) previewTitle() string {
	if app.flCursor < len(app.fileList) {
		return app.fileList[app.flCursor].Name
	}
	return ""
}

func (app *App) drawBuffer(x0, y0, x1, y1 int, focused bool) {
	innerH := y1 - y0
	for i := 0; i < innerH && i < len(app.buffer); i++ {
		var st tcell.Style
		switch {
		case i == app.bufCursor && focused:
			st = stSelected
		case i == app.bufCursor:
			st = stDimSel
		default:
			st = stDefault
		}
		drawText(app.screen, x0, y0+i, x1, app.buffer[i], st)
	}
}

func (app *App) drawPreview(x0, y0, x1, y1 int) {
	if app.flCursor >= len(app.fileList) {
		return
	}
	e := app.fileList[app.flCursor]
	innerH := y1 - y0

	target := filepath.Join(app.curDir, e.Name)
	if e.Name == ".." {
		target = filepath.Dir(app.curDir)
	}

	if e.IsDir {
		entries, err := os.ReadDir(target)
		if err != nil {
			drawText(app.screen, x0, y0, x1, "Error: "+err.Error(), stDefault)
			return
		}
		for i := 0; i < innerH && i < len(entries); i++ {
			name := entries[i].Name()
			if entries[i].IsDir() {
				name += "/"
			}
			drawText(app.screen, x0, y0+i, x1, name, stDefault)
		}
		return
	}

	if core.IsBinary(target) {
		drawText(app.screen, x0, y0, x1, "Binary file", stDefault)
		return
	}

	lines, err := core.ReadPreview(target, innerH)
	if err != nil {
		drawText(app.screen, x0, y0, x1, "Error: "+err.Error(), stDefault)
		return
	}
	for i, l := range lines {
		drawText(app.screen, x0, y0+i, x1, l, stDefault)
	}
}
