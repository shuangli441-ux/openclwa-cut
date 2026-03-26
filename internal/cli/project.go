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

// Project 描述一个可渲染的短视频项目。
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
	AIEdit     AIEditSettings   `json:"aiEdit,omitempty"`
	Publish    PublishSettings  `json:"publish,omitempty"`
	Output     Output           `json:"output"`
}

// Asset 表示项目里可引用的一份素材。
type Asset struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Path string `json:"path"`
}

// TimelineItem 表示时间线中的一个片段或字幕项。
type TimelineItem struct {
	Type  string  `json:"type"`
	Asset string  `json:"asset,omitempty"`
	Start float64 `json:"start,omitempty"`
	End   float64 `json:"end,omitempty"`
	Text  string  `json:"text,omitempty"`
}

// SubtitleSettings 定义字幕的样式和排版参数。
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

// MusicSettings 定义背景音乐的选择与混音策略。
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

// BrandingSettings 定义水印叠加配置。
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

// CoverSettings 定义封面导出及标题样式。
type CoverSettings struct {
	Enabled           bool    `json:"enabled,omitempty"`
	Path              string  `json:"path,omitempty"`
	Timestamp         float64 `json:"timestamp,omitempty"`
	Quality           int     `json:"quality,omitempty"`
	Title             string  `json:"title,omitempty"`
	TitleFontSize     int     `json:"titleFontSize,omitempty"`
	TitleColor        string  `json:"titleColor,omitempty"`
	TitleMarginBottom int     `json:"titleMarginBottom,omitempty"`
}

// AIEditSettings 定义智能剪辑模块的自动选段与节奏参数。
type AIEditSettings struct {
	Enabled            bool     `json:"enabled,omitempty"`
	Mode               string   `json:"mode,omitempty"`
	TemplateKind       string   `json:"templateKind,omitempty"`
	ScriptLines        []string `json:"scriptLines,omitempty"`
	MaxDurationSeconds float64  `json:"maxDurationSeconds,omitempty"`
	HookSeconds        float64  `json:"hookSeconds,omitempty"`
	CTASeconds         float64  `json:"ctaSeconds,omitempty"`
}

// PublishSettings 定义可直接复制到发布平台的标题、话题和说明文案。
type PublishSettings struct {
	Title       string   `json:"title,omitempty"`
	Hashtags    []string `json:"hashtags,omitempty"`
	Description string   `json:"description,omitempty"`
}

// Output 定义成片编码参数及输出位置。
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

// Dimensions 表示一个渲染画布尺寸。
type Dimensions struct {
	Width  int
	Height int
}

// InitProject 初始化一个空项目目录，并写入默认项目配置。
func InitProject(dir, name string) error {
	if err := ensureProjectDirectories(dir); err != nil {
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
		Cover: CoverSettings{
			Enabled: true,
			Title:   name,
		},
		AIEdit: AIEditSettings{
			Enabled:            true,
			Mode:               "smart",
			MaxDurationSeconds: 35,
			HookSeconds:        3,
			CTASeconds:         4,
		},
		Publish: PublishSettings{
			Title: name,
		},
		Output: Output{
			Path:         "output/video/final.mp4",
			Platform:     "douyin",
			VideoCodec:   "libx264",
			AudioCodec:   "aac",
			AudioBitrate: "160k",
			Preset:       "medium",
			CRF:          21,
		},
	}
	return writeProjectJSON(filepath.Join(dir, "project.json"), p)
}

// LoadProject 读取项目 JSON，并补齐默认值与路径。
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

