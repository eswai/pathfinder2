package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// drawBox draws a border rectangle. title is drawn in the top edge.
func drawBox(s tcell.Screen, x0, y0, x1, y1 int, borderSt, titleSt tcell.Style, title string) {
	// horizontal edges
	for x := x0 + 1; x < x1; x++ {
		s.SetContent(x, y0, '─', nil, borderSt)
		s.SetContent(x, y1, '─', nil, borderSt)
	}
	// vertical edges
	for y := y0 + 1; y < y1; y++ {
		s.SetContent(x0, y, '│', nil, borderSt)
		s.SetContent(x1, y, '│', nil, borderSt)
	}
	// corners
	s.SetContent(x0, y0, '┌', nil, borderSt)
	s.SetContent(x1, y0, '┐', nil, borderSt)
	s.SetContent(x0, y1, '└', nil, borderSt)
	s.SetContent(x1, y1, '┘', nil, borderSt)

	// title in top edge
	if title != "" {
		tx := x0 + 2
		for _, r := range " " + title + " " {
			if tx >= x1 {
				break
			}
			s.SetContent(tx, y0, r, nil, titleSt)
			tx += runewidth.RuneWidth(r)
		}
	}
}

// drawText writes text inside the pane at row y (content-relative),
// clipping to [x0..x1]. Returns the next x position after the text.
func drawText(s tcell.Screen, x, y int, maxX int, text string, st tcell.Style) {
	cx := x
	for _, r := range text {
		w := runewidth.RuneWidth(r)
		if cx+w > maxX {
			break
		}
		s.SetContent(cx, y, r, nil, st)
		if w == 2 {
			s.SetContent(cx+1, y, ' ', nil, st)
		}
		cx += w
	}
	// clear remainder of the line
	for cx < maxX {
		s.SetContent(cx, y, ' ', nil, st)
		cx++
	}
}
