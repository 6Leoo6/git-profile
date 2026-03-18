//go:build windows

package main

import (
	"os"

	"golang.org/x/sys/windows"
)

// enableANSI enables virtual terminal processing on Windows so that
// ANSI escape codes are rendered as colors rather than raw characters.
func enableANSI() {
	stdout := windows.Handle(os.Stdout.Fd())
	var mode uint32
	windows.GetConsoleMode(stdout, &mode)
	windows.SetConsoleMode(stdout, mode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
}
