package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shuangli441-ux/openclwa-cut/internal/ffmpeg"
)

const (
	// TemplateDouyinQA 表示抖音问答模板。
	TemplateDouyinQA = "douyin-qa"
	// TemplateDouyinGoods 表示抖音好物推荐模板。
	TemplateDouyinGoods = "douyin-goods"
	// TemplateDouyinAds 表示抖音投放短视频模板。
	TemplateDouyinAds = "douyin-ads"
)

// TemplateInitOptions 描述模板初始化时的可选输入。
type TemplateInitOptions struct {
	Kind        string
	InputVideo  string
	OutputPath  string
	MusicPath   string
	MusicStyle  string
	LogoPath    string
	Title       string
	BrandName   string
	CTA         string
	AIMode      string
	MaxSeconds  float64
	HookSeconds float64
	CTASeconds  float64
}

// ApplyDouyinQATemplate 使用默认参数生成抖音 QA 模板项目。
func ApplyDouyinQATemplate(projectPath string, inputVideo string, outputVideo string) error {
	return ApplyDouyinQATemplateWithOptions(projectPath, inputVideo, outputVideo, TemplateInitOptions{})
}

// ApplyDouyinQATemplateWithOptions 生成带品牌和标题设置的抖音 QA 模板项目。
func ApplyDouyinQATemplateWithOptions(projectPath string, inputVideo string, outputVideo string, opts TemplateInitOptions) error {
	project, err := buildTemplateProject(TemplateDouyinQA, defaultTemplateProjectName(TemplateDouyinQA), projectPath, inputVideo, outputVideo, opts)
	if err != nil {
		return err
	}
	return writeProjectJSON(projectPath, project)
}

// ApplyDouyinGoodsTemplate 使用默认参数生成抖音好物推荐模板项目。
func ApplyDouyinGoodsTemplate(projectPath string, inputVideo string, outputVideo string) error {
	return ApplyDouyinGoodsTemplateWithOptions(projectPath, inputVideo, outputVideo, TemplateInitOptions{})
}

// ApplyDouyinGoodsTemplateWithOptions 生成带品牌和标题设置的抖音好物推荐模板项目。
func ApplyDouyinGoodsTemplateWithOptions(projectPath string, inputVideo string, outputVideo string, opts TemplateInitOptions) error {
	project, err := buildTemplateProject(TemplateDouyinGoods, defaultTemplateProjectName(TemplateDouyinGoods), projectPath, inputVideo, outputVideo, opts)
	if err != nil {
		return err
	}
	return writeProjectJSON(projectPath, project)
}

// ApplyDouyinAdsTemplate 使用默认参数生成抖音投放模板项目。
func ApplyDouyinAdsTemplate(projectPath string, inputVideo string, outputVideo string) error {
	return ApplyDouyinAdsTemplateWithOptions(projectPath, inputVideo, outputVideo, TemplateInitOptions{})
}

// ApplyDouyinAdsTemplateWithOptions 生成适合抖音投放的钩子型广告模板项目。
func ApplyDouyinAdsTemplateWithOptions(projectPath string, inputVideo string, outputVideo string, opts TemplateInitOptions) error {
	project, err := buildTemplateProject(TemplateDouyinAds, defaultTemplateProjectName(TemplateDouyinAds), projectPath, inputVideo, outputVideo, opts)
	if err != nil {
		return err
	}
	return writeProjectJSON(projectPath, project)
}

// InitTemplateProject 在指定目录中创建带模板风格的完整项目脚手架。
func InitTemplateProject(dir, name string, opts TemplateInitOptions) error {
	if err := ensureProjectDirectories(dir); err != nil {
		return err
	}
	if strings.TrimSpace(name) == "" {
		name = defaultTemplateProjectName(opts.Kind)
	}
	projectPath := filepath.Join(dir, "project.json")
	project, err := buildTemplateProject(opts.Kind, name, projectPath, opts.InputVideo, opts.OutputPath, opts)
	if err != nil {
		return err
	}
	return writeProjectJSON(projectPath, project)
}

