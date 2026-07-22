package ui

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/term"
)

type Item struct {
	Label    string
	Hint     string
	Selected bool
}

type keyAction int

const (
	keyIgnore keyAction = iota
	keyUp
	keyDown
	keyToggle
	keySelectAll
	keySelectNone
	keyInvert
	keyConfirm
	keyCancel
)

var (
	cReset = "\x1b[0m"
	cCyan  = "\x1b[36m"
	cGreen = "\x1b[32m"
	cDim   = "\x1b[90m"
)

func init() {
	if os.Getenv("NO_COLOR") != "" {
		cReset, cCyan, cGreen, cDim = "", "", "", ""
	}
}

func ToggleList(title string, items []Item) []Item {
	if len(items) == 0 {
		return items
	}

	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) || !enableVTOutput() {
		return toggleListNumbered(title, items)
	}

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return toggleListNumbered(title, items)
	}
	enableVTInput()

	cursor, top := 0, 0
	lines := renderList(title, items, cursor, top, false, 0)
	pending := make([]byte, 0, 16)
	readBuf := make([]byte, 16)
	for {
		action, consumed, needMore := nextAction(pending)
		if needMore {
			n, err := os.Stdin.Read(readBuf)
			if err != nil || n == 0 {
				term.Restore(fd, oldState)
				fmt.Print("\r\n")
				return items
			}
			pending = append(pending, readBuf[:n]...)
			continue
		}
		pending = pending[consumed:]

		newCursor, done, cancelled := applyKey(items, cursor, action)
		cursor = newCursor
		if done {
			term.Restore(fd, oldState)
			fmt.Print("\r\n")
			if cancelled {
				fmt.Fprintln(os.Stderr, "Cancelled.")
				os.Exit(130)
			}
			return items
		}
		if action != keyIgnore {
			_, height := terminalSize()
			top = scrollTo(top, cursor, viewRows(len(items), height), len(items))
			lines = renderList(title, items, cursor, top, true, lines)
		}
	}
}

func nextAction(b []byte) (keyAction, int, bool) {
	if len(b) == 0 {
		return keyIgnore, 0, true
	}
	if b[0] == 0x1b {
		if len(b) < 2 {
			return keyIgnore, 0, true
		}
		if b[1] != '[' {
			return keyIgnore, 1, false
		}
		if len(b) < 3 {
			return keyIgnore, 0, true
		}
		switch b[2] {
		case 'A':
			return keyUp, 3, false
		case 'B':
			return keyDown, 3, false
		default:
			return keyIgnore, 3, false
		}
	}
	return decodeSingle(b[0]), 1, false
}

func decodeSingle(c byte) keyAction {
	switch c {
	case 0x03:
		return keyCancel
	case '\r', '\n':
		return keyConfirm
	case ' ':
		return keyToggle
	case 'k', 'K':
		return keyUp
	case 'j', 'J':
		return keyDown
	case 'a', 'A':
		return keySelectAll
	case 'n', 'N':
		return keySelectNone
	case 'i', 'I':
		return keyInvert
	case 'q', 'Q':
		return keyCancel
	}
	return keyIgnore
}

func applyKey(items []Item, cursor int, action keyAction) (int, bool, bool) {
	n := len(items)
	switch action {
	case keyUp:
		cursor = (cursor - 1 + n) % n
	case keyDown:
		cursor = (cursor + 1) % n
	case keyToggle:
		items[cursor].Selected = !items[cursor].Selected
	case keySelectAll:
		setAll(items, true)
	case keySelectNone:
		setAll(items, false)
	case keyInvert:
		invertAll(items)
	case keyConfirm:
		return cursor, true, false
	case keyCancel:
		return cursor, true, true
	}
	return cursor, false, false
}

func renderList(title string, items []Item, cursor, top int, redraw bool, prevLines int) int {
	width, height := terminalSize()
	rows := viewRows(len(items), height)

	var b strings.Builder
	if redraw {
		fmt.Fprintf(&b, "\x1b[%dA\x1b[J", prevLines)
	}

	writeRow(&b, title, width)
	writeRow(&b, scrollMarker(top, 0, cDim+"  ↑ %d more above"+cReset), width)

	for i := top; i < top+rows && i < len(items); i++ {
		it := items[i]
		pointer, box := "  ", "[ ]"
		if it.Selected {
			box = "[x]"
		}

		label, hint := fitLine(it.Label, it.Hint, width-len(pointer)-len(box)-1)
		if it.Selected {
			box = cGreen + box + cReset
		}
		if i == cursor {
			pointer = cCyan + "> " + cReset
			label = cCyan + label + cReset
		}

		line := pointer + box + " " + label
		if hint != "" {
			line += "  " + cDim + hint + cReset
		}
		b.WriteString(line + "\r\n")
	}

	below := len(items) - (top + rows)
	writeRow(&b, scrollMarker(below, 0, cDim+"  ↓ %d more below"+cReset), width)
	b.WriteString("\r\n")
	writeRow(&b, cDim+"space toggle  a all  n none  i invert  enter confirm  j/k move  q cancel"+cReset, width)

	fmt.Print(b.String())
	return rows + chromeLines
}

// scrollMarker returns an empty line rather than nothing when there is no
// overflow, so the drawn block keeps a constant height and the cursor-up
// redraw stays correct.
func scrollMarker(count, threshold int, format string) string {
	if count <= threshold {
		return ""
	}
	return fmt.Sprintf(format, count)
}

// writeRow emits exactly one screen line. Anything wider than the terminal
// would wrap onto a second line and desynchronise the redraw.
func writeRow(b *strings.Builder, text string, width int) {
	b.WriteString(clampWidth(text, width) + "\r\n")
}

func toggleListNumbered(title string, items []Item) []Item {
	for {
		fmt.Println()
		fmt.Println(title)
		for i, item := range items {
			mark := " "
			if item.Selected {
				mark = "x"
			}
			hint := ""
			if item.Hint != "" {
				hint = "  (" + item.Hint + ")"
			}
			fmt.Printf("  [%s] %2d. %s%s\n", mark, i+1, item.Label, hint)
		}
		fmt.Println()
		fmt.Println("Type numbers to toggle (e.g. 1 3 5), 'a' all, 'n' none, 'i' invert, Enter to confirm.")

		answer := Ask("> ")
		if answer == "" {
			return items
		}

		switch strings.ToLower(answer) {
		case "a":
			setAll(items, true)
			continue
		case "n":
			setAll(items, false)
			continue
		case "i":
			invertAll(items)
			continue
		}

		for _, token := range strings.Fields(answer) {
			if n, err := strconv.Atoi(token); err == nil && n >= 1 && n <= len(items) {
				items[n-1].Selected = !items[n-1].Selected
			}
		}
	}
}

func setAll(items []Item, value bool) {
	for i := range items {
		items[i].Selected = value
	}
}

func invertAll(items []Item) {
	for i := range items {
		items[i].Selected = !items[i].Selected
	}
}
