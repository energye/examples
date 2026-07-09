module github.com/energye/examples

go 1.20

require (
	github.com/ebitengine/purego v0.10.0
	github.com/energye/assetserve v0.0.0-20240622112126-c31d9026b671
	github.com/energye/cef v1.0.5
	github.com/energye/energy/v3 v3.0.0
	github.com/energye/lcl v1.0.9
	github.com/energye/widget v1.0.3
	github.com/energye/wv v1.0.10
	github.com/go-gl/gl v0.0.0-20260331235117-4566fea9a276
	github.com/go-gl/mathgl v1.2.0
	github.com/go-ole/go-ole v1.3.0
	github.com/goki/freetype v1.0.5
	golang.org/x/image v0.15.0
)

require (
	github.com/godbus/dbus/v5 v5.2.2 // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
)

replace (
	github.com/energye/cef => ../cef
	github.com/energye/energy/v3 => ../energy
	github.com/energye/lcl => ../lcl
)