func buildTemplateProject(kind, name, projectPath, inputVideo, outputVideo string, opts TemplateInitOptions) (Project, error) {
	kind = normalizeTemplateKind(kind)
	if kind == "" {
		return Project{}, fmt.Errorf("模板类型不支持，请使用 %s、%s 或 %s", TemplateDouyinQA, TemplateDouyinGoods, TemplateDouyinAds)
	}
	if strings.TrimSpace(inputVideo) == "" {
		return Project{}, fmt.Errorf("模板初始化需要提供输入视频，请使用 -input 指定素材文件")
	}

	inputPath, err := filepath.Abs(inputVideo)
	if err != nil {
		return Project{}, fmt.Errorf("解析输入视频路径失败：%w", err)
	}
	info, err := os.Stat(inputPath)
	if err != nil {
		return Project{}, fmt.Errorf("输入视频不存在或无法读取：%w", err)
	}
	if info.IsDir() {
		return Project{}, fmt.Errorf("输入视频不能是目录：%s", inputPath)
	}

	projectDir := filepath.Dir(projectPath)
	if strings.TrimSpace(outputVideo) == "" {
		outputVideo = "output/video/final.mp4"
	}
	title := strings.TrimSpace(opts.Title)
	if title == "" {
		title = name
	}
	brandName := strings.TrimSpace(opts.BrandName)
	cta := strings.TrimSpace(opts.CTA)

	duration := 6.0
	if probed, probeErr := ffmpeg.ProbeDuration(inputPath); probeErr == nil && probed > 0 {
		duration = probed
	}

	project := Project{
		Project:    name,
		Format:     "douyin-vertical",
		FPS:        30,
		Resolution: "1080x1920",
		Assets: []Asset{
			{
				ID:   "main",
				Type: "video",
				Path: normalizeProjectPathValue(projectDir, inputPath),
			},
		},
		Subtitles: SubtitleSettings{
			FontName:        "PingFang SC",
			PrimaryColor:    "#FFFFFF",
			OutlineColor:    "#000000",
			Outline:         3,
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
			VoiceBoost:       1.05,
		},
		Cover: CoverSettings{
			Enabled: true,
			Title:   title,
		},
		AIEdit: AIEditSettings{
			Enabled:      true,
			Mode:         "smart",
			TemplateKind: kind,
		},
		Publish: PublishSettings{
			Title: title,
		},
		Output: Output{
			Path:         outputVideo,
			Platform:     "douyin",
			VideoCodec:   "libx264",
			AudioCodec:   "aac",
			AudioBitrate: "160k",
			Preset:       "medium",
			CRF:          21,
		},
	}
	if strings.TrimSpace(opts.MusicPath) != "" {
		project.Music.Path = normalizeProjectPathValue(projectDir, opts.MusicPath)
	}
	if strings.TrimSpace(opts.MusicStyle) != "" {
		project.Music.Style = strings.TrimSpace(opts.MusicStyle)
	}
	if strings.TrimSpace(opts.LogoPath) != "" {
		project.Branding.WatermarkPath = normalizeProjectPathValue(projectDir, opts.LogoPath)
		project.Branding.WatermarkPosition = "bottom-right"
		project.Branding.WatermarkWidthRatio = 0.14
		project.Branding.WatermarkOpacity = 0.92
		project.Branding.MarginX = 36
		project.Branding.MarginY = 48
	}
	if strings.TrimSpace(opts.AIMode) != "" {
		project.AIEdit.Mode = strings.TrimSpace(opts.AIMode)
	}
	if opts.MaxSeconds > 0 {
		project.AIEdit.MaxDurationSeconds = opts.MaxSeconds
	}
	if opts.HookSeconds > 0 {
		project.AIEdit.HookSeconds = opts.HookSeconds
	}
	if opts.CTASeconds > 0 {
		project.AIEdit.CTASeconds = opts.CTASeconds
	}

	switch kind {
	case TemplateDouyinQA:
		project.Project = valueOrDefault(strings.TrimSpace(name), "douyin-qa-template")
		project.AIEdit.MaxDurationSeconds = maxFloat(project.AIEdit.MaxDurationSeconds, 35)
		project.AIEdit.HookSeconds = maxFloat(project.AIEdit.HookSeconds, 3)
		project.AIEdit.CTASeconds = maxFloat(project.AIEdit.CTASeconds, 5)
		project.AIEdit.ScriptLines = []string{
			"开头先抛问题，快速抓住注意力",
			"中段直接给结论，减少观众流失",
			"结尾补动作建议，引导评论或收藏",
		}
		project.Timeline = BuildSmartTemplateTimeline(duration, "main", project.AIEdit.ScriptLines, kind, project.AIEdit)
		if project.Music.Style == "" {
			project.Music.Style = "qa-short"
		}
		project.Publish.Hashtags = []string{"#知识分享", "#抖音问答", "#短视频干货"}
		project.Publish.Description = "三段式答疑结构，适合财税、运营、职场咨询等知识型短视频。"
	case TemplateDouyinGoods:
		project.Project = valueOrDefault(strings.TrimSpace(name), "douyin-goods-template")
		project.Subtitles.PrimaryColor = "#FFF4C9"
		project.Subtitles.OutlineColor = "#111111"
		project.Subtitles.Outline = 4
		project.Subtitles.MaxCharsPerLine = 16
		project.AIEdit.MaxDurationSeconds = maxFloat(project.AIEdit.MaxDurationSeconds, 30)
		project.AIEdit.HookSeconds = maxFloat(project.AIEdit.HookSeconds, 3)
		project.AIEdit.CTASeconds = maxFloat(project.AIEdit.CTASeconds, 4)
		project.AIEdit.ScriptLines = []string{
			"先说使用场景，开头就把痛点讲透",
			"展示核心卖点，补一句真实体验感受",
			"最后给购买理由，顺手引导点击和收藏",
		}
		project.Timeline = BuildSmartTemplateTimeline(duration, "main", project.AIEdit.ScriptLines, kind, project.AIEdit)
		if project.Music.Style == "" {
			project.Music.Style = "goods-recommend"
		}
		project.Publish.Hashtags = []string{"#好物推荐", "#抖音种草", "#购物分享"}
		project.Publish.Description = "三段式好物推荐结构，适合口播测评、开箱展示和高转化种草内容。"
	case TemplateDouyinAds:
		project.Project = valueOrDefault(strings.TrimSpace(name), "douyin-ads-template")
		project.Subtitles.PrimaryColor = "#FFF2B2"
		project.Subtitles.OutlineColor = "#111111"
		project.Subtitles.Outline = 4
		project.Subtitles.MaxCharsPerLine = 14
		project.Subtitles.MarginV = 200
		project.AIEdit.MaxDurationSeconds = maxFloat(project.AIEdit.MaxDurationSeconds, 28)
		project.AIEdit.HookSeconds = maxFloat(project.AIEdit.HookSeconds, 3)
		project.AIEdit.CTASeconds = maxFloat(project.AIEdit.CTASeconds, 5)
		project.AIEdit.ScriptLines = buildAdScriptLines(title, brandName, cta)
		project.Timeline = BuildSmartTemplateTimeline(duration, "main", project.AIEdit.ScriptLines, kind, project.AIEdit)
		if project.Music.Style == "" {
			project.Music.Style = "douyin-ads"
		}
		project.Publish.Hashtags = []string{"#抖音广告", "#信息流投放", "#短视频投放"}
		project.Publish.Description = buildAdPublishDescription(title, brandName, cta)
	default:
		return Project{}, fmt.Errorf("模板类型不支持：%s", kind)
	}

	project.ApplyDefaults()
	return project, nil
}

