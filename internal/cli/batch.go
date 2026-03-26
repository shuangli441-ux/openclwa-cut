package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type BatchSummary struct {
	Rendered int
	Skipped  int
	Failed   int
}

func RenderProjects(dir, pattern string, recursive bool) error {
	files, err := FindProjectFiles(dir, pattern, recursive)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no project files matched in %s", dir)
	}

	summary := BatchSummary{}
	for _, path := range files {
		p, err := LoadProject(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "clawcut: skip %s (%v)\n", path, err)
			summary.Skipped++
			continue
		}
		if err := p.Validate(); err != nil {
			fmt.Fprintf(os.Stderr, "clawcut: skip %s (%v)\n", path, err)
			summary.Skipped++
			continue
		}
		fmt.Printf("clawcut: render %s\n", path)
		if err := RenderProject(path); err != nil {
			fmt.Fprintf(os.Stderr, "clawcut: failed %s (%v)\n", path, err)
			summary.Failed++
			continue
		}
		summary.Rendered++
	}

	fmt.Printf(
		"clawcut: summary rendered=%d skipped=%d failed=%d\n",
		summary.Rendered,
		summary.Skipped,
		summary.Failed,
	)
	if summary.Failed > 0 {
		return fmt.Errorf("batch render completed with %d failure(s)", summary.Failed)
	}
	if summary.Rendered == 0 {
		return fmt.Errorf("no valid projects were rendered")
	}
	return nil
}

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
