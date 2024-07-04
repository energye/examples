package utils

import _ "embed"

var (
	//go:embed ipc.js
	IPCJavaScript []byte
	//go:embed browser.js
	BrowserJavaScript []byte
)
