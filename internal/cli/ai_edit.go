package cli

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/shuangli441-ux/openclwa-cut/internal/ffmpeg"
)

// BuildSmartTemplateTimeline 根据模板类型和 AI 剪辑参数生成更适合短视频投放的时间线。
func BuildSmartTemplateTimeline(totalDuration float64, assetID string, subtitles []string, kind string, settings AIEditSettings) []TimelineItem {
	if len(subtitles) == 0 {
		return nil
	}
	if stringsTrimLower(settings.Mode) == "basic" || !settings.Enabled {
		return buildTemplateTimeline(totalDuration, assetID, subtitles)
	}
	if totalDuration <= 0 {
		totalDuration = float64(len(subtitles)) * 3
	}

	maxDuration := settings.MaxDurationSeconds
	if maxDuration <= 0 || maxDuration > totalDuration {
		maxDuration = totalDuration
	}
	segmentDurations := buildSmartSegmentDurations(maxDuration, subtitles, kind, settings)
	segments := buildSmartSegments(totalDuration, segmentDurations)

	timeline := make([]TimelineItem, 0, len(subtitles)*2)
	compositionStart := 0.0
	for i, text := range subtitles {
		segment := segments[i]
		segmentDuration := segment.End - segment.Start
		compositionEnd := compositionStart + segmentDuration
		timeline = append(timeline,
			TimelineItem{
				Type:  "clip",
				Asset: assetID,
				Start: segment.Start,
				End:   segment.End,
			},
			TimelineItem{
				Type:  "subtitle",
				Start: compositionStart,
				End:   compositionEnd,
				Text:  text,
			},
		)
		compositionStart = compositionEnd
	}
	return timeline
}

// PrepareAIEditTimeline 在缺少 clip 片段时，基于 aiEdit.scriptLines 或已有字幕文本自动补齐时间线。
func (p *Project) PrepareAIEditTimeline() error {
	if p == nil || !p.AIEdit.Enabled || hasClipItems(p.Timeline) {
		return nil
	}

	scriptLines := p.ResolvedAIEditScriptLines()
	if len(scriptLines) == 0 {
		if p.ResolvedAIProvider() == AIProviderCodex {
			return fmt.Errorf("当前项目已配置 aiEdit.provider=codex，但还没有 aiEdit.scriptLines，请先运行 clawcut ai-script -project <project.json> 生成脚本")
		}
		return nil
	}

	assetID, asset, ok := p.primaryAIAsset()
	if !ok {
		return fmt.Errorf("AI 剪辑需要至少一个可用素材，请先在 assets 中提供主视频")
	}

	totalDuration := p.TotalDuration()
	if probed, err := ffmpeg.ProbeDuration(asset.Path); err == nil && probed > 0 {
		totalDuration = probed
	}

	kind := p.ResolveAIEditTemplateKind()
	timeline := BuildSmartTemplateTimeline(totalDuration, assetID, scriptLines, kind, p.AIEdit)
	if len(timeline) == 0 {
		return fmt.Errorf("AI 剪辑没有生成有效时间线，请检查 aiEdit.scriptLines 或字幕文案")
	}
	p.Timeline = timeline
	return nil
}

// ResolvedAIEditScriptLines 返回 AI 剪辑当前应使用的脚本文案。
func (p *Project) ResolvedAIEditScriptLines() []string {
	if p == nil {
		return nil
	}
	lines := normalizeScriptLines(p.AIEdit.ScriptLines)
	if len(lines) > 0 {
		return lines
	}
	lines = make([]string, 0, len(p.Timeline))
	for _, item := range p.Timeline {
		if item.Type != "subtitle" {
			continue
		}
		text := strings.TrimSpace(item.Text)
		if text == "" {
			continue
		}
		lines = append(lines, text)
	}
	return normalizeScriptLines(lines)
}

