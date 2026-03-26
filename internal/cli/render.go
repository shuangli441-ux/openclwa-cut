package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shuangli441-ux/openclwa-cut/internal/ffmpeg"
)

// RenderProject 从项目 JSON 读取配置并执行整条渲染流水线。
func RenderProject(projectPath string) error {
	p, err := LoadProject(projectPath)
	if err != nil {
		return err
	}
	SubtitleFontNameFromFile(p)
	return RenderLoadedProject(p)
}

// RenderLoadedProject 在已加载的项目配置上执行渲染。
func RenderLoadedProject(p *Project) (err error) {
	startedAt := time.Now()
	applyRenderPlatformDefaults(p)
	SubtitleFontNameFromFile(p)
	if err := p.Validate(); err != nil {
		return err
	}
	clipItems := clipTimelineItems(p)
	if len(clipItems) == 0 {
		return fmt.Errorf("时间线为空，请至少添加一个 clip 片段")
	}

	profile, err := renderProfileFromProject(p)
	if err != nil {
		return err
	}

	outputDir := filepath.Dir(p.Output.Path)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}
	workDir, err := os.MkdirTemp(outputDir, ".clawcut-build-")
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			_ = os.RemoveAll(workDir)
			return
		}
		fmt.Fprintf(os.Stderr, "clawcut: 保留渲染中间文件，便于排查：%s\n", workDir)
	}()

	progress := NewRenderProgress(renderStepCount(p, len(clipItems)))

	segments := make([]string, 0, len(clipItems))
	for idx, item := range clipItems {
		progress.Step(fmt.Sprintf("处理片段 %d/%d", idx+1, len(clipItems)))
		asset, ok := p.AssetByID(item.Asset)
		if !ok {
			return fmt.Errorf("找不到素材：%s", item.Asset)
		}
		seg := filepath.Join(workDir, fmt.Sprintf("seg_%02d_%s_%.0f_%.0f.mp4", idx, item.Asset, item.Start, item.End))
		dur := fmt.Sprintf("%g", item.End-item.Start)
		switch asset.Type {
		case "image":
			if err := ffmpeg.RenderImageSegment(asset.Path, dur, seg, profile); err != nil {
				return err
			}
		default:
			start := fmt.Sprintf("%g", item.Start)
			if err := ffmpeg.RenderSegment(asset.Path, start, dur, seg, profile); err != nil {
				return err
			}
		}
		segments = append(segments, seg)
	}
	if len(segments) == 0 {
		return fmt.Errorf("没有生成任何片段，请检查 timeline 配置")
	}

	currentOutput := segments[0]
	if len(segments) > 1 {
		progress.Step("拼接片段")
		currentOutput = filepath.Join(workDir, "render_base.mp4")
		if err := ffmpeg.ConcatSegments(segments, currentOutput, profile); err != nil {
			return err
		}
	}

	subtitleMode, subtitleSidecarPath, subtitledOutput := applySubtitlesWithFallback(p, workDir, currentOutput, profile, progress)
	currentOutput = subtitledOutput

	overlayResult := ffmpeg.OverlayResult{}
	if strings.TrimSpace(p.Branding.WatermarkPath) != "" {
		progress.Step("叠加品牌水印")
		nextOutput := filepath.Join(workDir, "render_brand.mp4")
		overlayOpts := ffmpeg.OverlayOptions{
			Position:   p.Branding.WatermarkPosition,
			WidthRatio: p.Branding.WatermarkWidthRatio,
			Opacity:    p.Branding.WatermarkOpacity,
			MarginX:    p.Branding.MarginX,
			MarginY:    p.Branding.MarginY,
			Start:      p.Branding.Start,
			End:        p.Branding.End,
		}
		overlayResult, err = ffmpeg.ApplyWatermark(currentOutput, p.Branding.WatermarkPath, nextOutput, profile, overlayOpts)
		if err != nil {
			return err
		}
		currentOutput = nextOutput
	}

	audioMixResult := ffmpeg.AudioMixResult{}
	musicPathUsed := ""
	musicRequested := strings.TrimSpace(p.Music.Path) != "" || strings.TrimSpace(p.Music.Style) != ""
	if musicPath, musicErr := ResolveMusicPathForRender(p, currentOutput); musicErr != nil {
		return musicErr
	} else if musicPath != "" {
		progress.Step("混入背景音乐并自动压低 BGM")
		musicPathUsed = musicPath
		opts := ffmpeg.AudioMixOptions{
			Volume:           p.Music.Volume,
			Loop:             !p.Music.DisableLoop,
			FadeOutSeconds:   p.Music.FadeOutSeconds,
			Ducking:          !p.Music.DisableDucking,
			DuckingThreshold: p.Music.DuckingThreshold,
			DuckingRatio:     p.Music.DuckingRatio,
			DuckingAttackMs:  p.Music.DuckingAttackMs,
			DuckingReleaseMs: p.Music.DuckingReleaseMs,
			VoiceEnhance:     !p.Music.DisableVoiceEnhance,
			VoiceHighpassHz:  p.Music.VoiceHighpassHz,
			VoiceLowpassHz:   p.Music.VoiceLowpassHz,
			VoiceBoost:       p.Music.VoiceBoost,
		}
		nextOutput := filepath.Join(workDir, "render_music.mp4")
		audioMixResult, err = ffmpeg.MixBackgroundMusic(currentOutput, musicPath, nextOutput, profile, opts)
		if err != nil {
			return err
		}
		currentOutput = nextOutput
	} else if musicRequested {
		progress.Step("跳过背景音乐（未匹配到可用 BGM）")
	}

	progress.Step("输出成片")
	_ = os.Remove(p.Output.Path)
	if err := os.Rename(currentOutput, p.Output.Path); err != nil {
		return err
	}

	coverPath := ""
	coverTimestamp := 0.0
	if p.CoverEnabled() {
		progress.Step("导出封面")
		coverPath = p.ResolveCoverPath()
		coverTimestamp = p.ResolveCoverTimestamp()
		if strings.TrimSpace(p.Cover.Title) != "" {
			if coverErr := ffmpeg.ExportCoverFrameWithTitle(
				p.Output.Path,
				coverPath,
				coverTimestamp,
				p.Cover.Quality,
				p.Cover.Title,
				ffmpeg.CoverTextOptions{
					FontSize:     p.Cover.TitleFontSize,
					FontColor:    p.Cover.TitleColor,
					MarginBottom: p.Cover.TitleMarginBottom,
				},
			); coverErr != nil {
				fmt.Fprintf(os.Stderr, "clawcut: 封面标题生成失败，已降级为普通封面：%v\n", coverErr)
				if err := ffmpeg.ExportCoverFrame(p.Output.Path, coverPath, coverTimestamp, p.Cover.Quality); err != nil {
					return err
				}
			}
		} else if err := ffmpeg.ExportCoverFrame(p.Output.Path, coverPath, coverTimestamp, p.Cover.Quality); err != nil {
			return err
		}
	}

	publishPath := ""
	if p.HasPublishCopy() {
		progress.Step("生成发布文案")
		publishPath = p.ResolvePublishPath()
		if err := WritePublishCopy(publishPath, p.BuildPublishCopy(p.Output.Path, coverPath, subtitleSidecarPath)); err != nil {
			return err
		}
	}

	progress.Step("写入渲染报告")
	report, err := BuildRenderReport(
		p,
		p.Output.Path,
		subtitleMode,
		subtitleSidecarPath,
		musicPathUsed,
		audioMixResult,
		overlayResult,
		coverPath,
		coverTimestamp,
		publishPath,
		startedAt,
		time.Now(),
	)
	if err != nil {
		return err
	}
	if err := WriteRenderReport(p.ResolveReportPath(), report); err != nil {
		return err
	}

	return nil
}

