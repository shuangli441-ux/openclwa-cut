package ffmpeg

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// VideoProfile 描述视频输出的统一编码规格。
type VideoProfile struct {
	Width        int
	Height       int
	FPS          int
	VideoCodec   string
	AudioCodec   string
	AudioBitrate string
	Preset       string
	CRF          int
}

// AudioMixOptions 定义背景音乐、ducking 与人声增强参数。
type AudioMixOptions struct {
	Volume           float64
	Loop             bool
	FadeOutSeconds   float64
	Ducking          bool
	DuckingThreshold float64
	DuckingRatio     float64
	DuckingAttackMs  int
	DuckingReleaseMs int
	VoiceEnhance     bool
	VoiceHighpassHz  int
	VoiceLowpassHz   int
	VoiceBoost       float64
}

// AudioFilterSupport 表示当前 FFmpeg 是否具备关键音频滤镜。
type AudioFilterSupport struct {
	Sidechain  bool
	Highpass   bool
	Lowpass    bool
	Compressor bool
	Limiter    bool
}

// AudioMixResult 记录本次音频链路中实际生效的能力。
type AudioMixResult struct {
	HasVoice              bool
	DuckingRequested      bool
	DuckingApplied        bool
	VoiceEnhanceRequested bool
	VoiceEnhanceApplied   bool
	VoiceBoost            float64
}

// OverlayOptions 定义水印叠加的位置、大小和显示时段。
type OverlayOptions struct {
	Position   string
	WidthRatio float64
	Opacity    float64
	MarginX    int
	MarginY    int
	Start      float64
	End        float64
}

// OverlayResult 记录品牌水印的实际叠加结果。
type OverlayResult struct {
	Applied        bool
	OpacityApplied bool
	Position       string
	Width          int
}

// CoverTextOptions 定义封面大字标题样式。
type CoverTextOptions struct {
	FontSize     int
	FontColor    string
	MarginBottom int
}

// Run 执行外部命令，并把输出直接透传到当前终端。
func Run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RunCapture 执行外部命令并返回合并后的标准输出和错误输出。
func RunCapture(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.CombinedOutput()
}

// Health 检查 ffmpeg 与 ffprobe 是否可用。
func Health() error {
	if err := Run("ffmpeg", "-version"); err != nil {
		return fmt.Errorf("ffmpeg 不可用：%w", err)
	}
	if err := Run("ffprobe", "-version"); err != nil {
		return fmt.Errorf("ffprobe 不可用：%w", err)
	}
	return nil
}

// Trim 从输入视频中裁出一个片段。
func Trim(input, start, duration, output string) error {
	return Run("ffmpeg", "-y", "-ss", start, "-i", input, "-t", duration, "-c:v", "libx264", "-c:a", "aac", "-movflags", "+faststart", output)
}

// Compress 按指定 CRF 压缩视频文件。
func Compress(input, output, crf string) error {
	return Run("ffmpeg", "-y", "-i", input, "-c:v", "libx264", "-preset", "medium", "-crf", crf, "-c:a", "aac", "-b:a", "128k", "-movflags", "+faststart", output)
}

// Concat 通过 concat demuxer 拼接多个媒体片段。
func Concat(listFile, output string) error {
	return Run("ffmpeg", "-y", "-f", "concat", "-safe", "0", "-i", listFile, "-c", "copy", output)
}

// RenderSegment 按统一规格渲染一个输入片段。
func RenderSegment(input, start, duration, output string, profile VideoProfile) error {
	return RenderVideoSegment(input, start, duration, output, profile)
}

// RenderVideoSegment 渲染视频素材片段，并在无音轨时自动补静音。
func RenderVideoSegment(input, start, duration, output string, profile VideoProfile) error {
	hasAudio, err := HasAudio(input)
	if err != nil {
		return err
	}
	args := []string{
		"-y",
		"-ss", start,
		"-i", input,
	}
	if !hasAudio {
		args = append(
			args,
			"-f", "lavfi",
			"-t", duration,
			"-i", "anullsrc=channel_layout=stereo:sample_rate=48000",
		)
	}
	args = append(
		args,
		"-t", duration,
		"-vf", BuildScalePadFilter(profile.Width, profile.Height, profile.FPS),
		"-c:v", profile.VideoCodec,
		"-preset", profile.Preset,
		"-crf", strconv.Itoa(profile.CRF),
		"-pix_fmt", "yuv420p",
	)
	if !hasAudio {
		args = append(args, "-map", "0:v:0", "-map", "1:a:0")
	}
	args = append(
		args,
		"-c:a", profile.AudioCodec,
		"-b:a", profile.AudioBitrate,
		"-ar", "48000",
		"-movflags", "+faststart",
		output,
	)
	return Run("ffmpeg", args...)
}

