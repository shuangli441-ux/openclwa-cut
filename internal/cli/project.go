package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Project struct {
	Project    string           `json:"project"`
	Format     string           `json:"format"`
	FPS        int              `json:"fps"`
	Resolution string           `json:"resolution"`
	Assets     []Asset          `json:"assets,omitempty"`
	Timeline   []TimelineItem   `json:"timeline,omitempty"`
	Subtitles  SubtitleSettings `json:"subtitles,omitempty"`
	Music      MusicSettings    `json:"music,omitempty"`
	Branding   BrandingSettings `json:"branding,omitempty"`
	Cover      CoverSettings    `json:"cover,omitempty"`
	Output     Output           `json:"output"`
}

type Asset struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Path string `json:"path"`
}

type TimelineItem struct {
	Type  string  `json:"type"`
	Asset string  `json:"asset,omitempty"`
	Start float64 `json:"start,omitempty"`
	End   float64 `json:"end,omitempty"`
	Text  string  `json:"text,omitempty"`
}

type SubtitleSettings struct {
	Disabled        bool   `json:"disabled,omitempty"`
	FontName        string `json:"fontName,omitempty"`
	FontFile        string `json:"fontFile,omitempty"`
	FontSize        int    `json:"fontSize,omitempty"`
	PrimaryColor    string `json:"primaryColor,omitempty"`
	OutlineColor    string `json:"outlineColor,omitempty"`
	BorderStyle     int    `json:"borderStyle,omitempty"`
	Outline         int    `json:"outline,omitempty"`
	Shadow          int    `json:"shadow,omitempty"`
	Alignment       int    `json:"alignment,omitempty"`
	MarginL         int    `json:"marginL,omitempty"`
	MarginR         int    `json:"marginR,omitempty"`
	MarginV         int    `json:"marginV,omitempty"`
	Bold            bool   `json:"bold,omitempty"`
	Italic          bool   `json:"italic,omitempty"`
	MaxCharsPerLine int    `json:"maxCharsPerLine,omitempty"`
}

type MusicSettings struct {
	Path                string  `json:"path,omitempty"`
	Library             string  `json:"library,omitempty"`
	Style               string  `json:"style,omitempty"`
	Volume              float64 `json:"volume,omitempty"`
	DisableLoop         bool    `json:"disableLoop,omitempty"`
	FadeOutSeconds      float64 `json:"fadeOutSeconds,omitempty"`
	DisableDucking      bool    `json:"disableDucking,omitempty"`
	DuckingThreshold    float64 `json:"duckingThreshold,omitempty"`
	DuckingRatio        float64 `json:"duckingRatio,omitempty"`
	DuckingAttackMs     int     `json:"duckingAttackMs,omitempty"`
	DuckingReleaseMs    int     `json:"duckingReleaseMs,omitempty"`
	DisableVoiceEnhance bool    `json:"disableVoiceEnhance,omitempty"`
	VoiceHighpassHz     int     `json:"voiceHighpassHz,omitempty"`
	VoiceLowpassHz      int     `json:"voiceLowpassHz,omitempty"`
	VoiceBoost          float64 `json:"voiceBoost,omitempty"`
}

type BrandingSettings struct {
	WatermarkPath       string  `json:"watermarkPath,omitempty"`
	WatermarkPosition   string  `json:"watermarkPosition,omitempty"`
	WatermarkWidthRatio float64 `json:"watermarkWidthRatio,omitempty"`
	WatermarkOpacity    float64 `json:"watermarkOpacity,omitempty"`
	MarginX             int     `json:"marginX,omitempty"`
	MarginY             int     `json:"marginY,omitempty"`
	Start               float64 `json:"start,omitempty"`
	End                 float64 `json:"end,omitempty"`
}

type CoverSettings struct {
	Enabled   bool    `json:"enabled,omitempty"`
	Path      string  `json:"path,omitempty"`
	Timestamp float64 `json:"timestamp,omitempty"`
	Quality   int     `json:"quality,omitempty"`
}

type Output struct {
	Path         string `json:"path"`
	Platform     string `json:"platform"`
	VideoCodec   string `json:"videoCodec,omitempty"`
	AudioCodec   string `json:"audioCodec,omitempty"`
	AudioBitrate string `json:"audioBitrate,omitempty"`
	Preset       string `json:"preset,omitempty"`
	CRF          int    `json:"crf,omitempty"`
	ReportPath   string `json:"reportPath,omitempty"`
}

type Dimensions struct {
	Width  int
	Height int
}

