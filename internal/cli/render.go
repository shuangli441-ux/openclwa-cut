package cli

import (
	"clawcut/internal/ffmpeg"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func RenderProject(projectPath string) (err error) {
	startedAt := time.Now()
	p, err := LoadProject(projectPath)
	if err != nil {
		return err
	}
	SubtitleFontNameFromFile(p)
	if err := p.Validate(); err != nil {
		return err
	}
	if len(p.Timeline) == 0 {
		return fmt.Errorf("timeline is empty")
	}
	dims, err := p.Dimensions()
	if err != nil {
		return err
	}
	profile := ffmpeg.VideoProfile{
		Width:        dims.Width,
		Height:       dims.Height,
		FPS:          p.FPS,
		VideoCodec:   p.Output.VideoCodec,
		AudioCodec:   p.Output.AudioCodec,
		AudioBitrate: p.Output.AudioBitrate,
		Preset:       p.Output.Preset,
		CRF:          p.Output.CRF,
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
		fmt.Fprintf(os.Stderr, "clawcut: kept render artifacts for debugging: %s\n", workDir)
	}()

	var segments []string
	for idx, item := range p.Timeline {
		if item.Type != "clip" || item.Asset == "" {
			continue
		}
		asset, ok := p.AssetByID(item.Asset)
		if !ok {
			return fmt.Errorf("asset not found: %s", item.Asset)
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
		return fmt.Errorf("no clip segments rendered")
	}
	currentOutput := segments[0]
	if len(segments) > 1 {
		currentOutput = filepath.Join(workDir, "render_base.mp4")
		if err := ffmpeg.ConcatSegments(segments, currentOutput, profile); err != nil {
			return err
		}
	}

	subtitleMode := "none"
	subtitleSidecarPath := ""
	audioMixResult := ffmpeg.AudioMixResult{}
	overlayResult := ffmpeg.OverlayResult{}
	coverPath := ""
	coverTimestamp := 0.0
	musicPathUsed := ""

	hasSubtitle := false
	for _, item := range p.Timeline {
		if item.Type == "subtitle" && item.Text != "" {
			hasSubtitle = true
			break
		}
	}
	if hasSubtitle && !p.Subtitles.Disabled {
		switch mode, modeErr := subtitleRenderMode(); {
		case modeErr != nil:
			return modeErr
		case mode == "subtitles":
			subtitleMode = "subtitles"
			subtitlePath, err := WriteASS(p, workDir)
			if err != nil {
				return err
			}
			nextOutput := filepath.Join(workDir, "render_subtitled.mp4")
			if err := ffmpeg.BurnSubtitles(currentOutput, subtitlePath, nextOutput, profile, SubtitleFontsDir(p)); err != nil {
				return err
			}
			currentOutput = nextOutput
		case mode == "drawtext":
			subtitleMode = "drawtext"
			filter, err := BuildDrawtextFilter(p)
			if err != nil {
				return err
			}
			nextOutput := filepath.Join(workDir, "render_subtitled.mp4")
			if err := ffmpeg.RenderVideoFilter(currentOutput, filter, nextOutput, profile); err != nil {
				return err
			}
			currentOutput = nextOutput
		default:
			subtitleMode = "sidecar"
			subtitlePath, err := WriteASS(p, workDir)
			if err != nil {
				return err
			}
			sidecarPath := strings.TrimSuffix(p.Output.Path, filepath.Ext(p.Output.Path)) + ".ass"
			if err := copyFile(subtitlePath, sidecarPath); err != nil {
				return err
			}
			subtitleSidecarPath = sidecarPath
			fmt.Fprintf(os.Stderr, "clawcut: ffmpeg subtitle filters unavailable, wrote sidecar subtitles: %s\n", sidecarPath)
		}
	} else if hasSubtitle {
		subtitleMode = "disabled"
	}

	if strings.TrimSpace(p.Branding.WatermarkPath) != "" {
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

	if musicPath, musicErr := p.ResolveMusicPath(); musicErr != nil {
		return musicErr
	} else if musicPath != "" {
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
	}

	_ = os.Remove(p.Output.Path)
	if err := os.Rename(currentOutput, p.Output.Path); err != nil {
		return err
	}

	if p.CoverEnabled() {
		coverPath = p.ResolveCoverPath()
		coverTimestamp = p.ResolveCoverTimestamp()
		if err := ffmpeg.ExportCoverFrame(p.Output.Path, coverPath, coverTimestamp, p.Cover.Quality); err != nil {
			return err
		}
	}

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

func subtitleRenderMode() (string, error) {
	hasSubtitles, err := ffmpeg.HasFilter("subtitles")
	if err != nil {
		return "", err
	}
	if hasSubtitles {
		return "subtitles", nil
	}
	hasDrawtext, err := ffmpeg.HasFilter("drawtext")
	if err != nil {
		return "", err
	}
	if hasDrawtext {
		return "drawtext", nil
	}
	return "sidecar", nil
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