// RenderImageSegment 把静态图片扩展成带静音音轨的视频片段。
func RenderImageSegment(input, duration, output string, profile VideoProfile) error {
	args := []string{
		"-y",
		"-loop", "1",
		"-i", input,
		"-f", "lavfi",
		"-t", duration,
		"-i", "anullsrc=channel_layout=stereo:sample_rate=48000",
		"-t", duration,
		"-vf", BuildScalePadFilter(profile.Width, profile.Height, profile.FPS),
		"-map", "0:v:0",
		"-map", "1:a:0",
		"-c:v", profile.VideoCodec,
		"-preset", profile.Preset,
		"-crf", strconv.Itoa(profile.CRF),
		"-pix_fmt", "yuv420p",
		"-c:a", profile.AudioCodec,
		"-b:a", profile.AudioBitrate,
		"-ar", "48000",
		"-shortest",
		"-movflags", "+faststart",
		output,
	}
	return Run("ffmpeg", args...)
}

// BurnSubtitles 使用 FFmpeg subtitles 滤镜烧录字幕。
func BurnSubtitles(input, subtitleFile, output string, profile VideoProfile, fontsDir string) error {
	filter := "subtitles=filename='" + escapeFilterPath(subtitleFile) + "'"
	if fontsDir != "" {
		filter += ":fontsdir='" + escapeFilterPath(fontsDir) + "'"
	}
	return RenderVideoFilter(input, filter, output, profile)
}

// RenderVideoFilter 对现有视频应用一条视频滤镜链。
func RenderVideoFilter(input, filter, output string, profile VideoProfile) error {
	args := []string{
		"-y",
		"-i", input,
		"-vf", filter,
		"-c:v", profile.VideoCodec,
		"-preset", profile.Preset,
		"-crf", strconv.Itoa(profile.CRF),
		"-pix_fmt", "yuv420p",
		"-c:a", profile.AudioCodec,
		"-b:a", profile.AudioBitrate,
		"-movflags", "+faststart",
		output,
	}
	return Run("ffmpeg", args...)
}

// ConcatSegments 把统一规格的片段稳定拼接成一个成片。
func ConcatSegments(inputs []string, output string, profile VideoProfile) error {
	if len(inputs) == 0 {
		return fmt.Errorf("没有可拼接的片段，请先检查时间线是否生成成功")
	}
	if len(inputs) == 1 {
		args := []string{
			"-y",
			"-i", inputs[0],
			"-c:v", profile.VideoCodec,
			"-preset", profile.Preset,
			"-crf", strconv.Itoa(profile.CRF),
			"-pix_fmt", "yuv420p",
			"-c:a", profile.AudioCodec,
			"-b:a", profile.AudioBitrate,
			"-movflags", "+faststart",
			output,
		}
		return Run("ffmpeg", args...)
	}

	args := []string{"-y"}
	filterParts := make([]string, 0, len(inputs)*2)
	for i, input := range inputs {
		args = append(args, "-i", input)
		filterParts = append(filterParts, fmt.Sprintf("[%d:v:0][%d:a:0]", i, i))
	}
	filter := strings.Join(filterParts, "") + fmt.Sprintf("concat=n=%d:v=1:a=1[vout][aout]", len(inputs))
	args = append(
		args,
		"-filter_complex", filter,
		"-map", "[vout]",
		"-map", "[aout]",
		"-c:v", profile.VideoCodec,
		"-preset", profile.Preset,
		"-crf", strconv.Itoa(profile.CRF),
		"-pix_fmt", "yuv420p",
		"-c:a", profile.AudioCodec,
		"-b:a", profile.AudioBitrate,
		"-ar", "48000",
		"-movflags", "+faststart",
		output,
	)
	return Run("ffmpeg", args...)
}

