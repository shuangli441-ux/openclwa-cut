package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// BatchSummary 汇总批量渲染的成功、跳过和失败数量。
type BatchSummary struct {
	Rendered int
	Skipped  int
	Failed   int
}

// RenderProjects 批量扫描并渲染目录下的项目配置文件。
func RenderProjects(dir, pattern string, recursive bool) error {
	files, err := FindProjectFiles(dir, pattern, recursive)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("在目录 %s 下没有找到匹配的项目文件", dir)
	}

	summary := BatchSummary{}
	for _, path := range files {
		p, err := LoadProject(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "clawcut: 跳过 %s（%v）\n", path, err)
			summary.Skipped++
			continue
		}
		if err := p.Validate(); err != nil {
			fmt.Fprintf(os.Stderr, "clawcut: 跳过 %s（%v）\n", path, err)
			summary.Skipped++
			continue
		}
		fmt.Printf("clawcut: 开始渲染 %s\n", path)
		if err := RenderProject(path); err != nil {
			fmt.Fprintf(os.Stderr, "clawcut: 渲染失败 %s（%v）\n", path, err)
			summary.Failed++
			continue
		}
		summary.Rendered++
	}

	fmt.Printf(
		"clawcut: 批量渲染完成，成功=%d 跳过=%d 失败=%d\n",
		summary.Rendered,
		summary.Skipped,
		summary.Failed,
	)
	if summary.Failed > 0 {
		return fmt.Errorf("批量渲染完成，但有 %d 个项目失败", summary.Failed)
	}
	if summary.Rendered == 0 {
		return fmt.Errorf("没有可成功渲染的有效项目")
	}
	return nil
}

// FindProjectFiles 按模式查找目录中的项目 JSON 文件。
func FindProjectFiles(dir, pattern string, recursive bool) ([]string, error) {
	if pattern == "" {
		pattern = "*.json"
	}
	var files []string
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != dir && !recursive {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".json" {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		match, err := filepath.Match(pattern, rel)
		if err != nil {
			return err
		}
		if !match {
			baseMatch, err := filepath.Match(pattern, filepath.Base(path))
			if err != nil {
				return err
			}
			if !baseMatch {
				return nil
			}
		}
		if strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}
