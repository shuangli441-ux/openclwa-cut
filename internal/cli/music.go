package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/shuangli441-ux/openclwa-cut/internal/ffmpeg"
)

var musicStyleAliases = map[string][]string{
	"qa-short":        {"qa-short", "tutorial-short", "electronic", "calm", "professional", "business", "general-short"},
	"tutorial-short":  {"tutorial-short", "qa-short", "electronic", "professional", "general-short"},
	"goods-recommend": {"goods-recommend", "promo-short", "upbeat", "pop", "general-short"},
	"promo-short":     {"promo-short", "douyin-ads", "upbeat", "pop", "trap", "electronic"},
	"douyin-ads":      {"douyin-ads", "promo-short", "upbeat", "pop", "trap", "electronic"},
	"general-short":   {"general-short", "electronic", "pop", "upbeat"},
	"upbeat":          {"upbeat", "pop", "promo-short"},
	"pop":             {"pop", "upbeat", "promo-short"},
	"electronic":      {"electronic", "qa-short", "tutorial-short", "douyin-ads"},
	"trap":            {"trap", "douyin-ads", "promo-short"},
	"guofeng":         {"guofeng", "general-short"},
	"professional":    {"professional", "electronic", "qa-short"},
	"business":        {"business", "professional", "qa-short"},
}

// MusicLibrary 表示本地音乐库索引文件。
type MusicLibrary struct {
	Provider string       `json:"provider"`
	Tracks   []MusicTrack `json:"tracks"`
}

// MusicTrack 表示一条可用于自动匹配的背景音乐。
type MusicTrack struct {
	ID              string   `json:"id"`
	Title           string   `json:"title"`
	Path            string   `json:"path"`
	Tags            []string `json:"tags"`
	Mood            string   `json:"mood"`
	UseFor          []string `json:"useFor"`
	DurationSeconds float64  `json:"durationSeconds,omitempty"`
	DurationBucket  string   `json:"durationBucket,omitempty"`
	MeanVolumeDB    float64  `json:"meanVolumeDb,omitempty"`
	Energy          string   `json:"energy,omitempty"`
}

// InitMusicLibrary 初始化一个空的音乐库索引文件。
func InitMusicLibrary(path string) error {
	lib := MusicLibrary{Provider: "local"}
	b, _ := json.MarshalIndent(lib, "", "  ")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, b, 0644)
}

// MatchMusic 按风格输出候选背景音乐，便于人工确认素材池。
func MatchMusic(libraryPath, style string) error {
	lib, err := LoadMusicLibrary(libraryPath)
	if err != nil {
		return err
	}
	fmt.Println("音乐匹配结果")
	fmt.Println("风格：", style)
	matchCount := 0
	for _, t := range lib.Tracks {
		if trackMatchesStyle(t, style) {
			matchCount++
			fmt.Printf("- 命中 %s | %s | %s | %.1f 秒 | 能量 %s\n", t.ID, t.Title, t.Path, t.DurationSeconds, t.Energy)
		}
	}
	if matchCount == 0 {
		fmt.Println("- 没有找到符合当前风格的背景音乐，请先扩充音乐库或更换 style")
	}
	return nil
}

