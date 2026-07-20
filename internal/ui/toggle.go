package ui

import (
	"fmt"
	"strconv"
	"strings"
)

type Item struct {
	Label    string
	Hint     string
	Selected bool
}

func SelectAll(items []Item) []Item {
	setAll(items, true)
	return items
}

func ToggleList(title string, items []Item) []Item {
	if len(items) == 0 {
		return items
	}

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
		fmt.Println("Type numbers to toggle (e.g. 1 3 5), 'a' all, 'n' none, Enter to confirm.")

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
