package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shuangli441-ux/openclwa-cut/internal/ffmpeg"
)

// RenderReport 汇总一次成片渲染的关键交付信息。
type RenderReport struct {
	Project         string               `json:"project"`
	Format          string               `json:"format"`
	OutputPath      string               `json:"outputPath"`
	ReportPath      string               `json:"reportPath"`
	Resolution      string               `json:"resolution"`
	FPS             int                  `json:"fps"`
	DurationSeconds float64              `json:"durationSeconds"`
	FileSizeBytes   int64                `json:"fileSizeBytes"`
	StartedAt       string               `json:"startedAt"`
	FinishedAt      string               `json:"finishedAt"`
	Subtitle        SubtitleRenderReport `json:"subtitle"`
	Audio           AudioRenderReport    `json:"audio"`
	Branding        BrandingRenderReport `json:"branding"`
	Cover           CoverRenderReport    `json:"cover"`
	Publish         PublishRenderReport  `json:"publish"`
}

// SubtitleRenderReport 记录字幕渲染方式与降级结果。
type SubtitleRenderReport struct {
	Requested   bool   `json:"requested"`
	Mode        string `json:"mode"`
	SidecarPath string `json:"sidecarPath,omitempty"`
}

// AudioRenderReport 记录音频混音、ducking 与人声增强结果。
type AudioRenderReport struct {
	Enabled               bool    `json:"enabled"`
	MusicPath             string  `json:"musicPath,omitempty"`
	HasVoice              bool    `json:"hasVoice"`
	DuckingRequested      bool    `json:"duckingRequested"`
	DuckingApplied        bool    `json:"duckingApplied"`
	VoiceEnhanceRequested bool    `json:"voiceEnhanceRequested"`
	VoiceEnhanceApplied   bool    `json:"voiceEnhanceApplied"`
	VoiceBoost            float64 `json:"voiceBoost,omitempty"`
}

// BrandingRenderReport 记录品牌水印叠加结果。
type BrandingRenderReport struct {
	WatermarkRequested bool   `json:"watermarkRequested"`
	WatermarkApplied   bool   `json:"watermarkApplied"`
	WatermarkPath      string `json:"watermarkPath,omitempty"`
	Position           string `json:"position,omitempty"`
	Width              int    `json:"width,omitempty"`
	OpacityApplied     bool   `json:"opacityApplied,omitempty"`
}

// CoverRenderReport 记录封面导出结果。
type CoverRenderReport struct {
	Enabled   bool    `json:"enabled"`
	Path      string  `json:"path,omitempty"`
	Timestamp float64 `json:"timestamp,omitempty"`
}

// PublishRenderReport 记录发布文案交付文件。
type PublishRenderReport struct {
	Generated bool     `json:"generated"`
	Path      string   `json:"path,omitempty"`
	Title     string   `json:"title,omitempty"`
	Hashtags  []string `json:"hashtags,omitempty"`
}

// BuildRenderReport 基于成片元数据和渲染过程信息构造渲染报告。
func BuildRenderReport(
	project *Project,
	outputPath string,
	subtitleMode string,
	subtitleSidecarPath string,
	musicPath string,
	audio ffmpeg.AudioMixResult,
	branding ffmpeg.OverlayResult,
	coverPath string,
	coverTimestamp float64,
	publishPath string,
	startedAt time.Time,
	finishedAt time.Time,
) (RenderReport, error) {
	duration, err := ffmpeg.ProbeDuration(outputPath)
	if err != nil {
		return RenderReport{}, err
	}
	info, err := os.Stat(outputPath)
	if err != nil {
		return RenderReport{}, err
	}
	report := RenderReport{
		Project:         project.Project,
		Format:          project.Format,
		OutputPath:      outputPath,
		ReportPath:      project.ResolveReportPath(),
		Resolution:      project.Resolution,
		FPS:             project.FPS,
		DurationSeconds: duration,
		FileSizeBytes:   info.Size(),
		StartedAt:       startedAt.Format(time.RFC3339),
		FinishedAt:      finishedAt.Format(time.RFC3339),
		Subtitle: SubtitleRenderReport{
			Requested: project.HasSubtitleItems(),
			Mode:      subtitleMode,
		},
		Audio: AudioRenderReport{
			Enabled:               strings.TrimSpace(musicPath) != "",
			MusicPath:             musicPath,
			HasVoice:              audio.HasVoice,
			DuckingRequested:      audio.DuckingRequested,
			DuckingApplied:        audio.DuckingApplied,
			VoiceEnhanceRequested: audio.VoiceEnhanceRequested,
			VoiceEnhanceApplied:   audio.VoiceEnhanceApplied,
			VoiceBoost:            audio.VoiceBoost,
		},
		Branding: BrandingRenderReport{
			WatermarkRequested: strings.TrimSpace(project.Branding.WatermarkPath) != "",
			WatermarkApplied:   branding.Applied,
			WatermarkPath:      project.Branding.WatermarkPath,
			Position:           branding.Position,
			Width:              branding.Width,
			OpacityApplied:     branding.OpacityApplied,
		},
		Cover: CoverRenderReport{
			Enabled:   project.CoverEnabled(),
			Path:      coverPath,
			Timestamp: coverTimestamp,
		},
		Publish: PublishRenderReport{
			Generated: strings.TrimSpace(publishPath) != "",
			Path:      publishPath,
			Title:     project.ResolvedPublishTitle(),
			Hashtags:  project.ResolvedPublishHashtags(),
		},
	}
	if subtitleSidecarPath != "" {
		report.Subtitle.SidecarPath = subtitleSidecarPath
	}
	return report, nil
}

// WriteRenderReport 将渲染报告写入标准 JSON 文件。
func WriteRenderReport(path string, report RenderReport) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0644)
}

// WritePublishCopy 将发布文案写入文本交付文件。
func WritePublishCopy(path string, content string) error {
	if strings.TrimSpace(path) == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}
