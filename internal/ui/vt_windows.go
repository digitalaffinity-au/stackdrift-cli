//go:build windows

package ui

import (
	"os"

	"golang.org/x/sys/windows"
)

func enableVTOutput() bool {
	h := windows.Handle(os.Stdout.Fd())
	var mode uint32
	if err := windows.GetConsoleMode(h, &mode); err != nil {
		return false
	}
	if err := windows.SetConsoleMode(h, mode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING); err != nil {
		return false
	}
	return true
}

func enableVTInput() {
	h := windows.Handle(os.Stdin.Fd())
	var mode uint32
	if err := windows.GetConsoleMode(h, &mode); err != nil {
		return
	}
	windows.SetConsoleMode(h, mode|windows.ENABLE_VIRTUAL_TERMINAL_INPUT)
}
