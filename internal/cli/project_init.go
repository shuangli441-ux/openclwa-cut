package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/shuangli441-ux/openclwa-cut/internal/ffmpeg"
)

// InitProjectOptions 描述从单个视频快速初始化项目时的可选参数。
type InitProjectOptions struct {
	InputVideo string
	OutputPath string
	MusicPath  string
	MusicStyle string
}

// InitProjectWithOptions 根据输入视频生成一个可直接渲染的最小项目。
func InitProjectWithOptions(dir, name string, opts InitProjectOptions) error {
	if err := ensureProjectDirectories(dir); err != nil {
		return err
	}

	project := Project{
		Project: name,
		Output: Output{
			Path:     "output/video/final.mp4",
			Platform: "douyin",
		},
		Cover: CoverSettings{
			Enabled: true,
		},
	}

	if opts.OutputPath != "" {
		project.Output.Path = opts.OutputPath
	}
	if opts.MusicPath != "" {
		project.Music.Path = normalizeProjectPathValue(dir, opts.MusicPath)
	}
	if opts.MusicStyle != "" {
		project.Music.Style = opts.MusicStyle
	}
	if opts.InputVideo != "" {
		inputPath, err := filepath.Abs(opts.InputVideo)
		if err != nil {
			return fmt.Errorf("解析输入视频路径失败：%w", err)
		}
		info, err := os.Stat(inputPath)
		if err != nil {
			return fmt.Errorf("输入视频不存在：%w", err)
		}
		if info.IsDir() {
			return fmt.Errorf("输入视频不能是目录：%s", inputPath)
		}
		duration, err := ffmpeg.ProbeDuration(inputPath)
		if err != nil {
			return fmt.Errorf("读取输入视频时长失败：%w", err)
		}
		project.Assets = []Asset{
			{
				ID:   "main",
				Type: "video",
				Path: normalizeProjectPathValue(dir, inputPath),
			},
		}
		project.Timeline = []TimelineItem{
			{
				Type:  "clip",
				Asset: "main",
				Start: 0,
				End:   duration,
			},
		}
		project.Cover.Title = name
	}

	project.ApplyDefaults()
	return writeProjectJSON(filepath.Join(dir, "project.json"), project)
}

// normalizeProjectPathValue 尽量把项目内引用写成相对路径，便于跨机器迁移。
func normalizeProjectPathValue(projectDir, value string) string {
	absProjectDir, err := filepath.Abs(projectDir)
	if err != nil {
		return value
	}
	absValue, err := filepath.Abs(value)
	if err != nil {
		return value
	}
	if rel, err := filepath.Rel(absProjectDir, absValue); err == nil && rel != "" && rel != "." && rel != absValue && rel[0] != '.' {
		return filepath.ToSlash(rel)
	}
	return absValue
}

func ensureProjectDirectories(dir string) error {
	for _, subDir := range []string{
		"assets",
		"audio",
		"subtitles",
		filepath.Join("output", "video"),
		filepath.Join("output", "cover"),
		filepath.Join("output", "subtitles"),
		filepath.Join("output", "report"),
	} {
		if err := os.MkdirAll(filepath.Join(dir, subDir), 0755); err != nil {
			return err
		}
	}
	return nil
}

func writeProjectJSON(path string, project Project) error {
	data, err := json.MarshalIndent(project, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0644)
}