// MixBackgroundMusic 给成片混入 BGM，并在支持时启用 ducking 与人声增强。
func MixBackgroundMusic(video, audio, output string, profile VideoProfile, opts AudioMixOptions) (AudioMixResult, error) {
	if opts.Volume <= 0 {
		opts.Volume = 0.14
	}
	if opts.DuckingThreshold <= 0 {
		opts.DuckingThreshold = 0.035
	}
	if opts.DuckingRatio <= 0 {
		opts.DuckingRatio = 10
	}
	if opts.DuckingAttackMs <= 0 {
		opts.DuckingAttackMs = 20
	}
	if opts.DuckingReleaseMs <= 0 {
		opts.DuckingReleaseMs = 350
	}
	if opts.VoiceHighpassHz <= 0 {
		opts.VoiceHighpassHz = 120
	}
	if opts.VoiceLowpassHz <= 0 {
		opts.VoiceLowpassHz = 9000
	}
	if opts.VoiceBoost <= 0 {
		opts.VoiceBoost = 1.0
	}
	duration, err := ProbeDuration(video)
	if err != nil {
		return AudioMixResult{}, err
	}
	hasAudio, err := HasAudio(video)
	if err != nil {
		return AudioMixResult{}, err
	}

	args := []string{"-y", "-i", video}
	if opts.Loop {
		args = append(args, "-stream_loop", "-1")
	}
	args = append(args, "-i", audio)

	support := AudioFilterSupport{}
	if hasAudio && opts.Ducking {
		support.Sidechain, err = HasFilter("sidechaincompress")
		if err != nil {
			return AudioMixResult{}, err
		}
	}
	if hasAudio && opts.VoiceEnhance {
		support.Highpass, err = HasFilter("highpass")
		if err != nil {
			return AudioMixResult{}, err
		}
		support.Lowpass, err = HasFilter("lowpass")
		if err != nil {
			return AudioMixResult{}, err
		}
		support.Compressor, err = HasFilter("acompressor")
		if err != nil {
			return AudioMixResult{}, err
		}
		support.Limiter, err = HasFilter("alimiter")
		if err != nil {
			return AudioMixResult{}, err
		}
	}
	filter, result := BuildAudioMixFilter(hasAudio, support, opts, duration)

	args = append(
		args,
		"-filter_complex", filter,
		"-map", "0:v",
		"-map", "[aout]",
		"-c:v", "copy",
		"-c:a", profile.AudioCodec,
		"-b:a", profile.AudioBitrate,
		"-shortest",
		"-movflags", "+faststart",
		output,
	)
	if err := Run("ffmpeg", args...); err != nil {
		return result, err
	}
	return result, nil
}