// ResolveAIEditTemplateKind 推断当前项目应该采用的智能剪辑节奏模板。
func (p *Project) ResolveAIEditTemplateKind() string {
	if p == nil {
		return TemplateDouyinQA
	}
	if strings.TrimSpace(p.AIEdit.TemplateKind) != "" {
		if kind := normalizeTemplateKind(p.AIEdit.TemplateKind); kind != "" {
			return kind
		}
	}

	candidates := []string{
		p.Music.Style,
		p.Project,
		p.ResolvedPublishTitle(),
		p.Publish.Description,
		strings.Join(p.Publish.Hashtags, " "),
		strings.Join(p.ResolvedAIEditScriptLines(), " "),
	}
	combined := stringsTrimLower(strings.Join(candidates, " "))
	switch {
	case containsAny(combined, "ads", "广告", "投放", "promo", "cta", "私信", "领取", "咨询"):
		return TemplateDouyinAds
	case containsAny(combined, "goods", "好物", "种草", "带货", "推荐", "购买"):
		return TemplateDouyinGoods
	default:
		return TemplateDouyinQA
	}
}

type smartSegment struct {
	Start float64
	End   float64
}

func buildSmartSegmentDurations(totalDuration float64, subtitles []string, kind string, settings AIEditSettings) []float64 {
	segmentCount := len(subtitles)
	if segmentCount <= 0 {
		return nil
	}
	weights := templateSegmentWeights(kind, subtitles)
	durations := make([]float64, segmentCount)
	remaining := totalDuration

	for i, weight := range weights {
		durations[i] = totalDuration * weight
		if durations[i] < 2 {
			durations[i] = 2
		}
	}

	if segmentCount > 0 && settings.HookSeconds > 0 {
		durations[0] = math.Min(settings.HookSeconds, totalDuration)
	}
	if segmentCount > 1 && settings.CTASeconds > 0 {
		durations[segmentCount-1] = math.Min(settings.CTASeconds, totalDuration)
	}

	for _, duration := range durations {
		remaining -= duration
	}
	if remaining != 0 && segmentCount > 0 {
		middleStart := 0
		middleEnd := segmentCount
		if segmentCount > 1 {
			middleStart = 1
			middleEnd = segmentCount - 1
		}
		targetCount := middleEnd - middleStart
		if targetCount <= 0 {
			targetCount = segmentCount
			middleStart = 0
			middleEnd = segmentCount
		}
		adjust := remaining / float64(targetCount)
		for i := middleStart; i < middleEnd; i++ {
			durations[i] += adjust
			if durations[i] < 1.5 {
				durations[i] = 1.5
			}
		}
	}

	scale := totalDuration / sumFloat64(durations)
	for i := range durations {
		durations[i] *= scale
	}
	return durations
}

func buildSmartSegments(totalDuration float64, durations []float64) []smartSegment {
	segments := make([]smartSegment, len(durations))
	if len(durations) == 0 {
		return segments
	}
	workingDuration := sumFloat64(durations)
	if totalDuration <= workingDuration+0.4 {
		start := 0.0
		for i, duration := range durations {
			end := start + duration
			if i == len(durations)-1 || end > totalDuration {
				end = totalDuration
			}
			segments[i] = smartSegment{Start: start, End: end}
			start = end
		}
		return segments
	}

	firstDuration := durations[0]
	lastDuration := durations[len(durations)-1]
	segments[0] = smartSegment{Start: 0, End: firstDuration}
	lastStart := totalDuration - lastDuration
	if lastStart < segments[0].End {
		lastStart = segments[0].End
	}
	segments[len(durations)-1] = smartSegment{Start: lastStart, End: totalDuration}

	if len(durations) == 1 || len(durations) == 2 {
		return segments
	}

	leftBound := segments[0].End
	rightBound := segments[len(durations)-1].Start
	for i := 1; i < len(durations)-1; i++ {
		duration := durations[i]
		anchorRatio := float64(i) / float64(len(durations)-1)
		anchor := anchorRatio * totalDuration
		start := anchor - duration/2
		minStart := leftBound
		maxStart := rightBound - duration
		if maxStart < minStart {
			maxStart = minStart
		}
		if start < minStart {
			start = minStart
		}
		if start > maxStart {
			start = maxStart
		}
		end := start + duration
		if end > rightBound {
			end = rightBound
			start = end - duration
		}
		if start < leftBound {
			start = leftBound
		}
		if end <= start {
			end = math.Min(totalDuration, start+duration)
		}
		segments[i] = smartSegment{Start: start, End: end}
		leftBound = end
	}
	return segments
}

