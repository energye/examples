//go:build !windows

package main

import "fmt"

func startMouseTest(s *Scene) {
	fmt.Println("mousetest is Windows-only (it injects WM_MOUSEMOVE messages)")
}
