// rename-module 把项目从当前模块路径整体迁移到新的模块路径：
// 更新 go.mod、所有 Go import，以及 README 等文档里的路径引用。
//
//	go run ./tools/rename-module github.com/yourname/your-project
//	go run ./tools/rename-module -dry-run github.com/yourname/your-project
package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	skipDirs = map[string]bool{
		".git": true, ".idea": true, ".vscode": true, "vendor": true,
		"bin": true, "logs": true, "node_modules": true,
	}
	rewriteExts = map[string]bool{
		".go": true, ".md": true, ".yaml": true, ".yml": true,
	}
)

func main() {
	dryRun := flag.Bool("dry-run", false, "只列出将要修改的文件，不写入")
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "用法: go run ./tools/rename-module [-dry-run] github.com/yourname/your-project")
		os.Exit(1)
	}

	newModule := strings.TrimSuffix(strings.TrimSpace(flag.Arg(0)), "/")
	if newModule == "" || strings.ContainsAny(newModule, " \t\n") {
		fmt.Fprintln(os.Stderr, "模块路径不能为空，也不能包含空白字符")
		os.Exit(1)
	}

	oldModule, err := readModulePath("go.mod")
	if err != nil {
		fmt.Fprintln(os.Stderr, "读取 go.mod 失败:", err)
		os.Exit(1)
	}
	if oldModule == newModule {
		fmt.Printf("模块路径已经是 %s，无需修改\n", newModule)
		return
	}

	files, err := findFiles(".", oldModule)
	if err != nil {
		fmt.Fprintln(os.Stderr, "扫描文件失败:", err)
		os.Exit(1)
	}
	if len(files) == 0 {
		fmt.Println("没有找到引用了旧模块路径的文件")
		return
	}

	if *dryRun {
		fmt.Printf("将把 %s 改为 %s，涉及 %d 个文件：\n", oldModule, newModule, len(files))
		for _, f := range files {
			fmt.Println(" ", f)
		}
		return
	}

	if err := exec.Command("go", "mod", "edit", "-module", newModule).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "更新 go.mod 失败:", err)
		os.Exit(1)
	}
	for _, f := range files {
		if err := replaceInFile(f, oldModule, newModule); err != nil {
			fmt.Fprintf(os.Stderr, "写入 %s 失败: %v\n", f, err)
			os.Exit(1)
		}
	}

	fmt.Printf("模块路径已从 %s 改为 %s，共修改 %d 个文件。\n", oldModule, newModule, len(files))
	fmt.Println("下一步建议执行: go mod tidy && go build ./...")
}

func readModulePath(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(line, "module "); ok {
			return strings.TrimSpace(after), nil
		}
	}
	return "", fmt.Errorf("%s 中没有 module 声明", path)
}

// findFiles 找出引用了旧模块路径的文本文件（go.mod 由 go mod edit 处理）
func findFiles(root, oldModule string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != "." && skipDirs[d.Name()] {
				return fs.SkipDir
			}
			return nil
		}
		if path == "go.mod" || !rewriteExts[filepath.Ext(path)] {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(data), oldModule) {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func replaceInFile(path, oldModule, newModule string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	updated := strings.ReplaceAll(string(data), oldModule, newModule)
	if updated == string(data) {
		return nil
	}
	return os.WriteFile(path, []byte(updated), info.Mode())
}