func applyRenderPlatformDefaults(p *Project) {
	if p.Format == "douyin-vertical" || p.Output.Platform == "douyin" || strings.TrimSpace(p.Format) == "" {
		p.Format = "douyin-vertical"
		p.Output.Platform = "douyin"
		p.Resolution = "1080x1920"
		p.FPS = 30
	}
}

func renderProfileFromProject(p *Project) (ffmpeg.VideoProfile, error) {
	dims, err := p.Dimensions()
	if err != nil {
		return ffmpeg.VideoProfile{}, err
	}
	return ffmpeg.VideoProfile{
		Width:        dims.Width,
		Height:       dims.Height,
		FPS:          p.FPS,
		VideoCodec:   p.Output.VideoCodec,
		AudioCodec:   p.Output.AudioCodec,
		AudioBitrate: p.Output.AudioBitrate,
		Preset:       p.Output.Preset,
		CRF:          p.Output.CRF,
	}, nil
}

func clipTimelineItems(p *Project) []TimelineItem {
	items := make([]TimelineItem, 0, len(p.Timeline))
	for _, item := range p.Timeline {
		if item.Type == "clip" && item.Asset != "" {
			items = append(items, item)
		}
	}
	return items
}

func renderStepCount(p *Project, clipCount int) int {
	total := clipCount + 2
	if clipCount > 1 {
		total++
	}
	if p.HasSubtitleItems() && !p.Subtitles.Disabled {
		total++
	}
	if strings.TrimSpace(p.Branding.WatermarkPath) != "" {
		total++
	}
	if strings.TrimSpace(p.Music.Path) != "" || strings.TrimSpace(p.Music.Style) != "" {
		total++
	}
	if p.CoverEnabled() {
		total++
	}
	if p.HasPublishCopy() {
		total++
	}
	return total
}

