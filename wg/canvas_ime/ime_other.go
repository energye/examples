//go:build !windows

package main

// Non-Windows builds compile the same demo; IME composition APIs are Win32
// specific, so this stub keeps the architecture example portable. On GTK/Cocoa
// the equivalent would be GtkIMContext / NSTextInputClient.
func imeComposition(hwnd uintptr) (string, int32, bool) { return "", 0, false }
func imeResult(hwnd uintptr) string                     { return "" }
func imePosition(hwnd uintptr, x, y int32) string       { return "SKIP: Win32 only" }
func imeSelfTest() string                               { return "SKIP: Win32 only" }
func libSelfTest(libPath string) string                 { return "SKIP: Win32 only" }
