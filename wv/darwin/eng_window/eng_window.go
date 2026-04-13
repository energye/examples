package main

import (
	"encoding/json"
	"fmt"
	"github.com/energye/energy/v3/application"
	"github.com/energye/examples/wv/darwin/eng_window/src"
	"github.com/energye/lcl/lcl"
	"os/exec"
)

func main() {
	env, err := GetGoEnv()
	fmt.Println(env["CGO_ENABLED"], err)
	application.GApplication = &application.Application{
		Options: application.Options{
			Frameless: true,
			Windows:   application.Windows{Theme: application.Dark},
		},
	}
	lcl.Init(nil, nil)
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)
	lcl.Application.NewForm(&src.MainForm)
	lcl.Application.Run()
}

// GetGoEnv 获取当前 Go 环境的所有配置 (永不返回空)
func GetGoEnv() (map[string]string, error) {
	// 执行 go env -json 输出完整环境（官方接口）
	cmd := exec.Command("go", "env", "-json")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// 解析成 map
	var env map[string]string
	if err := json.Unmarshal(output, &env); err != nil {
		return nil, err
	}

	return env, nil
}