// ApplyDefaults 为项目补齐生产环境默认参数。
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
	if p.Cover.TitleColor == "" {
		p.Cover.TitleColor = "#FFFFFF"
	}
	if p.Cover.TitleFontSize == 0 {
		if dims, err := p.Dimensions(); err == nil {
			size := dims.Height / 18
			if size < 64 {
				size = 64
			}
			p.Cover.TitleFontSize = size
		} else {
			p.Cover.TitleFontSize = 88
		}
	}
	if p.Cover.TitleMarginBottom == 0 {
		if dims, err := p.Dimensions(); err == nil {
			margin := dims.Height / 7
			if margin < 180 {
				margin = 180
			}
			p.Cover.TitleMarginBottom = margin
		} else {
			p.Cover.TitleMarginBottom = 240
		}
	}
	if !p.AIEdit.Enabled && p.AIEdit.Mode == "" && p.AIEdit.MaxDurationSeconds == 0 && p.AIEdit.HookSeconds == 0 && p.AIEdit.CTASeconds == 0 {
		p.AIEdit.Enabled = true
	}
	if p.AIEdit.Mode == "" {
		p.AIEdit.Mode = "smart"
	}
	if p.AIEdit.MaxDurationSeconds == 0 {
		p.AIEdit.MaxDurationSeconds = 35
	}
	if p.AIEdit.HookSeconds == 0 {
		p.AIEdit.HookSeconds = 3
	}
	if p.AIEdit.CTASeconds == 0 {
		p.AIEdit.CTASeconds = 4
	}
	if strings.TrimSpace(p.Publish.Title) == "" {
		if strings.TrimSpace(p.Cover.Title) != "" {
			p.Publish.Title = strings.TrimSpace(p.Cover.Title)
		} else {
			p.Publish.Title = strings.TrimSpace(p.Project)
		}
	}
	for i := range p.Publish.Hashtags {
		p.Publish.Hashtags[i] = strings.TrimSpace(p.Publish.Hashtags[i])
	}
	p.Publish.Description = strings.TrimSpace(p.Publish.Description)
	if strings.TrimSpace(p.Output.Path) == "" {
		p.Output.Path = "output/video/final.mp4"
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

// ResolvePaths 将项目中的相对路径解析为绝对路径。
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

// Dimensions 解析配置中的分辨率字符串。
func (p *Project) Dimensions() (Dimensions, error) {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(p.Resolution)), "x")
	if len(parts) != 2 {
		return Dimensions{}, fmt.Errorf("分辨率格式不正确：%q，请使用 WIDTHxHEIGHT，例如 1080x1920", p.Resolution)
	}
	width, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil || width <= 0 {
		return Dimensions{}, fmt.Errorf("分辨率宽度不正确：%q", parts[0])
	}
	height, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil || height <= 0 {
		return Dimensions{}, fmt.Errorf("分辨率高度不正确：%q", parts[1])
	}
	return Dimensions{Width: width, Height: height}, nil
}

// HasSubtitleItems 判断时间线里是否包含字幕项。
func (p *Project) HasSubtitleItems() bool {
	for _, item := range p.Timeline {
		if item.Type == "subtitle" && strings.TrimSpace(item.Text) != "" {
			return true
		}
	}
	return false
}

// TotalDuration 计算所有 clip 片段拼接后的总时长。
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

// ResolveMusicPath 根据显式路径或 style 解析最终要使用的 BGM。
func (p *Project) ResolveMusicPath() (string, error) {
	if strings.TrimSpace(p.Music.Path) != "" {
		return p.Music.Path, nil
	}
	if strings.TrimSpace(p.Music.Style) == "" {
		return "", nil
	}
	libraryPath := p.ResolveMusicLibraryPath()
	if _, err := os.Stat(libraryPath); err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("读取音乐库失败：%w", err)
	}
	return PickMusicForStyle(libraryPath, p.Music.Style)
}

// ResolveMusicLibraryPath 返回音乐库文件路径。
func (p *Project) ResolveMusicLibraryPath() string {
	if strings.TrimSpace(p.Music.Library) != "" {
		return p.Music.Library
	}
	return DefaultMusicLibraryPath()
}

// CoverEnabled 判断当前项目是否需要导出封面。
func (p *Project) CoverEnabled() bool {
	return p.Cover.Enabled || strings.TrimSpace(p.Cover.Path) != ""
}

// ResolveCoverPath 返回封面图的标准输出路径。
func (p *Project) ResolveCoverPath() string {
	if !p.CoverEnabled() {
		return ""
	}
	if strings.TrimSpace(p.Cover.Path) != "" {
		return p.Cover.Path
	}
	return p.resolveOutputArtifactPath("cover", "_cover.jpg")
}

// ResolveCoverTimestamp 计算默认封面截帧时间。
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

// ResolveReportPath 返回渲染报告的标准输出路径。
func (p *Project) ResolveReportPath() string {
	if strings.TrimSpace(p.Output.ReportPath) != "" {
		return p.Output.ReportPath
	}
	return p.resolveOutputArtifactPath("report", ".render.json")
}

// ResolveSubtitlePath 返回字幕侧输出文件路径。
func (p *Project) ResolveSubtitlePath() string {
	return p.resolveOutputArtifactPath("subtitles", ".ass")
}

// HasPublishCopy 判断当前项目是否需要生成发布文案交付文件。
func (p *Project) HasPublishCopy() bool {
	return p.Output.Platform == "douyin" ||
		strings.TrimSpace(p.Publish.Title) != "" ||
		len(p.ResolvedPublishHashtags()) > 0 ||
		strings.TrimSpace(p.Publish.Description) != ""
}