// BuildAudioMixFilter 构造适配当前环境能力的音频滤镜链。
func BuildAudioMixFilter(hasVoice bool, support AudioFilterSupport, opts AudioMixOptions, duration float64) (string, AudioMixResult) {
	result := AudioMixResult{
		HasVoice:              hasVoice,
		DuckingRequested:      hasVoice && opts.Ducking,
		VoiceEnhanceRequested: hasVoice && opts.VoiceEnhance,
		VoiceBoost:            opts.VoiceBoost,
	}
	bgmChain := fmt.Sprintf("[1:a]aresample=48000,volume=%.3f", opts.Volume)
	if opts.FadeOutSeconds > 0 && duration > opts.FadeOutSeconds {
		start := duration - opts.FadeOutSeconds
		bgmChain += fmt.Sprintf(",afade=t=out:st=%.3f:d=%.3f", start, opts.FadeOutSeconds)
	}
	if !hasVoice {
		return bgmChain + "[bgm];[bgm]anull[aout]", result
	}

	voiceFilters := []string{"[0:a]aresample=48000"}
	if opts.VoiceEnhance {
		if support.Highpass && opts.VoiceHighpassHz > 0 {
			voiceFilters = append(voiceFilters, fmt.Sprintf("highpass=f=%d", opts.VoiceHighpassHz))
			result.VoiceEnhanceApplied = true
		}
		if support.Lowpass && opts.VoiceLowpassHz > 0 {
			voiceFilters = append(voiceFilters, fmt.Sprintf("lowpass=f=%d", opts.VoiceLowpassHz))
			result.VoiceEnhanceApplied = true
		}
		if support.Compressor {
			voiceFilters = append(voiceFilters, "acompressor=threshold=0.089:ratio=3.500:attack=5:release=80:makeup=1")
			result.VoiceEnhanceApplied = true
		}
	}
	if math.Abs(opts.VoiceBoost-1.0) > 0.001 {
		voiceFilters = append(voiceFilters, fmt.Sprintf("volume=%.3f", opts.VoiceBoost))
	}
	if opts.VoiceEnhance && support.Limiter {
		voiceFilters = append(voiceFilters, "alimiter=limit=0.950")
		result.VoiceEnhanceApplied = true
	}
	voiceChain := strings.Join(voiceFilters, ",") + "[voiceprep]"

	if support.Sidechain && opts.Ducking {
		result.DuckingApplied = true
		return voiceChain +
				";[voiceprep]asplit=2[voice_mix][voice_sc]" +
				";" + bgmChain + "[bgmraw]" +
				fmt.Sprintf(
					";[bgmraw][voice_sc]sidechaincompress=threshold=%.3f:ratio=%.3f:attack=%d:release=%d:makeup=1[bgmduck]",
					opts.DuckingThreshold,
					opts.DuckingRatio,
					opts.DuckingAttackMs,
					opts.DuckingReleaseMs,
				) +
				";[voice_mix][bgmduck]amix=inputs=2:duration=first:dropout_transition=2:weights='1 1':normalize=0[aout]",
			result
	}

	return voiceChain +
			";" + bgmChain + "[bgm]" +
			";[voiceprep][bgm]amix=inputs=2:duration=first:dropout_transition=2:weights='1 1':normalize=0[aout]",
		result
}

// ApplyWatermark 把品牌 Logo 按指定位置叠加到视频上。
func ApplyWatermark(video, image, output string, profile VideoProfile, opts OverlayOptions) (OverlayResult, error) {
	if opts.Position == "" {
		opts.Position = "top-right"
	}
	if opts.WidthRatio <= 0 {
		opts.WidthRatio = 0.18
	}
	if opts.Opacity <= 0 || opts.Opacity > 1 {
		opts.Opacity = 0.92
	}
	overlayAvailable, err := HasFilter("overlay")
	if err != nil {
		return OverlayResult{}, err
	}
	if !overlayAvailable {
		return OverlayResult{}, fmt.Errorf("当前 FFmpeg 不支持 overlay 滤镜，无法叠加水印")
	}
	opacityAvailable, err := HasFilter("colorchannelmixer")
	if err != nil {
		return OverlayResult{}, err
	}
	filter, result := BuildWatermarkFilter(profile.Width, profile.Height, opacityAvailable, opts)
	args := []string{
		"-y",
		"-i", video,
		"-loop", "1",
		"-i", image,
		"-filter_complex", filter,
		"-map", "[vout]",
		"-map", "0:a?",
		"-c:v", profile.VideoCodec,
		"-preset", profile.Preset,
		"-crf", strconv.Itoa(profile.CRF),
		"-pix_fmt", "yuv420p",
		"-c:a", "copy",
		"-shortest",
		"-movflags", "+faststart",
		output,
	}
	if err := Run("ffmpeg", args...); err != nil {
		return result, err
	}
	return result, nil
}