func applySubtitlesWithFallback(
	p *Project,
	workDir string,
	currentOutput string,
	profile ffmpeg.VideoProfile,
	progress *RenderProgress,
) (string, string, string) {
	if !p.HasSubtitleItems() {
		return "none", "", currentOutput
	}
	if p.Subtitles.Disabled {
		return "disabled", "", currentOutput
	}

	progress.Step("生成字幕")
	subtitlePath, err := WriteASS(p, workDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "clawcut: 字幕文件生成失败，已跳过字幕：%v\n", err)
		return "bypass", "", currentOutput
	}

	subtitleSidecarPath := p.ResolveSubtitlePath()
	if err := copyFile(subtitlePath, subtitleSidecarPath); err != nil {
		fmt.Fprintf(os.Stderr, "clawcut: 字幕外挂文件写入失败，继续输出成片：%v\n", err)
		subtitleSidecarPath = ""
	}

	if nextOutput, ok := tryBurnSubtitles(p, currentOutput, subtitlePath, workDir, profile); ok {
		return "subtitles", subtitleSidecarPath, nextOutput
	}
	if nextOutput, ok := tryDrawtextSubtitles(p, currentOutput, workDir, profile); ok {
		return "drawtext", subtitleSidecarPath, nextOutput
	}
	if subtitleSidecarPath != "" {
		fmt.Fprintf(os.Stderr, "clawcut: 当前环境不支持硬字幕，已降级为外挂字幕：%s\n", subtitleSidecarPath)
		return "sidecar", subtitleSidecarPath, currentOutput
	}

	fmt.Fprintln(os.Stderr, "clawcut: 字幕已旁路输出，渲染继续，不影响成片生成")
	return "bypass", "", currentOutput
}

func tryBurnSubtitles(
	p *Project,
	currentOutput string,
	subtitlePath string,
	workDir string,
	profile ffmpeg.VideoProfile,
) (string, bool) {
	hasSubtitles, err := ffmpeg.HasFilter("subtitles")
	if err != nil {
		fmt.Fprintf(os.Stderr, "clawcut: 检查 subtitles 滤镜失败，尝试降级：%v\n", err)
		return "", false
	}
	if !hasSubtitles {
		return "", false
	}
	nextOutput := filepath.Join(workDir, "render_subtitled.mp4")
	if err := ffmpeg.BurnSubtitles(currentOutput, subtitlePath, nextOutput, profile, SubtitleFontsDir(p)); err != nil {
		fmt.Fprintf(os.Stderr, "clawcut: 硬字幕烧录失败，尝试降级：%v\n", err)
		return "", false
	}
	return nextOutput, true
}

func tryDrawtextSubtitles(
	p *Project,
	currentOutput string,
	workDir string,
	profile ffmpeg.VideoProfile,
) (string, bool) {
	hasDrawtext, err := ffmpeg.HasFilter("drawtext")
	if err != nil {
		fmt.Fprintf(os.Stderr, "clawcut: 检查 drawtext 滤镜失败，尝试降级：%v\n", err)
		return "", false
	}
	if !hasDrawtext {
		return "", false
	}
	filter, err := BuildDrawtextFilter(p)
	if err != nil {
		fmt.Fprintf(os.Stderr, "clawcut: 生成 drawtext 字幕失败，尝试降级：%v\n", err)
		return "", false
	}
	nextOutput := filepath.Join(workDir, "render_subtitled.mp4")
	if err := ffmpeg.RenderVideoFilter(currentOutput, filter, nextOutput, profile); err != nil {
		fmt.Fprintf(os.Stderr, "clawcut: drawtext 字幕烧录失败，尝试降级：%v\n", err)
		return "", false
	}
	return nextOutput, true
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
