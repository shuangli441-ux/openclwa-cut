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

type AudioFilterSupport struct {
	Sidechain  bool
	Highpass   bool
	Lowpass    bool
	Compressor bool
	Limiter    bool
}

type AudioMixResult struct {
	HasVoice              bool
	DuckingRequested      bool
	DuckingApplied        bool
	VoiceEnhanceRequested bool
	VoiceEnhanceApplied   bool
	VoiceBoost            float64
}

type OverlayOptions struct {
	Position   string
	WidthRatio float64
	Opacity    float64
	MarginX    int
	MarginY    int
	Start      float64
	End        float64
}

type OverlayResult struct {
	Applied        bool
	OpacityApplied bool
	Position       string
	Width          int
}

func Run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func RunCapture(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.CombinedOutput()
}

func Health() error {
	if err := Run("ffmpeg", "-version"); err != nil {
		return fmt.Errorf("ffmpeg unavailable: %w", err)
	}
	if err := Run("ffprobe", "-version"); err != nil {
		return fmt.Errorf("ffprobe unavailable: %w", err)
	}
	return nil
}

func Trim(input, start, duration, output string) error {
	return Run("ffmpeg", "-y", "-ss", start, "-i", input, "-t", duration, "-c:v", "libx264", "-c:a", "aac", "-movflags", "+faststart", output)
}

func Compress(input, output, crf string) error {
	return Run("ffmpeg", "-y", "-i", input, "-c:v", "libx264", "-preset", "medium", "-crf", crf, "-c:a", "aac", "-b:a", "128k", "-movflags", "+faststart", output)
}

func Concat(listFile, output string) error {
	return Run("ffmpeg", "-y", "-f", "concat", "-safe", "0", "-i", listFile, "-c", "copy", output)
}

func RenderSegment(input, start, duration, output string, profile VideoProfile) error {
	return RenderVideoSegment(input, start, duration, output, profile)
}

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

func BurnSubtitles(input, subtitleFile, output string, profile VideoProfile, fontsDir string) error {
	filter := "subtitles=filename='" + escapeFilterPath(subtitleFile) + "'"
	if fontsDir != "" {
		filter += ":fontsdir='" + escapeFilterPath(fontsDir) + "'"
	}
	return RenderVideoFilter(input, filter, output, profile)
}

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

func ConcatSegments(inputs []string, output string, profile VideoProfile) error {
	if len(inputs) == 0 {
		return fmt.Errorf("no input segments to concat")
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
		return OverlayResult{}, fmt.Errorf("ffmpeg overlay filter unavailable")
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
	}
	switch strings.ToLower(filepath.Ext(output)) {
	case ".jpg", ".jpeg":
		args = append(args, "-q:v", strconv.Itoa(quality))
	}
	args = append(args, output)
	return Run("ffmpeg", args...)
}

func ProbeDuration(path string) (float64, error) {
	out, err := RunCapture(
		"ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path,
	)
	if err != nil {
		return 0, fmt.Errorf("ffprobe duration failed: %w", err)
	}
	value := strings.TrimSpace(string(out))
	duration, parseErr := strconv.ParseFloat(value, 64)
	if parseErr != nil {
		return 0, fmt.Errorf("parse duration %q: %w", value, parseErr)
	}
	return duration, nil
}

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
		return false, fmt.Errorf("ffprobe audio streams failed: %w", err)
	}
	return strings.TrimSpace(string(out)) != "", nil
}

func HasFilter(name string) (bool, error) {
	out, err := RunCapture("ffmpeg", "-hide_banner", "-filters")
	if err != nil {
		return false, fmt.Errorf("ffmpeg filter list failed: %w", err)
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
