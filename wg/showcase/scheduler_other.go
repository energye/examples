//go:build !windows

package main

func setTimerResolution(ms uint32) {}
func restoreTimerResolution()      {}
