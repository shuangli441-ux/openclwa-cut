package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"
)

var defaultSubtitleFonts = []string{
	"/System/Library/Fonts/PingFang.ttc",
	"/System/Library/Fonts/Supplemental/Arial Unicode.ttf",
	"/System/Library/Fonts/Supplemental/Arial.ttf",
}

func WriteASS(project *Project, outputDir string) (string, error) {
	dims, err := project.Dimensions()
	if err != nil {
		return "", err
	}
	path := filepath.Join(outputDir, "subtitles.ass")
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	header := []string{
		"[Script Info]",
		"ScriptType: v4.00+",
		fmt.Sprintf("PlayResX: %d", dims.Width),
		fmt.Sprintf("PlayResY: %d", dims.Height),
		"ScaledBorderAndShadow: yes",
		"",
		"[V4+ Styles]",
		"Format: Name,Fontname,Fontsize,PrimaryColour,SecondaryColour,OutlineColour,BackColour,Bold,Italic,Underline,StrikeOut,ScaleX,ScaleY,Spacing,Angle,BorderStyle,Outline,Shadow,Alignment,MarginL,MarginR,MarginV,Encoding",
		fmt.Sprintf(
			"Style: Default,%s,%d,%s,%s,%s,%s,%d,%d,0,0,100,100,0,0,%d,%d,%d,%d,%d,%d,%d,1",
			assValue(project.Subtitles.FontName),
			project.Subtitles.FontSize,
			assColor(project.Subtitles.PrimaryColor),
			assColor(project.Subtitles.PrimaryColor),
			assColor(project.Subtitles.OutlineColor),
			"&H64000000",
			assBool(project.Subtitles.Bold),
			assBool(project.Subtitles.Italic),
			project.Subtitles.BorderStyle,
			project.Subtitles.Outline,
			project.Subtitles.Shadow,
			project.Subtitles.Alignment,
			project.Subtitles.MarginL,
			project.Subtitles.MarginR,
			project.Subtitles.MarginV,
		),
		"",
		"[Events]",
		"Format: Layer,Start,End,Style,Name,MarginL,MarginR,MarginV,Effect,Text",
	}
	if _, err := fmt.Fprintln(f, strings.Join(header, "\n")); err != nil {
		return "", err
	}
	for _, item := range project.Timeline {
		if item.Type != "subtitle" || item.Text == "" {
			continue
		}
		text := wrapSubtitleText(item.Text, project.Subtitles.MaxCharsPerLine)
		line := fmt.Sprintf(
			"Dialogue: 0,%s,%s,Default,,0,0,0,,%s",
			formatASSTime(item.Start),
			formatASSTime(item.End),
			escapeASSText(text),
		)
		if _, err := fmt.Fprintln(f, line); err != nil {
			return "", err
		}
	}
	return path, nil
}

func formatSRTTime(sec float64) string {
	h := int(sec) / 3600
	m := (int(sec) % 3600) / 60
	s := int(sec) % 60
	ms := int((sec - float64(int(sec))) * 1000)
	return fmt.Sprintf("%02d:%02d:%02d,%03d", h, m, s, ms)
}

func formatASSTime(sec float64) string {
	if sec < 0 {
		sec = 0
	}
	totalCentiseconds := int(sec * 100)
	cs := totalCentiseconds % 100
	totalSeconds := totalCentiseconds / 100
	s := totalSeconds % 60
	totalMinutes := totalSeconds / 60
	m := totalMinutes % 60
	h := totalMinutes / 60
	return fmt.Sprintf("%d:%02d:%02d.%02d", h, m, s, cs)
}

func assColor(hex string) string {
	hex = strings.TrimSpace(strings.TrimPrefix(hex, "#"))
	if len(hex) != 6 {
		return "&H00FFFFFF"
	}
	r := hex[0:2]
	g := hex[2:4]
	b := hex[4:6]
	return "&H00" + strings.ToUpper(b+g+r)
}

func assBool(value bool) int {
	if value {
		return -1
	}
	return 0
}

func assValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "PingFang SC"
	}
	return strings.ReplaceAll(value, ",", "")
}

func wrapSubtitleText(text string, maxChars int) string {
	text = strings.TrimSpace(text)
	if text == "" || maxChars <= 0 {
		return text
	}
	if utf8.RuneCountInString(text) <= maxChars {
		return text
	}
	words := strings.Fields(text)
	if len(words) > 1 {
		return wrapByWords(words, maxChars)
	}
	runes := []rune(text)
	var lines []string
	for len(runes) > 0 {
		n := maxChars
		if len(runes) < n {
			n = len(runes)
		}
		lines = append(lines, string(runes[:n]))
		runes = runes[n:]
	}
	return strings.Join(lines, "\n")
}