func InitProject(dir, name string) error {
	if err := os.MkdirAll(filepath.Join(dir, "assets"), 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(dir, "audio"), 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(dir, "subtitles"), 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(dir, "output"), 0755); err != nil {
		return err
	}
	p := Project{
		Project:    name,
		Format:     "douyin-vertical",
		FPS:        30,
		Resolution: "1080x1920",
		Subtitles: SubtitleSettings{
			FontName:        "PingFang SC",
			PrimaryColor:    "#FFFFFF",
			OutlineColor:    "#000000",
			Outline:         3,
			Shadow:          0,
			Alignment:       2,
			MarginV:         160,
			MaxCharsPerLine: 18,
		},
		Music: MusicSettings{
			Volume:           0.14,
			FadeOutSeconds:   1.2,
			DuckingThreshold: 0.035,
			DuckingRatio:     10,
			DuckingAttackMs:  20,
			DuckingReleaseMs: 350,
			VoiceHighpassHz:  120,
			VoiceLowpassHz:   9000,
			VoiceBoost:       1.0,
		},
		Output: Output{
			Path:         "output/final.mp4",
			Platform:     "douyin",
			VideoCodec:   "libx264",
			AudioCodec:   "aac",
			AudioBitrate: "160k",
			Preset:       "medium",
			CRF:          21,
		},
	}
	b, _ := json.MarshalIndent(p, "", "  ")
	return os.WriteFile(filepath.Join(dir, "project.json"), b, 0644)
}

func LoadProject(path string) (*Project, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var p Project
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	p.ApplyDefaults()
	p.ResolvePaths(filepath.Dir(path))
	return &p, nil
}

func (p *Project) ApplyDefaults() {
	if p.Format == "" {
		p.Format = "douyin-vertical"
	}
	switch p.Format {
	case "douyin-vertical":
		if p.Resolution == "" {
			p.Resolution = "1080x1920"
		}
		if p.FPS == 0 {
			p.FPS = 30
		}
		if p.Output.Platform == "" {
			p.Output.Platform = "douyin"
		}
	case "landscape-1080p":
		if p.Resolution == "" {
			p.Resolution = "1920x1080"
		}
		if p.FPS == 0 {
			p.FPS = 30
		}
		if p.Output.Platform == "" {
			p.Output.Platform = "general"
		}
	case "square-1080":
		if p.Resolution == "" {
			p.Resolution = "1080x1080"
		}
		if p.FPS == 0 {
			p.FPS = 30
		}
		if p.Output.Platform == "" {
			p.Output.Platform = "general"
		}
	default:
		if p.Resolution == "" {
			p.Resolution = "1080x1920"
		}
		if p.FPS == 0 {
			p.FPS = 30
		}
	}
	if p.Subtitles.FontName == "" {
		p.Subtitles.FontName = "PingFang SC"
	}
	if p.Subtitles.FontFile == "" {
		p.Subtitles.FontFile = DefaultSubtitleFontFile()
	}
	if p.Subtitles.FontSize == 0 {
		if dims, err := p.Dimensions(); err == nil {
			size := dims.Height / 32
			if size < 34 {
				size = 34
			}
			p.Subtitles.FontSize = size
		} else {
			p.Subtitles.FontSize = 60
		}
	}
	if p.Subtitles.PrimaryColor == "" {
		p.Subtitles.PrimaryColor = "#FFFFFF"
	}
	if p.Subtitles.OutlineColor == "" {
		p.Subtitles.OutlineColor = "#000000"
	}
	if p.Subtitles.BorderStyle == 0 {
		p.Subtitles.BorderStyle = 1
	}
	if p.Subtitles.Outline == 0 {
		p.Subtitles.Outline = 3
	}
	if p.Subtitles.Alignment == 0 {
		p.Subtitles.Alignment = 2
	}
	if p.Subtitles.MaxCharsPerLine == 0 {
		p.Subtitles.MaxCharsPerLine = 18
	}
	if p.Subtitles.MarginL == 0 {
		p.Subtitles.MarginL = 60
	}
	if p.Subtitles.MarginR == 0 {
		p.Subtitles.MarginR = 60
	}
	if p.Subtitles.MarginV == 0 {
		if dims, err := p.Dimensions(); err == nil {
			margin := dims.Height / 12
			if margin < 80 {
				margin = 80
			}
			p.Subtitles.MarginV = margin
		} else {
			p.Subtitles.MarginV = 160
		}
	}
	if p.Music.Volume == 0 {
		p.Music.Volume = 0.14
	}
	if p.Music.FadeOutSeconds == 0 {
		p.Music.FadeOutSeconds = 1.2
	}
	if p.Music.DuckingThreshold == 0 {
		p.Music.DuckingThreshold = 0.035
	}
	if p.Music.DuckingRatio == 0 {
		p.Music.DuckingRatio = 10
	}
	if p.Music.DuckingAttackMs == 0 {
		p.Music.DuckingAttackMs = 20
	}
	if p.Music.DuckingReleaseMs == 0 {
		p.Music.DuckingReleaseMs = 350
	}
	if p.Music.VoiceHighpassHz == 0 {
		p.Music.VoiceHighpassHz = 120
	}
	if p.Music.VoiceLowpassHz == 0 {
		p.Music.VoiceLowpassHz = 9000
	}
	if p.Music.VoiceBoost == 0 {
		p.Music.VoiceBoost = 1.0
	}
	if p.Branding.WatermarkPosition == "" {
		p.Branding.WatermarkPosition = "top-right"
	}
	if p.Branding.WatermarkWidthRatio == 0 {
		p.Branding.WatermarkWidthRatio = 0.18
	}
	if p.Branding.WatermarkOpacity == 0 {
		p.Branding.WatermarkOpacity = 0.92
	}
	if p.Branding.MarginX == 0 {
		if dims, err := p.Dimensions(); err == nil {
			margin := dims.Width / 30
			if margin < 24 {
				margin = 24
			}
			p.Branding.MarginX = margin
		} else {
			p.Branding.MarginX = 36
		}
	}
	if p.Branding.MarginY == 0 {
		if dims, err := p.Dimensions(); err == nil {
			margin := dims.Height / 30
			if margin < 24 {
				margin = 24
			}
			p.Branding.MarginY = margin
		} else {
			p.Branding.MarginY = 36
		}
	}
	if p.Cover.Quality == 0 {
		p.Cover.Quality = 2
	}
	if p.Output.VideoCodec == "" {
		p.Output.VideoCodec = "libx264"
	}
	if p.Output.AudioCodec == "" {
		p.Output.AudioCodec = "aac"
	}
	if p.Output.AudioBitrate == "" {
		p.Output.AudioBitrate = "160k"
	}
	if p.Output.Preset == "" {
		p.Output.Preset = "medium"
	}
	if p.Output.CRF == 0 {
		p.Output.CRF = 21
	}
}

