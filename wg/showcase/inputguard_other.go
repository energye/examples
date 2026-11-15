//go:build !windows

package main

// Non-Windows builds have no native mouse state; the guard is inert there.
func mouseLeftDown() bool { return true }
