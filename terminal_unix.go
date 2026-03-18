//go:build !windows

package main

// enableANSI is a no-op on Unix-like systems — ANSI codes work out of the box.
func enableANSI() {}