// ScanMusicLibrary 扫描本地音乐目录并生成带元数据的音乐库索引。
func ScanMusicLibrary(musicDir, libraryPath string) error {
	info, err := os.Stat(musicDir)
	if err != nil {
		return fmt.Errorf("音乐目录不存在或无法读取：%w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("音乐目录不能是文件：%s", musicDir)
	}
	var tracks []MusicTrack
	err = filepath.WalkDir(musicDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !isMusicFile(path) {
			return nil
		}
		duration, durationErr := ffmpeg.ProbeDuration(path)
		if durationErr != nil {
			return fmt.Errorf("分析音乐时长失败 %s: %w", path, durationErr)
		}
		meanVolume, volumeErr := ffmpeg.ProbeMeanVolume(path)
		if volumeErr != nil {
			meanVolume = -18
		}
		relPath, err := filepath.Rel(filepath.Dir(libraryPath), path)
		if err != nil {
			relPath = path
		}
		fileName := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		track := MusicTrack{
			ID:              musicTrackID(fileName),
			Title:           fileName,
			Path:            filepath.ToSlash(relPath),
			Tags:            inferMusicTags(fileName, relPath),
			Mood:            inferMusicMood(fileName, relPath),
			UseFor:          inferMusicUseFor(fileName, relPath, duration),
			DurationSeconds: duration,
			DurationBucket:  inferDurationBucket(duration),
			MeanVolumeDB:    meanVolume,
			Energy:          inferMusicEnergy(meanVolume),
		}
		tracks = append(tracks, track)
		return nil
	})
	if err != nil {
		return err
	}
	sort.Slice(tracks, func(i, j int) bool {
		return strings.ToLower(tracks[i].Title) < strings.ToLower(tracks[j].Title)
	})
	lib := MusicLibrary{
		Provider: "local-scan",
		Tracks:   tracks,
	}
	data, err := json.MarshalIndent(lib, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(libraryPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(libraryPath, append(data, '\n'), 0644)
}

// LoadMusicLibrary 读取音乐库并把相对路径解析成绝对路径。
func LoadMusicLibrary(path string) (*MusicLibrary, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var lib MusicLibrary
	if err := json.Unmarshal(b, &lib); err != nil {
		return nil, err
	}
	baseDir := filepath.Dir(path)
	for i := range lib.Tracks {
		lib.Tracks[i].Path = resolvePath(baseDir, lib.Tracks[i].Path)
	}
	return &lib, nil
}

// FindByStyle 按 style 返回第一首匹配的曲目。
func (lib *MusicLibrary) FindByStyle(style string) *MusicTrack {
	style = strings.TrimSpace(style)
	if style == "" {
		return nil
	}
	for i := range lib.Tracks {
		if trackMatchesStyle(lib.Tracks[i], style) {
			return &lib.Tracks[i]
		}
	}
	return nil
}

// FindBest 按风格、时长和人声音量选出最合适的背景音乐。
func (lib *MusicLibrary) FindBest(style string, targetDuration float64, voiceMeanVolume float64) *MusicTrack {
	var best *MusicTrack
	bestScore := 1e9
	for i := range lib.Tracks {
		track := &lib.Tracks[i]
		score, ok := scoreMusicTrack(*track, style, targetDuration, voiceMeanVolume)
		if !ok {
			continue
		}
		if score < bestScore {
			best = track
			bestScore = score
		}
	}
	return best
}

func isMusicFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".mp3", ".m4a", ".aac", ".wav", ".flac", ".ogg":
		return true
	default:
		return false
	}
}

func trackMatchesStyle(track MusicTrack, style string) bool {
	return musicMatchStrength(track, style) > 0
}

func musicMatchStrength(track MusicTrack, style string) int {
	requested := normalizeMusicToken(style)
	if requested == "" {
		return 1
	}
	aliases := expandMusicStyleAliases(requested)
	strength := 0
	for _, item := range track.UseFor {
		value := normalizeMusicToken(item)
		if value == requested {
			return 4
		}
		if containsNormalized(aliases, value) && strength < 3 {
			strength = 3
		}
	}
	for _, item := range track.Tags {
		value := normalizeMusicToken(item)
		if value == requested && strength < 3 {
			strength = 3
		}
		if containsNormalized(aliases, value) && strength < 2 {
			strength = 2
		}
	}
	if normalizeMusicToken(track.Mood) == requested && strength < 1 {
		strength = 1
	}
	return strength
}

func scoreMusicTrack(track MusicTrack, style string, targetDuration float64, voiceMeanVolume float64) (float64, bool) {
	matchStrength := musicMatchStrength(track, style)
	if style != "" && matchStrength == 0 {
		return 0, false
	}
	score := 0.0
	if style != "" {
		score -= float64(matchStrength) * 2.5
	}
	if targetDuration > 0 && track.DurationSeconds > 0 {
		diff := track.DurationSeconds - targetDuration
		if diff < 0 {
			diff = -diff
		}
		score += diff / 6
		if inferDurationBucket(targetDuration) == track.DurationBucket {
			score -= 1
		}
	}
	if voiceMeanVolume > -18 && track.MeanVolumeDB > -18 {
		score += track.MeanVolumeDB + 18
	}
	if voiceMeanVolume > -14 && track.Energy == "high" {
		score += 4
	}
	if strings.Contains(normalizeMusicToken(style), "qa") {
		if track.Energy == "high" {
			score += 2
		}
		if track.Mood == "professional" || track.Mood == "calm" {
			score -= 1
		}
	}
	if strings.Contains(normalizeMusicToken(style), "goods") || strings.Contains(normalizeMusicToken(style), "promo") || strings.Contains(normalizeMusicToken(style), "ads") {
		if track.Energy == "medium" {
			score -= 0.8
		}
		if track.Energy == "high" {
			score -= 1.2
		}
	}
	if style == "" && track.Energy == "low" {
		score -= 0.5
	}
	return score, true
}

func inferDurationBucket(duration float64) string {
	switch {
	case duration <= 20:
		return "short"
	case duration <= 60:
		return "medium"
	default:
		return "long"
	}
}

