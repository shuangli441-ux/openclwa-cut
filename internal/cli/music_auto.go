package cli

import (
	"fmt"
	"os"

	"github.com/shuangli441-ux/openclwa-cut/internal/ffmpeg"
)

// PickMusicForStyle 按风格从音乐库里挑一首可用的背景音乐。
func PickMusicForStyle(libraryPath, style string) (string, error) {
	return PickMusicForVideo(libraryPath, style, 0, -30)
}

// PickMusicForVideo 按风格、时长和人声强度自动选择背景音乐。
func PickMusicForVideo(libraryPath, style string, targetDuration float64, voiceMeanVolume float64) (string, error) {
	lib, err := LoadMusicLibrary(libraryPath)
	if err != nil {
		return "", err
	}
	if track := lib.FindBest(style, targetDuration, voiceMeanVolume); track != nil {
		return track.Path, nil
	}
	if style != "" {
		return "", fmt.Errorf("音乐库里没有找到适合风格 %q 的背景音乐", style)
	}
	return "", fmt.Errorf("音乐库为空，无法自动匹配背景音乐")
}

// ResolveMusicPathForRender 在渲染前解析本次实际要使用的背景音乐路径。
func ResolveMusicPathForRender(project *Project, sourceVideo string) (string, error) {
	if project == nil {
		return "", nil
	}
	if project.Music.Path != "" {
		return project.Music.Path, nil
	}
	if project.Music.Style == "" {
		return "", nil
	}
	libraryPath := project.ResolveMusicLibraryPath()
	if _, err := os.Stat(libraryPath); err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "clawcut: 未找到音乐库 %s，已跳过自动匹配 BGM，成片继续生成\n", libraryPath)
			return "", nil
		}
		fmt.Fprintf(os.Stderr, "clawcut: 读取音乐库失败，已跳过自动匹配 BGM：%v\n", err)
		return "", nil
	}

	targetDuration, err := ffmpeg.ProbeDuration(sourceVideo)
	if err != nil {
		targetDuration = project.TotalDuration()
	}
	voiceMeanVolume := -30.0
	if hasAudio, audioErr := ffmpeg.HasAudio(sourceVideo); audioErr == nil && hasAudio {
		if meanVolume, volumeErr := ffmpeg.ProbeMeanVolume(sourceVideo); volumeErr == nil {
			voiceMeanVolume = meanVolume
		}
	}
	musicPath, pickErr := PickMusicForVideo(libraryPath, project.Music.Style, targetDuration, voiceMeanVolume)
	if pickErr != nil {
		fmt.Fprintf(os.Stderr, "clawcut: 自动匹配背景音乐失败，已跳过 BGM：%v\n", pickErr)
		return "", nil
	}
	return musicPath, nil
}