func (p *Project) ResolvePaths(baseDir string) {
	for i := range p.Assets {
		p.Assets[i].Path = resolvePath(baseDir, p.Assets[i].Path)
	}
	p.Subtitles.FontFile = resolvePath(baseDir, p.Subtitles.FontFile)
	p.Music.Path = resolvePath(baseDir, p.Music.Path)
	p.Music.Library = resolvePath(baseDir, p.Music.Library)
	p.Branding.WatermarkPath = resolvePath(baseDir, p.Branding.WatermarkPath)
	p.Cover.Path = resolvePath(baseDir, p.Cover.Path)
	p.Output.Path = resolvePath(baseDir, p.Output.Path)
	p.Output.ReportPath = resolvePath(baseDir, p.Output.ReportPath)
}

func (p *Project) Dimensions() (Dimensions, error) {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(p.Resolution)), "x")
	if len(parts) != 2 {
		return Dimensions{}, fmt.Errorf("invalid resolution %q, want WIDTHxHEIGHT", p.Resolution)
	}
	width, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil || width <= 0 {
		return Dimensions{}, fmt.Errorf("invalid resolution width %q", parts[0])
	}
	height, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil || height <= 0 {
		return Dimensions{}, fmt.Errorf("invalid resolution height %q", parts[1])
	}
	return Dimensions{Width: width, Height: height}, nil
}

func (p *Project) HasSubtitleItems() bool {
	for _, item := range p.Timeline {
		if item.Type == "subtitle" && strings.TrimSpace(item.Text) != "" {
			return true
		}
	}
	return false
}

func (p *Project) TotalDuration() float64 {
	var total float64
	for _, item := range p.Timeline {
		if item.Type != "clip" {
			continue
		}
		total += item.End - item.Start
	}
	return total
}

func (p *Project) ResolveMusicPath() (string, error) {
	if strings.TrimSpace(p.Music.Path) != "" {
		return p.Music.Path, nil
	}
	if strings.TrimSpace(p.Music.Style) == "" {
		return "", nil
	}
	libraryPath := p.Music.Library
	if libraryPath == "" {
		libraryPath = "/Users/apple/.openclaw/workspace/tools/clawcut/config/music_library.json"
	}
	return PickMusicForStyle(libraryPath, p.Music.Style)
}