// ResolvePublishPath 返回发布文案文件的标准输出路径。
func (p *Project) ResolvePublishPath() string {
	if !p.HasPublishCopy() {
		return ""
	}
	return p.resolveOutputArtifactPath("report", ".publish.txt")
}

// ResolvedPublishTitle 返回最终用于发布文案的标题。
func (p *Project) ResolvedPublishTitle() string {
	if title := strings.TrimSpace(p.Publish.Title); title != "" {
		return title
	}
	if title := strings.TrimSpace(p.Cover.Title); title != "" {
		return title
	}
	return strings.TrimSpace(p.Project)
}

// ResolvedPublishHashtags 返回整理后的话题标签列表。
func (p *Project) ResolvedPublishHashtags() []string {
	if len(p.Publish.Hashtags) == 0 {
		return nil
	}
	result := make([]string, 0, len(p.Publish.Hashtags))
	seen := make(map[string]struct{}, len(p.Publish.Hashtags))
	for _, item := range p.Publish.Hashtags {
		tag := strings.TrimSpace(item)
		if tag == "" {
			continue
		}
		if !strings.HasPrefix(tag, "#") {
			tag = "#" + tag
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		result = append(result, tag)
	}
	return result
}

// BuildPublishCopy 生成可直接复制到平台后台的发布文案。
func (p *Project) BuildPublishCopy(videoPath, coverPath, subtitlePath string) string {
	lines := []string{
		"标题：",
		p.ResolvedPublishTitle(),
		"",
	}
	if hashtags := p.ResolvedPublishHashtags(); len(hashtags) > 0 {
		lines = append(lines, "话题：", strings.Join(hashtags, " "), "")
	}
	description := strings.TrimSpace(p.Publish.Description)
	if description != "" {
		lines = append(lines, "文案：", description, "")
	}
	lines = append(lines, "交付物：", "成片："+videoPath)
	if strings.TrimSpace(coverPath) != "" {
		lines = append(lines, "封面："+coverPath)
	}
	if strings.TrimSpace(subtitlePath) != "" {
		lines = append(lines, "字幕："+subtitlePath)
	}
	return strings.Join(lines, "\n") + "\n"
}

// DefaultConfigDir 返回 clawcut 默认配置目录。
func DefaultConfigDir() string {
	if dir, err := os.UserConfigDir(); err == nil && strings.TrimSpace(dir) != "" {
		return filepath.Join(dir, "clawcut")
	}
	if home, err := os.UserHomeDir(); err == nil && strings.TrimSpace(home) != "" {
		return filepath.Join(home, ".clawcut")
	}
	return filepath.Join(".", ".clawcut")
}

// DefaultMusicLibraryPath 返回默认音乐库索引文件路径。
func DefaultMusicLibraryPath() string {
	return filepath.Join(DefaultConfigDir(), "music_library.json")
}

func (p *Project) resolveOutputArtifactPath(kind string, suffix string) string {
	baseName := strings.TrimSuffix(filepath.Base(p.Output.Path), filepath.Ext(p.Output.Path))
	if baseName == "" {
		baseName = "final"
	}
	root := filepath.Dir(p.Output.Path)
	if filepath.Base(root) == "video" {
		root = filepath.Dir(root)
	}
	if kind == "video" {
		return filepath.Join(root, "video", baseName+suffix)
	}
	return filepath.Join(root, kind, baseName+suffix)
}

// Validate 在真正渲染前尽可能提前发现配置错误。
func (p *Project) Validate() error {
	if err := p.PrepareAIEditTimeline(); err != nil {
		return err
	}
	if strings.TrimSpace(p.Project) == "" {
		return errors.New("项目名不能为空，请填写 project")
	}
	if strings.TrimSpace(p.Output.Path) == "" {
		return errors.New("输出路径不能为空，请填写 output.path")
	}
	if p.FPS <= 0 {
		return fmt.Errorf("帧率必须大于 0，当前是 %d", p.FPS)
	}
	if _, err := p.Dimensions(); err != nil {
		return err
	}
	if len(p.Assets) == 0 {
		return errors.New("至少要提供一个素材文件")
	}
	if ext := strings.ToLower(filepath.Ext(p.Output.Path)); ext != ".mp4" {
		return fmt.Errorf("当前只支持输出 mp4 文件，请检查 output.path：%s", p.Output.Path)
	}

	assetMap := make(map[string]Asset, len(p.Assets))
	for i, asset := range p.Assets {
		if strings.TrimSpace(asset.ID) == "" {
			return errors.New("素材缺少 asset.id")
		}
		if _, exists := assetMap[asset.ID]; exists {
			return fmt.Errorf("素材 ID 重复：%q", asset.ID)
		}
		if strings.TrimSpace(asset.Path) == "" {
			return fmt.Errorf("素材 %q 缺少 path", asset.ID)
		}
		info, err := os.Stat(asset.Path)
		if err != nil {
			return fmt.Errorf("素材 %q 不存在或无法读取：%w", asset.ID, err)
		}
		if info.IsDir() {
			return fmt.Errorf("素材 %q 不能指向目录：%s", asset.ID, asset.Path)
		}
		asset.Type = normalizeAssetType(asset)
		switch asset.Type {
		case "video", "image":
		default:
			return fmt.Errorf("素材 %q 类型不支持：%q，只支持 video 或 image", asset.ID, asset.Type)
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
				return fmt.Errorf("时间线第 %d 项 clip 缺少 asset", i)
			}
			if _, ok := assetMap[item.Asset]; !ok {
				return fmt.Errorf("时间线第 %d 项引用了不存在的素材 %q", i, item.Asset)
			}
			if item.End <= item.Start {
				return fmt.Errorf("时间线第 %d 项 clip 的 end 必须大于 start", i)
			}
		case "subtitle":
			if strings.TrimSpace(item.Text) == "" {
				return fmt.Errorf("时间线第 %d 项字幕内容为空", i)
			}
			if item.End <= item.Start {
				return fmt.Errorf("时间线第 %d 项字幕的 end 必须大于 start", i)
			}
		default:
			return fmt.Errorf("时间线第 %d 项类型不支持：%q", i, item.Type)
		}
	}
	if clipCount == 0 {
		return errors.New("时间线里至少要有一个 clip 片段")
	}

	totalDuration := p.TotalDuration()
	for i, item := range p.Timeline {
		if item.Type != "subtitle" {
			continue
		}
		if item.End > totalDuration+0.05 {
			return fmt.Errorf("时间线第 %d 项字幕超出了成片总时长 %.2f 秒", i, totalDuration)
		}
	}

	if p.Music.Volume < 0 {
		return fmt.Errorf("music.volume 不能小于 0，当前为 %.2f", p.Music.Volume)
	}
	if p.Music.DuckingThreshold < 0 {
		return fmt.Errorf("music.duckingThreshold 不能小于 0，当前为 %.3f", p.Music.DuckingThreshold)
	}
	if p.Music.DuckingRatio < 0 {
		return fmt.Errorf("music.duckingRatio 不能小于 0，当前为 %.2f", p.Music.DuckingRatio)
	}
	if p.Music.DuckingAttackMs < 0 {
		return fmt.Errorf("music.duckingAttackMs 不能小于 0，当前为 %d", p.Music.DuckingAttackMs)
	}
	if p.Music.DuckingReleaseMs < 0 {
		return fmt.Errorf("music.duckingReleaseMs 不能小于 0，当前为 %d", p.Music.DuckingReleaseMs)
	}
	if p.Music.VoiceHighpassHz < 0 {
		return fmt.Errorf("music.voiceHighpassHz 不能小于 0，当前为 %d", p.Music.VoiceHighpassHz)
	}
	if p.Music.VoiceLowpassHz < 0 {
		return fmt.Errorf("music.voiceLowpassHz 不能小于 0，当前为 %d", p.Music.VoiceLowpassHz)
	}
	if p.Music.VoiceBoost <= 0 {
		return fmt.Errorf("music.voiceBoost 必须大于 0，当前为 %.2f", p.Music.VoiceBoost)
	}
	if p.AIEdit.MaxDurationSeconds < 0 {
		return fmt.Errorf("aiEdit.maxDurationSeconds 不能小于 0，当前为 %.2f", p.AIEdit.MaxDurationSeconds)
	}
	if p.AIEdit.HookSeconds < 0 {
		return fmt.Errorf("aiEdit.hookSeconds 不能小于 0，当前为 %.2f", p.AIEdit.HookSeconds)
	}
	if p.AIEdit.CTASeconds < 0 {
		return fmt.Errorf("aiEdit.ctaSeconds 不能小于 0，当前为 %.2f", p.AIEdit.CTASeconds)
	}
	if mode := strings.TrimSpace(strings.ToLower(p.AIEdit.Mode)); mode != "" && mode != "smart" && mode != "basic" {
		return fmt.Errorf("aiEdit.mode 仅支持 smart 或 basic，当前为 %q", p.AIEdit.Mode)
	}
	if p.AIEdit.TemplateKind != "" && normalizeTemplateKind(p.AIEdit.TemplateKind) == "" {
		return fmt.Errorf("aiEdit.templateKind 仅支持 %s、%s 或 %s，当前为 %q", TemplateDouyinQA, TemplateDouyinGoods, TemplateDouyinAds, p.AIEdit.TemplateKind)
	}
	musicPath, err := p.ResolveMusicPath()
	if err != nil {
		return fmt.Errorf("背景音乐配置有问题：%w", err)
	}
	if musicPath != "" {
		if info, err := os.Stat(musicPath); err != nil {
			return fmt.Errorf("背景音乐文件不存在或无法读取：%w", err)
		} else if info.IsDir() {
			return fmt.Errorf("背景音乐路径不能是目录：%s", musicPath)
		}
	}
	if p.Subtitles.FontFile != "" {
		if info, err := os.Stat(p.Subtitles.FontFile); err != nil {
			return fmt.Errorf("字幕字体文件不存在：%w", err)
		} else if info.IsDir() {
			return fmt.Errorf("字幕字体路径不能是目录：%s", p.Subtitles.FontFile)
		}
	}
	if strings.TrimSpace(p.Branding.WatermarkPath) != "" {
		if info, err := os.Stat(p.Branding.WatermarkPath); err != nil {
			return fmt.Errorf("品牌水印文件不存在：%w", err)
		} else if info.IsDir() {
			return fmt.Errorf("品牌水印路径不能是目录：%s", p.Branding.WatermarkPath)
		}
		switch p.Branding.WatermarkPosition {
		case "top-left", "top-center", "top-right", "center", "bottom-left", "bottom-center", "bottom-right":
		default:
			return fmt.Errorf("branding.watermarkPosition 不支持：%q", p.Branding.WatermarkPosition)
		}
		if p.Branding.WatermarkWidthRatio <= 0 {
			return fmt.Errorf("branding.watermarkWidthRatio 必须大于 0，当前为 %.3f", p.Branding.WatermarkWidthRatio)
		}
		if p.Branding.WatermarkOpacity <= 0 || p.Branding.WatermarkOpacity > 1 {
			return fmt.Errorf("branding.watermarkOpacity 必须在 (0,1] 之间，当前为 %.3f", p.Branding.WatermarkOpacity)
		}
		if p.Branding.MarginX < 0 {
			return fmt.Errorf("branding.marginX 不能小于 0，当前为 %d", p.Branding.MarginX)
		}
		if p.Branding.MarginY < 0 {
			return fmt.Errorf("branding.marginY 不能小于 0，当前为 %d", p.Branding.MarginY)
		}
		if p.Branding.Start < 0 {
			return fmt.Errorf("branding.start 不能小于 0，当前为 %.2f", p.Branding.Start)
		}
		if p.Branding.End < 0 {
			return fmt.Errorf("branding.end 不能小于 0，当前为 %.2f", p.Branding.End)
		}
		if p.Branding.End > 0 && p.Branding.End <= p.Branding.Start {
			return errors.New("品牌水印的 end 必须大于 start")
		}
	}
	if p.CoverEnabled() {
		if p.Cover.Timestamp < 0 {
			return fmt.Errorf("cover.timestamp 不能小于 0，当前为 %.2f", p.Cover.Timestamp)
		}
		if p.Cover.Quality <= 0 || p.Cover.Quality > 31 {
			return fmt.Errorf("cover.quality 必须在 [1,31] 之间，当前为 %d", p.Cover.Quality)
		}
	}
	return nil
}

// AssetByID 根据素材 ID 查找具体素材。
func (p *Project) AssetByID(id string) (Asset, bool) {
	for _, asset := range p.Assets {
		if asset.ID == id {
			return asset, true
		}
	}
	return Asset{}, false
}

func resolvePath(baseDir, value string) string {
	value = strings.TrimSpace(os.ExpandEnv(value))
	if value == "" {
		return value
	}
	if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
		return value
	}
	if strings.HasPrefix(value, "~"+string(filepath.Separator)) || value == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			if value == "~" {
				value = home
			} else {
				value = filepath.Join(home, strings.TrimPrefix(value, "~"+string(filepath.Separator)))
			}
		}
	}
	if filepath.IsAbs(value) {
		return filepath.Clean(value)
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
