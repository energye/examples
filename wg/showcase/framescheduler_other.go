//go:build !windows

package main

func postFrameMessage(h uintptr) bool { return false }
