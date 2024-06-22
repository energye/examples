package main

import (
	"fmt"
	_ "github.com/energye/examples/syso"
	"github.com/energye/lcl/inits"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"time"
)

func main() {
	inits.Init(nil, nil)
	iniFile := lcl.NewIniFile(".\\test.ini", types.NewSet())
	defer iniFile.Free()

	iniFile.WriteBool("First", "Bool", true)
	iniFile.WriteString("First", "String", "这是字符串")
	iniFile.WriteDateTime("First", "Time", types.TDateTime(time.Now().Unix()))
	iniFile.WriteInteger("First", "Integer", 123456)
	iniFile.WriteFloat("First", "Float", 1.2555)

	fmt.Println("Bool:", iniFile.ReadBool("First", "Bool", false))
	fmt.Println("String:", iniFile.ReadString("First", "String", ""))
	fmt.Println("Time:", iniFile.ReadDate("First", "Time", types.TDateTime(time.Now().Unix())))
	fmt.Println("Integer:", iniFile.ReadInteger("First", "Integer", 0))
	fmt.Println("Float:", iniFile.ReadFloat("First", "Float", 0.0))
}
