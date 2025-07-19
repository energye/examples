package main

import (
	"fmt"
	. "github.com/energye/examples/syso"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/pkgs/win"
)

// 注册表操作示例

func init() {
	TestLoadLibPath()
}
func main() {
	lcl.Init(nil, nil)
	// 64位下传入KEY_WOW64_64KEY
	//reg := lcl.NewRegistry(win.KEY_ALL_ACCESS|win.KEY_WOW64_64KEY)
	reg := lcl.NewRegistryWithLongword(win.KEY_ALL_ACCESS | win.KEY_WOW64_64KEY)
	defer reg.Free()
	reg.SetRootKey(win.HKEY_CURRENT_USER)
	if reg.OpenKeyReadOnlyWithString("SOFTWARE\\Google\\Chrome\\BLBeacon") {
		defer reg.CloseKey()
		fmt.Println("version:", reg.ReadStringWithString("version"))
		fmt.Println("state:", reg.ReadIntegerWithString("state"))
		fmt.Println("BLBeacon Exists:", reg.KeyExistsWithString("BLBeacon"))
		fmt.Println("failed_count Exists:", reg.ValueExistsWithString("failed_count"))
		//
		// reg.WriteBool()
	} else {
		fmt.Println("打开失败！")
	}
}
