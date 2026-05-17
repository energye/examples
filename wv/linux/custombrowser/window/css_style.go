package window

import (
	gtk3 "github.com/energye/energy/v3/platform/linux/gtk3/cgo"
	gtk3types "github.com/energye/energy/v3/platform/linux/types"
)

func addCSSStyles() {
	provider := gtk3.NewCssProvider()
	defer provider.Unref()
	css := `
.tab {
	background: rgba(56, 57, 60, 1);
	border: 0;
	border-radius: 4px 4px 0 0;
	margin-top: 2px;
	padding: 4px 8px;
	color: #FFFFFF;
	transition: all 0.2s ease;
}

.tab.active {
	background: rgba(80, 80, 80, 0.7);
}

.tab.click {
	background: rgba(80, 85, 85, 1);
}

.tab.inactive {
	background: rgba(56, 57, 60, 1);
}

.tab-close-button {
	border-radius: 2px;
	border: none;
	background: transparent;
	padding: 2px;
	min-width: 16px;
	min-height: 16px;
	color: #FFFFFF;
	transition: background-color 0.1s;
}

.tab-close-button:hover {
	background-color: rgba(100, 100, 100, 1);
}

.tab-close-button:active {
	background-color: rgba(100, 120, 130, 1);
}
`
	provider.LoadFromData(css)

	screen := gtk3.ScreenGetDefault()
	gtk3.AddProviderForScreen(screen.(*gtk3.Screen), provider, gtk3types.STYLE_PROVIDER_PRIORITY_APPLICATION)
}