func templateSegmentWeights(kind string, subtitles []string) []float64 {
	defaults := []float64{0.18, 0.50, 0.32}
	switch normalizeTemplateKind(kind) {
	case TemplateDouyinAds:
		defaults = []float64{0.16, 0.24, 0.34, 0.26}
	case TemplateDouyinGoods:
		defaults = []float64{0.18, 0.47, 0.35}
	}

	weights := normalizedWeights(len(subtitles), defaults)
	textBias := buildTextPacingBias(subtitles)
	for i := range weights {
		weights[i] *= textBias[i]
	}

	total := sumFloat64(weights)
	if total <= 0 {
		return normalizedWeights(len(subtitles), defaults)
	}
	for i := range weights {
		weights[i] /= total
	}
	return weights
}

func normalizedWeights(segmentCount int, defaults []float64) []float64 {
	weights := make([]float64, segmentCount)
	if segmentCount <= len(defaults) {
		copy(weights, defaults[:segmentCount])
	} else {
		copy(weights, defaults)
		for i := len(defaults); i < segmentCount; i++ {
			weights[i] = 1
		}
	}
	total := sumFloat64(weights)
	if total <= 0 {
		for i := range weights {
			weights[i] = 1 / float64(segmentCount)
		}
		return weights
	}
	for i := range weights {
		weights[i] /= total
	}
	return weights
}

func buildTextPacingBias(subtitles []string) []float64 {
	bias := make([]float64, len(subtitles))
	for i, text := range subtitles {
		value := 1.0
		text = strings.TrimSpace(text)
		length := utf8.RuneCountInString(text)
		switch {
		case length >= 22:
			value += 0.16
		case length >= 14:
			value += 0.08
		case length <= 6:
			value -= 0.06
		}
		if isHookLine(text) {
			value -= 0.08
		}
		if isCTALine(text) {
			value += 0.12
		}
		if i > 0 && i < len(subtitles)-1 && containsAny(text, "因为", "所以", "方案", "步骤", "卖点", "效果", "对比", "原因", "细节") {
			value += 0.06
		}
		if value < 0.72 {
			value = 0.72
		}
		bias[i] = value
	}
	return bias
}

func isHookLine(text string) bool {
	text = stringsTrimLower(text)
	return containsAny(text, "?", "？", "为什么", "怎么", "别再", "立刻", "马上", "3步", "三步", "一定要", "避坑", "别划走", "先看结论")
}

func isCTALine(text string) bool {
	text = stringsTrimLower(text)
	return containsAny(text, "点击", "私信", "评论", "收藏", "关注", "咨询", "下单", "领取", "预约", "试用", "现在")
}

func containsAny(text string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(text, stringsTrimLower(needle)) {
			return true
		}
	}
	return false
}

func sumFloat64(values []float64) float64 {
	total := 0.0
	for _, value := range values {
		total += value
	}
	return total
}

func stringsTrimLower(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeScriptLines(lines []string) []string {
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		result = append(result, line)
	}
	return result
}

func hasClipItems(items []TimelineItem) bool {
	for _, item := range items {
		if item.Type == "clip" && strings.TrimSpace(item.Asset) != "" {
			return true
		}
	}
	return false
}

func (p *Project) primaryAIAsset() (string, Asset, bool) {
	if asset, ok := p.AssetByID("main"); ok {
		return "main", asset, true
	}
	for _, asset := range p.Assets {
		if normalizeAssetType(asset) == "video" {
			return asset.ID, asset, true
		}
	}
	if len(p.Assets) == 0 {
		return "", Asset{}, false
	}
	return p.Assets[0].ID, p.Assets[0], true
}

func buildTemplateTimeline(duration float64, assetID string, subtitles []string) []TimelineItem {
	if len(subtitles) == 0 {
		return nil
	}
	if duration <= 0 {
		duration = float64(len(subtitles)) * 2
	}
	if strings.TrimSpace(assetID) == "" {
		assetID = "main"
	}
	segmentDuration := duration / float64(len(subtitles))
	timeline := make([]TimelineItem, 0, len(subtitles)*2)
	start := 0.0
	compositionStart := 0.0
	for i, text := range subtitles {
		end := start + segmentDuration
		if i == len(subtitles)-1 || end > duration {
			end = duration
		}
		compositionEnd := compositionStart + (end - start)
		timeline = append(timeline,
			TimelineItem{
				Type:  "clip",
				Asset: assetID,
				Start: start,
				End:   end,
			},
			TimelineItem{
				Type:  "subtitle",
				Start: compositionStart,
				End:   compositionEnd,
				Text:  text,
			},
		)
		start = end
		compositionStart = compositionEnd
	}
	return timeline
}
