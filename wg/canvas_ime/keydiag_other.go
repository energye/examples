//go:build !windows

package main

import "fmt"

func runKeySelfTest() {
	fmt.Println("keyselftest is Windows-only (it injects Win32 key messages)")
}