// BuildWatermarkFilter 生成品牌水印对应的 FFmpeg overlay 滤镜。
func BuildWatermarkFilter(canvasWidth, canvasHeight int, opacityAvailable bool, opts OverlayOptions) (string, OverlayResult) {
	width := int(float64(canvasWidth) * opts.WidthRatio)
	if width < 64 {
		width = 64
	}
	maxWidth := canvasWidth / 2
	if maxWidth > 0 && width > maxWidth {
		width = maxWidth
	}
	position := normalizeOverlayPosition(opts.Position)
	x, y := overlayPositionXY(position, opts.MarginX, opts.MarginY)
	result := OverlayResult{
		Applied:  true,
		Position: position,
		Width:    width,
	}
	overlaySource := fmt.Sprintf("[1:v]scale=%d:-1", width)
	if opacityAvailable && opts.Opacity < 0.999 {
		overlaySource += fmt.Sprintf(",format=rgba,colorchannelmixer=aa=%.3f", opts.Opacity)
		result.OpacityApplied = true
	}
	overlaySource += "[wm]"
	filter := overlaySource + fmt.Sprintf(";[0:v][wm]overlay=%s:%s", x, y)
	if enableExpr := overlayEnable(opts.Start, opts.End); enableExpr != "" {
		filter += ":enable='" + enableExpr + "'"
	}
	filter += "[vout]"
	_ = canvasHeight
	return filter, result
}

// ExportCoverFrame 从成片中导出指定时间点的封面帧。
func ExportCoverFrame(input, output string, timestamp float64, quality int) error {
	if quality <= 0 {
		quality = 2
	}
	if timestamp < 0 {
		timestamp = 0
	}
	if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
		return err
	}
	args := []string{
		"-y",
		"-ss", fmt.Sprintf("%.3f", timestamp),
		"-i", input,
		"-frames:v", "1",
		"-update", "1",
	}
	switch strings.ToLower(filepath.Ext(output)) {
	case ".jpg", ".jpeg":
		args = append(args, "-q:v", strconv.Itoa(quality))
	}
	args = append(args, output)
	return Run("ffmpeg", args...)
}

// ExportCoverFrameWithTitle 导出封面时叠加标题大字。
func ExportCoverFrameWithTitle(input, output string, timestamp float64, quality int, title string, opts CoverTextOptions) error {
	if strings.TrimSpace(title) == "" {
		return ExportCoverFrame(input, output, timestamp, quality)
	}
	hasDrawtext, err := HasFilter("drawtext")
	if err != nil {
		return err
	}
	if !hasDrawtext {
		return fmt.Errorf("ffmpeg 缺少 drawtext 滤镜")
	}
	if quality <= 0 {
		quality = 2
	}
	if timestamp < 0 {
		timestamp = 0
	}
	if opts.FontSize <= 0 {
		opts.FontSize = 88
	}
	if opts.FontColor == "" {
		opts.FontColor = "#FFFFFF"
	}
	if opts.MarginBottom <= 0 {
		opts.MarginBottom = 240
	}
	if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
		return err
	}
	filter := fmt.Sprintf(
		"drawtext=font='Sans':text='%s':fontcolor=%s:fontsize=%d:borderw=4:bordercolor=black:x=(w-text_w)/2:y=h-th-%d",
		escapeDrawtextText(title),
		drawtextColor(opts.FontColor),
		opts.FontSize,
		opts.MarginBottom,
	)
	args := []string{
		"-y",
		"-ss", fmt.Sprintf("%.3f", timestamp),
		"-i", input,
		"-frames:v", "1",
		"-update", "1",
		"-vf", filter,
	}
	switch strings.ToLower(filepath.Ext(output)) {
	case ".jpg", ".jpeg":
		args = append(args, "-q:v", strconv.Itoa(quality))
	}
	args = append(args, output)
	return Run("ffmpeg", args...)
}

// ProbeDuration 读取媒体总时长，单位为秒。
func ProbeDuration(path string) (float64, error) {
	out, err := RunCapture(
		"ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path,
	)
	if err != nil {
		return 0, fmt.Errorf("读取媒体时长失败：%w", err)
	}
	value := strings.TrimSpace(string(out))
	duration, parseErr := strconv.ParseFloat(value, 64)
	if parseErr != nil {
		return 0, fmt.Errorf("解析媒体时长失败 %q：%w", value, parseErr)
	}
	return duration, nil
}

