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
			Tags:            inferMusicTags(fileName),
			Mood:            inferMusicMood(fileName),
			UseFor:          inferMusicUseFor(fileName, duration),
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
	style = strings.TrimSpace(strings.ToLower(style))
	if style == "" {
		return true
	}
	for _, item := range track.UseFor {
		if strings.EqualFold(item, style) {
			return true
		}
	}
	for _, item := range track.Tags {
		if strings.EqualFold(item, style) {
			return true
		}
	}
	return false
}

func scoreMusicTrack(track MusicTrack, style string, targetDuration float64, voiceMeanVolume float64) (float64, bool) {
	if style != "" && !trackMatchesStyle(track, style) {
		return 0, false
	}
	score := 0.0
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

func inferMusicMood(name string) string {
	name = strings.ToLower(name)
	switch {
	case strings.Contains(name, "calm"), strings.Contains(name, "soft"), strings.Contains(name, "ambient"):
		return "calm"
	case strings.Contains(name, "happy"), strings.Contains(name, "bright"), strings.Contains(name, "fun"):
		return "bright"
	case strings.Contains(name, "tech"), strings.Contains(name, "business"):
		return "professional"
	default:
		return "general"
	}
}

func inferMusicTags(name string) []string {
	name = strings.ToLower(name)
	replacer := strings.NewReplacer("-", " ", "_", " ")
	fields := strings.Fields(replacer.Replace(name))
	if len(fields) == 0 {
		return []string{"general"}
	}
	seen := map[string]struct{}{}
	tags := make([]string, 0, len(fields))
	for _, field := range fields {
		if _, ok := seen[field]; ok {
			continue
		}
		seen[field] = struct{}{}
		tags = append(tags, field)
	}
	return tags
}

func inferMusicUseFor(name string, duration float64) []string {
	name = strings.ToLower(name)
	var useFor []string
	switch {
	case strings.Contains(name, "qa"), strings.Contains(name, "talk"), strings.Contains(name, "tutorial"):
		useFor = append(useFor, "qa-short", "tutorial-short")
	case strings.Contains(name, "goods"), strings.Contains(name, "product"), strings.Contains(name, "sale"):
		useFor = append(useFor, "goods-recommend", "promo-short")
	default:
		useFor = append(useFor, "general-short")
	}
	switch inferDurationBucket(duration) {
	case "short":
		useFor = append(useFor, "short-form")
	case "medium":
		useFor = append(useFor, "mid-form")
	default:
		useFor = append(useFor, "long-form")
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
