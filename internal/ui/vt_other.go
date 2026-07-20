//go:build !windows

package ui

func enableVTOutput() bool { return true }

func enableVTInput() {}