func inferMusicEnergy(meanVolume float64) string {
	switch {
	case meanVolume >= -12:
		return "high"
	case meanVolume >= -18:
		return "medium"
	default:
		return "low"
	}
}

func inferMusicMood(name string, relPath string) string {
	combined := strings.ToLower(name + " " + relPath)
	switch {
	case strings.Contains(combined, "calm"), strings.Contains(combined, "soft"), strings.Contains(combined, "ambient"), strings.Contains(combined, "guofeng"):
		return "calm"
	case strings.Contains(combined, "happy"), strings.Contains(combined, "bright"), strings.Contains(combined, "fun"), strings.Contains(combined, "upbeat"), strings.Contains(combined, "pop"):
		return "bright"
	case strings.Contains(combined, "tech"), strings.Contains(combined, "business"), strings.Contains(combined, "electronic"):
		return "professional"
	default:
		return "general"
	}
}

func inferMusicTags(name string, relPath string) []string {
	seen := map[string]struct{}{}
	add := func(values ...string) {
		for _, value := range values {
			value = normalizeMusicToken(value)
			if value == "" {
				continue
			}
			if _, ok := seen[value]; ok {
				continue
			}
			seen[value] = struct{}{}
		}
	}

	add(splitMusicTokens(name)...)
	add(pathMusicTokens(relPath)...)

	if len(seen) == 0 {
		add("general")
	}

	tags := make([]string, 0, len(seen))
	for value := range seen {
		tags = append(tags, value)
	}
	sort.Strings(tags)
	return tags
}

func inferMusicUseFor(name string, relPath string, duration float64) []string {
	combined := strings.ToLower(name + " " + relPath)
	useFor := make([]string, 0, 6)
	add := func(values ...string) {
		for _, value := range values {
			value = normalizeMusicToken(value)
			if value == "" {
				continue
			}
			if containsNormalized(useFor, value) {
				continue
			}
			useFor = append(useFor, value)
		}
	}

	switch {
	case strings.Contains(combined, "qa"), strings.Contains(combined, "talk"), strings.Contains(combined, "tutorial"), strings.Contains(combined, "electronic"), strings.Contains(combined, "ambient"):
		add("qa-short", "tutorial-short")
	case strings.Contains(combined, "goods"), strings.Contains(combined, "product"), strings.Contains(combined, "sale"), strings.Contains(combined, "pop"), strings.Contains(combined, "upbeat"):
		add("goods-recommend", "promo-short")
	case strings.Contains(combined, "trap"), strings.Contains(combined, "ads"), strings.Contains(combined, "promo"):
		add("douyin-ads", "promo-short")
	default:
		add("general-short")
	}

	switch inferDurationBucket(duration) {
	case "short":
		add("short-form")
	case "medium":
		add("mid-form")
	default:
		add("long-form")
	}
	return useFor
}

func musicTrackID(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	replacer := strings.NewReplacer(" ", "-", "_", "-", "/", "-", "\\", "-", "--", "-")
	name = replacer.Replace(name)
	name = strings.Trim(name, "-")
	if name == "" {
		return "track"
	}
	return name
}

func expandMusicStyleAliases(style string) []string {
	style = normalizeMusicToken(style)
	if style == "" {
		return nil
	}
	result := []string{style}
	if aliases, ok := musicStyleAliases[style]; ok {
		for _, item := range aliases {
			item = normalizeMusicToken(item)
			if item == "" || containsNormalized(result, item) {
				continue
			}
			result = append(result, item)
		}
	}
	return result
}

func normalizeMusicToken(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	replacer := strings.NewReplacer("-", " ", "_", " ", "/", " ", "\\", " ", ".", " ")
	fields := strings.Fields(replacer.Replace(value))
	return strings.Join(fields, "-")
}

func splitMusicTokens(value string) []string {
	replacer := strings.NewReplacer("-", " ", "_", " ", "/", " ", "\\", " ", ".", " ")
	fields := strings.Fields(strings.ToLower(replacer.Replace(value)))
	result := make([]string, 0, len(fields))
	for _, field := range fields {
		normalized := normalizeMusicToken(field)
		if normalized == "" {
			continue
		}
		result = append(result, normalized)
	}
	return result
}

func pathMusicTokens(relPath string) []string {
	dir := filepath.Dir(relPath)
	if dir == "." || dir == "" {
		return nil
	}
	parts := strings.Split(filepath.ToSlash(dir), "/")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		normalized := normalizeMusicToken(part)
		if normalized == "" {
			continue
		}
		result = append(result, normalized)
	}
	return result
}

func containsNormalized(values []string, target string) bool {
	target = normalizeMusicToken(target)
	for _, value := range values {
		if normalizeMusicToken(value) == target {
			return true
		}
	}
	return false
}