func buildAdScriptLines(title, brandName, cta string) []string {
	benefit := "30 秒看完核心卖点，马上知道值不值得试"
	if strings.TrimSpace(title) != "" {
		benefit = title
	}
	hook := "前三秒直接给结果，先把注意力拉住"
	if strings.TrimSpace(title) != "" {
		hook = title
	}
	problem := "把用户最常见痛点说透，降低划走概率"
	if strings.TrimSpace(brandName) != "" {
		problem = brandName + " 这类场景，先讲痛点再给方案更容易转化"
	}
	ctaLine := "结尾明确行动指令，引导点击、私信或咨询"
	if strings.TrimSpace(cta) != "" {
		ctaLine = cta
	}
	return []string{hook, problem, benefit, ctaLine}
}

func buildAdPublishDescription(title, brandName, cta string) string {
	parts := []string{"四段式投放模板：钩子、痛点、卖点、CTA，适合抖音信息流和本地推广短视频。"}
	if strings.TrimSpace(title) != "" {
		parts = append(parts, "主卖点："+strings.TrimSpace(title)+"。")
	}
	if strings.TrimSpace(brandName) != "" {
		parts = append(parts, "品牌："+strings.TrimSpace(brandName)+"。")
	}
	if strings.TrimSpace(cta) != "" {
		parts = append(parts, "行动指令："+strings.TrimSpace(cta)+"。")
	}
	return strings.Join(parts, "")
}

func normalizeTemplateKind(kind string) string {
	switch strings.TrimSpace(strings.ToLower(kind)) {
	case "", TemplateDouyinQA:
		return TemplateDouyinQA
	case TemplateDouyinGoods, "goods", "douyin-goods-recommend":
		return TemplateDouyinGoods
	case TemplateDouyinAds, "ads", "ad", "douyin-promo", "promo":
		return TemplateDouyinAds
	default:
		return ""
	}
}

func defaultTemplateProjectName(kind string) string {
	switch normalizeTemplateKind(kind) {
	case TemplateDouyinGoods:
		return "douyin-goods-template"
	case TemplateDouyinAds:
		return "douyin-ads-template"
	default:
		return "douyin-qa-template"
	}
}

func valueOrDefault(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func maxFloat(value float64, fallback float64) float64 {
	if value <= 0 {
		return fallback
	}
	return value
}
