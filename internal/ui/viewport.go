package ui

import (
	"os"
	"strings"

	"golang.org/x/term"
)

const (
	fallbackWidth  = 80
	fallbackHeight = 24

	// Title, the two scroll markers, a blank line and the key help.
	chromeLines  = 5
	minViewRows  = 1
	minTextWidth = 20
)

// terminalSize reports the usable screen. The redraw moves the cursor up by a
// fixed number of lines, so anything that makes the drawn block taller than the
// screen, or wider than it and therefore wrapped, corrupts that arithmetic.
func terminalSize() (int, int) {
	for _, f := range []*os.File{os.Stdout, os.Stdin} {
		if width, height, err := term.GetSize(int(f.Fd())); err == nil && width > 0 && height > 0 {
			return width, height
		}
	}
	return fallbackWidth, fallbackHeight
}

func viewRows(total, height int) int {
	rows := height - chromeLines
	if rows < minViewRows {
		rows = minViewRows
	}
	if rows > total {
		rows = total
	}
	return rows
}

// scrollTo keeps the cursor inside the visible window, moving the window as
// little as possible.
func scrollTo(top, cursor, rows, total int) int {
	if rows >= total {
		return 0
	}
	if cursor < top {
		top = cursor
	}
	if cursor >= top+rows {
		top = cursor - rows + 1
	}
	if top > total-rows {
		top = total - rows
	}
	if top < 0 {
		top = 0
	}
	return top
}

// fitLine trims a label and its hint to the space left on one screen row. The
// hint gives way first, since the label identifies the item.
func fitLine(label, hint string, budget int) (string, string) {
	if budget <= 0 {
		return "", ""
	}

	labelRunes := []rune(label)
	if len(labelRunes) > budget {
		return truncate(labelRunes, budget), ""
	}

	remaining := budget - len(labelRunes) - 2
	if hint == "" || remaining < minTextWidth/4 {
		return label, ""
	}

	hintRunes := []rune(hint)
	if len(hintRunes) > remaining {
		return label, truncate(hintRunes, remaining)
	}
	return label, hint
}

func truncate(runes []rune, budget int) string {
	if budget <= 1 {
		return string(runes[:budget])
	}
	return string(runes[:budget-1]) + "…"
}

// clampWidth cuts a rendered line to width visible columns. Colour escapes are
// copied through without counting, since they occupy no space on screen, and a
// reset is appended if the cut lands inside styled text.
func clampWidth(s string, width int) string {
	if width <= 0 {
		return s
	}

	var out strings.Builder
	runes := []rune(s)
	visible, styled := 0, false

	for i := 0; i < len(runes); i++ {
		if runes[i] == 0x1b {
			end := i
			for end < len(runes) && runes[end] != 'm' {
				end++
			}
			if end == len(runes) {
				break
			}
			out.WriteString(string(runes[i : end+1]))
			styled = true
			i = end
			continue
		}

		if visible == width {
			if styled {
				out.WriteString(cReset)
			}
			return out.String()
		}
		out.WriteRune(runes[i])
		visible++
	}
	return out.String()
}