// ProbeMeanVolume 读取媒体平均音量，单位为 dB。
func ProbeMeanVolume(path string) (float64, error) {
	out, err := RunCapture(
		"ffmpeg",
		"-i", path,
		"-vn",
		"-af", "volumedetect",
		"-f", "null",
		"-",
	)
	if err != nil && !strings.Contains(string(out), "mean_volume") {
		return 0, fmt.Errorf("ffmpeg 音量分析失败: %w", err)
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if !strings.Contains(line, "mean_volume:") {
			continue
		}
		value := strings.TrimSpace(strings.TrimSuffix(strings.SplitN(line, "mean_volume:", 2)[1], " dB"))
		meanVolume, parseErr := strconv.ParseFloat(value, 64)
		if parseErr != nil {
			return 0, fmt.Errorf("解析音量信息失败 %q: %w", value, parseErr)
		}
		return meanVolume, nil
	}
	return 0, fmt.Errorf("未能从 ffmpeg 输出中识别平均音量")
}

// HasAudio 判断媒体文件是否包含音轨。
func HasAudio(path string) (bool, error) {
	out, err := RunCapture(
		"ffprobe",
		"-v", "error",
		"-select_streams", "a",
		"-show_entries", "stream=index",
		"-of", "csv=p=0",
		path,
	)
	if err != nil {
		return false, fmt.Errorf("读取媒体音轨失败：%w", err)
	}
	return strings.TrimSpace(string(out)) != "", nil
}

// HasFilter 判断当前 FFmpeg 是否支持指定滤镜。
func HasFilter(name string) (bool, error) {
	out, err := RunCapture("ffmpeg", "-hide_banner", "-filters")
	if err != nil {
		return false, fmt.Errorf("读取 FFmpeg 滤镜列表失败：%w", err)
	}
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		if fields[1] == name {
			return true, nil
		}
	}
	return false, nil
}

// BuildScalePadFilter 生成等比缩放并补黑边的标准化滤镜。
func BuildScalePadFilter(width, height, fps int) string {
	return fmt.Sprintf(
		"scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2:black,setsar=1,fps=%d,format=yuv420p",
		width,
		height,
		width,
		height,
		fps,
	)
}

// SetVolume 从字符串解析并更新背景音乐音量。
func (opts *AudioMixOptions) SetVolume(value string) error {
	parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return err
	}
	opts.Volume = parsed
	return nil
}

func escapeFilterPath(path string) string {
	path = filepath.Clean(path)
	replacer := strings.NewReplacer(
		`\\`, `\\\\`,
		`\`, `\\`,
		":", `\:`,
		"'", `\'`,
		",", `\,`,
		"[", `\[`,
		"]", `\]`,
	)
	return replacer.Replace(path)
}

func escapeDrawtextText(value string) string {
	replacer := strings.NewReplacer(
		`\`, `\\`,
		":", `\:`,
		"'", `\'`,
		"%", `\%`,
		"[", `\[`,
		"]", `\]`,
		",", `\,`,
		"\n", `\n`,
	)
	return replacer.Replace(value)
}

func drawtextColor(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "white"
	}
	return value
}

func normalizeOverlayPosition(position string) string {
	switch strings.ToLower(strings.TrimSpace(position)) {
	case "top-left", "top-center", "top-right", "center", "bottom-left", "bottom-center", "bottom-right":
		return strings.ToLower(strings.TrimSpace(position))
	default:
		return "top-right"
	}
}

func overlayPositionXY(position string, marginX, marginY int) (string, string) {
	switch position {
	case "top-left":
		return strconv.Itoa(marginX), strconv.Itoa(marginY)
	case "top-center":
		return "(W-w)/2", strconv.Itoa(marginY)
	case "center":
		return "(W-w)/2", "(H-h)/2"
	case "bottom-left":
		return strconv.Itoa(marginX), fmt.Sprintf("H-h-%d", marginY)
	case "bottom-center":
		return "(W-w)/2", fmt.Sprintf("H-h-%d", marginY)
	case "bottom-right":
		return fmt.Sprintf("W-w-%d", marginX), fmt.Sprintf("H-h-%d", marginY)
	default:
		return fmt.Sprintf("W-w-%d", marginX), strconv.Itoa(marginY)
	}
}

func overlayEnable(start, end float64) string {
	if end > start && end > 0 {
		return fmt.Sprintf("between(t,%.3f,%.3f)", start, end)
	}
	if start > 0 {
		return fmt.Sprintf("gte(t,%.3f)", start)
	}
	return ""
}
