package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteASSIncludesStyleAndWrappedText(t *testing.T) {
	tmpDir := t.TempDir()
	p := &Project{
		Project:    "subtitle-demo",
		Format:     "douyin-vertical",
		FPS:        30,
		Resolution: "1080x1920",
		Subtitles: SubtitleSettings{
			FontName:        "PingFang SC",
			FontSize:        60,
			PrimaryColor:    "#FFFFFF",
			OutlineColor:    "#000000",
			Outline:         3,
			Alignment:       2,
			MarginL:         60,
			MarginR:         60,
			MarginV:         160,
			MaxCharsPerLine: 8,
		},
		Timeline: []TimelineItem{
			{Type: "subtitle", Start: 0, End: 2.5, Text: "这是一个很长很长的测试字幕"},
		},
	}
	p.ApplyDefaults()

	path, err := WriteASS(p, tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(content)
	if !strings.Contains(text, "Style: Default,PingFang SC,60,&H00FFFFFF") {
		t.Fatalf("expected style line in ass, got %s", text)
	}
	if !strings.Contains(text, `\N`) {
		t.Fatalf("expected wrapped subtitle line, got %s", text)
	}
	if filepath.Ext(path) != ".ass" {
		t.Fatalf("expected ass file, got %q", path)
	}
}

func TestBuildDrawtextFilterIncludesExpectedOptions(t *testing.T) {
	p := &Project{
		Project:    "drawtext-demo",
		Format:     "douyin-vertical",
		FPS:        30,
		Resolution: "1080x1920",
		Subtitles: SubtitleSettings{
			FontFile:        "/System/Library/Fonts/PingFang.ttc",
			FontSize:        60,
			PrimaryColor:    "#FFFFFF",
			OutlineColor:    "#000000",
			Outline:         3,
			Alignment:       2,
			MarginL:         60,
			MarginR:         60,
			MarginV:         160,
			MaxCharsPerLine: 10,
		},
		Timeline: []TimelineItem{
			{Type: "subtitle", Start: 1, End: 3, Text: "字幕测试内容"},
		},
	}

	filter, err := BuildDrawtextFilter(p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(filter, "drawtext=") {
		t.Fatalf("expected drawtext filter, got %s", filter)
	}
	if !strings.Contains(filter, "fontfile='/System/Library/Fonts/PingFang.ttc'") {
		t.Fatalf("expected font file in filter, got %s", filter)
	}
	if !strings.Contains(filter, "enable='between(t\\,1.000\\,3.000)'") {
		t.Fatalf("expected enable clause in filter, got %s", filter)
	}
}