func wrapByWords(words []string, maxChars int) string {
	var lines []string
	var current []string
	currentLen := 0
	for _, word := range words {
		wordLen := utf8.RuneCountInString(word)
		need := wordLen
		if len(current) > 0 {
			need++
		}
		if currentLen+need > maxChars && len(current) > 0 {
			lines = append(lines, strings.Join(current, " "))
			current = []string{word}
			currentLen = wordLen
			continue
		}
		current = append(current, word)
		currentLen += need
	}
	if len(current) > 0 {
		lines = append(lines, strings.Join(current, " "))
	}
	return strings.Join(lines, "\n")
}

func escapeASSText(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\n", `\N`)
	text = strings.ReplaceAll(text, "{", `\{`)
	text = strings.ReplaceAll(text, "}", `\}`)
	return text
}

func SubtitleFontsDir(project *Project) string {
	if project.Subtitles.FontFile == "" {
		return ""
	}
	return filepath.Dir(project.Subtitles.FontFile)
}

func DefaultSubtitleFontFile() string {
	for _, path := range defaultSubtitleFonts {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path
		}
	}
	return ""
}

func SubtitleFontNameFromFile(project *Project) {
	if project.Subtitles.FontName != "" || project.Subtitles.FontFile == "" {
		return
	}
	base := filepath.Base(project.Subtitles.FontFile)
	project.Subtitles.FontName = strings.TrimSuffix(base, filepath.Ext(base))
}

func ParseFloatString(value string) (float64, error) {
	return strconv.ParseFloat(strings.TrimSpace(value), 64)
}

func BuildDrawtextFilter(project *Project) (string, error) {
	var filters []string
	for _, item := range project.Timeline {
		if item.Type != "subtitle" || strings.TrimSpace(item.Text) == "" {
			continue
		}
		text := wrapSubtitleText(item.Text, project.Subtitles.MaxCharsPerLine)
		filter := fmt.Sprintf(
			"drawtext=%s:text='%s':fontcolor=%s:fontsize=%d:borderw=%d:bordercolor=%s:x=%s:y=%s:line_spacing=12:enable='%s'",
			drawtextFontOption(project.Subtitles),
			escapeDrawtextValue(text),
			drawtextColor(project.Subtitles.PrimaryColor, "white"),
			project.Subtitles.FontSize,
			project.Subtitles.Outline,
			drawtextColor(project.Subtitles.OutlineColor, "black"),
			drawtextX(project.Subtitles),
			drawtextY(project.Subtitles),
			drawtextEnable(item.Start, item.End),
		)
		if project.Subtitles.Shadow > 0 {
			filter += fmt.Sprintf(":shadowx=%d:shadowy=%d", project.Subtitles.Shadow, project.Subtitles.Shadow)
		}
		filters = append(filters, filter)
	}
	if len(filters) == 0 {
		return "", fmt.Errorf("no subtitle items available for drawtext")
	}
	return strings.Join(filters, ","), nil
}

func drawtextFontOption(style SubtitleSettings) string {
	if style.FontFile != "" {
		return "fontfile='" + escapeDrawtextPath(style.FontFile) + "'"
	}
	if style.FontName != "" {
		return "font='" + escapeDrawtextValue(style.FontName) + "'"
	}
	return "font='Sans'"
}

func drawtextColor(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	if strings.HasPrefix(value, "#") {
		return "0x" + strings.TrimPrefix(value, "#")
	}
	return value
}

func drawtextX(style SubtitleSettings) string {
	switch style.Alignment {
	case 1, 4, 7:
		return strconv.Itoa(style.MarginL)
	case 3, 6, 9:
		return fmt.Sprintf("w-text_w-%d", style.MarginR)
	default:
		return "(w-text_w)/2"
	}
}

func drawtextY(style SubtitleSettings) string {
	switch style.Alignment {
	case 7, 8, 9:
		return strconv.Itoa(style.MarginV)
	case 4, 5, 6:
		return "(h-text_h)/2"
	default:
		return fmt.Sprintf("h-text_h-%d", style.MarginV)
	}
}

func drawtextEnable(start, end float64) string {
	return fmt.Sprintf("between(t\\,%.3f\\,%.3f)", start, end)
}

func escapeDrawtextValue(value string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		":", "\\:",
		"'", "\\'",
		"%", "\\%",
		",", "\\,",
		"[", "\\[",
		"]", "\\]",
		"\n", "\\n",
		"\r", "",
	)
	return replacer.Replace(value)
}

func escapeDrawtextPath(path string) string {
	return escapeDrawtextValue(filepath.Clean(path))
}