func (p *Project) CoverEnabled() bool {
	return p.Cover.Enabled || strings.TrimSpace(p.Cover.Path) != ""
}

func (p *Project) ResolveCoverPath() string {
	if !p.CoverEnabled() {
		return ""
	}
	if strings.TrimSpace(p.Cover.Path) != "" {
		return p.Cover.Path
	}
	base := strings.TrimSuffix(p.Output.Path, filepath.Ext(p.Output.Path))
	return base + "_cover.jpg"
}

func (p *Project) ResolveCoverTimestamp() float64 {
	if p.Cover.Timestamp > 0 {
		return p.Cover.Timestamp
	}
	total := p.TotalDuration()
	if total <= 0 {
		return 0
	}
	timestamp := total / 3
	if timestamp > 1.2 {
		timestamp = 1.2
	}
	if timestamp >= total {
		timestamp = total / 2
	}
	if timestamp < 0 {
		return 0
	}
	return timestamp
}

func (p *Project) ResolveReportPath() string {
	if strings.TrimSpace(p.Output.ReportPath) != "" {
		return p.Output.ReportPath
	}
	base := strings.TrimSuffix(p.Output.Path, filepath.Ext(p.Output.Path))
	return base + ".render.json"
}

func (p *Project) Validate() error {
	if strings.TrimSpace(p.Project) == "" {
		return errors.New("project name is required")
	}
	if strings.TrimSpace(p.Output.Path) == "" {
		return errors.New("output.path is required")
	}
	if p.FPS <= 0 {
		return fmt.Errorf("fps must be greater than 0, got %d", p.FPS)
	}
	if _, err := p.Dimensions(); err != nil {
		return err
	}
	if len(p.Assets) == 0 {
		return errors.New("at least one asset is required")
	}

	assetMap := make(map[string]Asset, len(p.Assets))
	for i, asset := range p.Assets {
		if strings.TrimSpace(asset.ID) == "" {
			return errors.New("asset.id is required")
		}
		if _, exists := assetMap[asset.ID]; exists {
			return fmt.Errorf("duplicate asset id %q", asset.ID)
		}
		if strings.TrimSpace(asset.Path) == "" {
			return fmt.Errorf("asset %q path is required", asset.ID)
		}
		info, err := os.Stat(asset.Path)
		if err != nil {
			return fmt.Errorf("asset %q path invalid: %w", asset.ID, err)
		}
		if info.IsDir() {
			return fmt.Errorf("asset %q path must be a file: %s", asset.ID, asset.Path)
		}
		asset.Type = normalizeAssetType(asset)
		switch asset.Type {
		case "video", "image":
		default:
			return fmt.Errorf("asset %q has unsupported type %q", asset.ID, asset.Type)
		}
		p.Assets[i].Type = asset.Type
		assetMap[asset.ID] = asset
	}

	clipCount := 0
	for i, item := range p.Timeline {
		switch item.Type {
		case "clip":
			clipCount++
			if item.Asset == "" {
				return fmt.Errorf("timeline[%d] clip is missing asset", i)
			}
			if _, ok := assetMap[item.Asset]; !ok {
				return fmt.Errorf("timeline[%d] references unknown asset %q", i, item.Asset)
			}
			if item.End <= item.Start {
				return fmt.Errorf("timeline[%d] clip end must be greater than start", i)
			}
		case "subtitle":
			if strings.TrimSpace(item.Text) == "" {
				return fmt.Errorf("timeline[%d] subtitle text is empty", i)
			}
			if item.End <= item.Start {
				return fmt.Errorf("timeline[%d] subtitle end must be greater than start", i)
			}
		default:
			return fmt.Errorf("timeline[%d] unsupported type %q", i, item.Type)
		}
	}
	if clipCount == 0 {
		return errors.New("timeline must include at least one clip item")
	}

	totalDuration := p.TotalDuration()
	for i, item := range p.Timeline {
		if item.Type != "subtitle" {
			continue
		}
		if item.End > totalDuration+0.05 {
			return fmt.Errorf("timeline[%d] subtitle exceeds total render duration %.2fs", i, totalDuration)
		}
	}

	if p.Music.Volume < 0 {
		return fmt.Errorf("music.volume must be >= 0, got %.2f", p.Music.Volume)
	}
	if p.Music.DuckingThreshold < 0 {
		return fmt.Errorf("music.duckingThreshold must be >= 0, got %.3f", p.Music.DuckingThreshold)
	}
	if p.Music.DuckingRatio < 0 {
		return fmt.Errorf("music.duckingRatio must be >= 0, got %.2f", p.Music.DuckingRatio)
	}
	if p.Music.DuckingAttackMs < 0 {
		return fmt.Errorf("music.duckingAttackMs must be >= 0, got %d", p.Music.DuckingAttackMs)
	}
	if p.Music.DuckingReleaseMs < 0 {
		return fmt.Errorf("music.duckingReleaseMs must be >= 0, got %d", p.Music.DuckingReleaseMs)
	}
	if p.Music.VoiceHighpassHz < 0 {
		return fmt.Errorf("music.voiceHighpassHz must be >= 0, got %d", p.Music.VoiceHighpassHz)
	}
	if p.Music.VoiceLowpassHz < 0 {
		return fmt.Errorf("music.voiceLowpassHz must be >= 0, got %d", p.Music.VoiceLowpassHz)
	}
	if p.Music.VoiceBoost <= 0 {
		return fmt.Errorf("music.voiceBoost must be > 0, got %.2f", p.Music.VoiceBoost)
	}
	musicPath, err := p.ResolveMusicPath()
	if err != nil {
		return err
	}
	if musicPath != "" {
		if info, err := os.Stat(musicPath); err != nil {
			return fmt.Errorf("music path invalid: %w", err)
		} else if info.IsDir() {
			return fmt.Errorf("music path must be a file: %s", musicPath)
		}
	}
	if p.Subtitles.FontFile != "" {
		if info, err := os.Stat(p.Subtitles.FontFile); err != nil {
			return fmt.Errorf("subtitles.fontFile invalid: %w", err)
		} else if info.IsDir() {
			return fmt.Errorf("subtitles.fontFile must be a file: %s", p.Subtitles.FontFile)
		}
	}
	if strings.TrimSpace(p.Branding.WatermarkPath) != "" {
		if info, err := os.Stat(p.Branding.WatermarkPath); err != nil {
			return fmt.Errorf("branding.watermarkPath invalid: %w", err)
		} else if info.IsDir() {
			return fmt.Errorf("branding.watermarkPath must be a file: %s", p.Branding.WatermarkPath)
		}
		switch p.Branding.WatermarkPosition {
		case "top-left", "top-center", "top-right", "center", "bottom-left", "bottom-center", "bottom-right":
		default:
			return fmt.Errorf("branding.watermarkPosition unsupported: %q", p.Branding.WatermarkPosition)
		}
		if p.Branding.WatermarkWidthRatio <= 0 {
			return fmt.Errorf("branding.watermarkWidthRatio must be > 0, got %.3f", p.Branding.WatermarkWidthRatio)
		}
		if p.Branding.WatermarkOpacity <= 0 || p.Branding.WatermarkOpacity > 1 {
			return fmt.Errorf("branding.watermarkOpacity must be in (0,1], got %.3f", p.Branding.WatermarkOpacity)
		}
		if p.Branding.MarginX < 0 {
			return fmt.Errorf("branding.marginX must be >= 0, got %d", p.Branding.MarginX)
		}
		if p.Branding.MarginY < 0 {
			return fmt.Errorf("branding.marginY must be >= 0, got %d", p.Branding.MarginY)
		}
		if p.Branding.Start < 0 {
			return fmt.Errorf("branding.start must be >= 0, got %.2f", p.Branding.Start)
		}
		if p.Branding.End < 0 {
			return fmt.Errorf("branding.end must be >= 0, got %.2f", p.Branding.End)
		}
		if p.Branding.End > 0 && p.Branding.End <= p.Branding.Start {
			return errors.New("branding.end must be greater than branding.start")
		}
	}
	if p.CoverEnabled() {
		if p.Cover.Timestamp < 0 {
			return fmt.Errorf("cover.timestamp must be >= 0, got %.2f", p.Cover.Timestamp)
		}
		if p.Cover.Quality <= 0 || p.Cover.Quality > 31 {
			return fmt.Errorf("cover.quality must be in [1,31], got %d", p.Cover.Quality)
		}
	}
	return nil
}

func (p *Project) AssetByID(id string) (Asset, bool) {
	for _, asset := range p.Assets {
		if asset.ID == id {
			return asset, true
		}
	}
	return Asset{}, false
}

func resolvePath(baseDir, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return value
	}
	if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
		return value
	}
	if filepath.IsAbs(value) {
		return value
	}
	return filepath.Clean(filepath.Join(baseDir, value))
}

func normalizeAssetType(asset Asset) string {
	assetType := strings.ToLower(strings.TrimSpace(asset.Type))
	if assetType != "" {
		return assetType
	}
	switch strings.ToLower(filepath.Ext(asset.Path)) {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif":
		return "image"
	default:
		return "video"
	}
}
