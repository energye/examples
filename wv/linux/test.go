package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// 获取所有Linux发行版（含国产系统）可能的库文件路径
func getAllPossibleLibraryPaths(libName string) []string {
	var allPaths []string

	// 基础路径（所有发行版通用）
	baseDirs := []string{
		"/lib",
		"/lib64",
		"/usr/lib",
		"/usr/lib64",
		"/usr/local/lib",
		"/usr/local/lib64",
		"/opt/lib",
		"/opt/lib64",
	}

	// 1. 系统标准库路径（所有发行版通用）
	for _, dir := range baseDirs {
		allPaths = append(allPaths, filepath.Join(dir, libName))
	}

	// 2. 架构相关子目录
	architectures := []string{
		"x86_64-linux-gnu",        // x86_64架构（Debian/Ubuntu系）
		"aarch64-linux-gnu",       // ARM64架构（国产系统常用）
		"arm-linux-gnueabihf",     // ARM32架构
		"mips64el-linux-gnuabi64", // MIPS架构
		"loongarch64-linux-gnu",   // 龙芯架构
		"riscv64-linux-gnu",       // RISC-V架构
	}
	for _, dir := range baseDirs {
		for _, arch := range architectures {
			allPaths = append(allPaths, filepath.Join(dir, arch, libName))
		}
	}

	// 3. 发行版特有的路径
	distroSpecificDirs := []string{
		// Debian/Ubuntu系及衍生版
		"/usr/lib/x86_64-linux-gnu/webkit2gtk-4.0",

		// RedHat系及衍生版
		"/usr/lib64/webkit2gtk-4.0",

		// 通用webkit2gtk路径
		"/usr/lib/webkit2gtk-4.0",
		"/usr/local/webkit2gtk/lib",
		"/opt/webkit2gtk/lib",
	}
	for _, dir := range distroSpecificDirs {
		allPaths = append(allPaths, filepath.Join(dir, libName))
	}

	// 4. 国产Linux系统特有路径
	chinaLinuxDirs := []string{
		// 深度/统信UOS
		"/usr/lib/deepin",
		"/opt/apps/*/lib",
		"/opt/dde-libs/lib",

		// 银河麒麟（Kylin）
		"/usr/lib/kylin",
		"/usr/lib/arm-linux-gnueabihf/kylin",

		// 中标麒麟（NeoKylin）
		"/usr/lib/neokylin",
		"/usr/lib64/neokylin",

		// 凝思磐石
		"/usr/lib/think/sys",
	}
	for _, dir := range chinaLinuxDirs {
		allPaths = append(allPaths, filepath.Join(dir, libName))
	}

	// 5. 沙箱环境路径
	sandboxPaths := []string{
		// Snap路径
		"/snap/*/*/usr/lib/*/" + libName,
		"/snap/*/*/usr/lib/" + libName,
		"/snap/*/current/usr/lib/*/" + libName,
		"/snap/*/current/usr/lib/" + libName,

		// Flatpak路径
		"/var/lib/flatpak/app/*/current/active/files/lib/*/" + libName,
		"/var/lib/flatpak/app/*/current/active/files/lib/" + libName,
		"~/.local/share/flatpak/app/*/current/active/files/lib/*/" + libName,
		"~/.local/share/flatpak/app/*/current/active/files/lib/" + libName,
	}
	allPaths = append(allPaths, sandboxPaths...)

	// 处理库文件的常见变体（带版本号等）
	if strings.HasPrefix(libName, "lib") && strings.Contains(libName, ".so") {
		baseLib := libName[:strings.Index(libName, ".so")]
		allPaths = append(allPaths,
			baseLib+".so.*",   // 带版本号的库文件
			baseLib+".so.?",   // 主版本号
			baseLib+".so.?.?", // 主.次版本号
		)
	}

	// 去重并验证路径有效性
	uniquePaths := deduplicatePaths(allPaths)
	return filterExistingPaths(uniquePaths)
}

// 路径去重
func deduplicatePaths(paths []string) []string {
	seen := make(map[string]bool)
	unique := []string{}
	for _, p := range paths {
		// 处理带~的路径（用户主目录）
		expanded, err := expandUser(p)
		if err != nil {
			expanded = p // 处理失败时使用原始路径
		}
		if !seen[expanded] {
			seen[expanded] = true
			unique = append(unique, expanded)
		}
	}
	return unique
}

// 处理用户主目录~符号
func expandUser(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	if path == "~" {
		return home, nil
	}
	return filepath.Join(home, path[1:]), nil
}

// 过滤出实际存在的路径
func filterExistingPaths(paths []string) []string {
	var (
		mu       sync.Mutex
		existing []string
		wg       sync.WaitGroup
	)

	// 并行检查路径，提高效率
	for _, p := range paths {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()

			// 处理Glob模式路径
			if strings.Contains(path, "*") {
				matches, err := filepath.Glob(path)
				if err != nil {
					return // 忽略无效的glob模式
				}

				for _, match := range matches {
					if isLibraryFile(match) {
						mu.Lock()
						existing = append(existing, match)
						mu.Unlock()
					}
				}
			} else {
				// 直接检查文件
				if isLibraryFile(path) {
					mu.Lock()
					existing = append(existing, path)
					mu.Unlock()
				}
			}
		}(p)
	}

	wg.Wait()
	// 再次去重并排序
	return deduplicatePaths(existing)
}

// 检查路径是否为有效的库文件
func isLibraryFile(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}

	// 排除目录，只保留文件或符号链接
	if fileInfo.IsDir() {
		return false
	}

	// 检查是否为共享库（简单判断）
	return strings.HasSuffix(path, ".so") ||
		strings.HasSuffix(path, ".so.") ||
		strings.Contains(path, ".so.")
}

func FindLib(libName string) {
	paths := getAllPossibleLibraryPaths(libName)

	if len(paths) == 0 {
		fmt.Printf("未找到 %s 在任何已知路径中\n", libName)
		return
	}

	fmt.Printf("找到 %d 个 %s 的有效路径：\n", len(paths), libName)
	for i, path := range paths {
		fmt.Printf("%d. %s\n", i+1, path)
	}
}

func main() {
	FindLib("libwebkit2gtk-4.0.so")
	FindLib("libjavascriptcoregtk-4.0.so")
	FindLib("libsoup-2.4.so")
}
