package window

import "github.com/energye/examples/wv/linux/gtkhelper"

func addCSSStyles() {
	provider := gtkhelper.NewCssProvider()
	defer provider.Unref()
	css := `
.tab {
	background-color: #f0f0f0;
	border: 1px solid #dddddd;
	border-bottom: none;
	border-radius: 4px 4px 0 0;
	margin-top: 2px;
	padding: 4px 8px;
	color: #333333;
	transition: all 0.2s ease;
}

.tab.active {
	background-color: #ffffff;
}

.tab.inactive {
	background-color: #f8f8f8;
}

.tab-close-button {
	border-radius: 2px;
	border: none;
	background: transparent;
	padding: 2px;
	min-width: 16px;
	min-height: 16px;
	transition: background-color 0.1s;
}

.tab-close-button:hover {
	background-color: rgba(0, 0, 0, 0.1);
}

.tab-close-button:active {
	background-color: rgba(0, 0, 0, 0.2);
}

	`
	provider.LoadFromData(css)

	screen := gtkhelper.ScreenGetDefault()
	gtkhelper.AddProviderForScreen(screen, provider, gtkhelper.STYLE_PROVIDER_PRIORITY_APPLICATION)
}
