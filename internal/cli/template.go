package cli

import (
	"encoding/json"
	"os"
)

func ApplyDouyinQATemplate(projectPath string, inputVideo string, outputVideo string) error {
	p := Project{
		Project:    "douyin-qa-template",
		Format:     "douyin-vertical",
		FPS:        30,
		Resolution: "1080x1920",
		Assets:     []Asset{{ID: "clip1", Type: "video", Path: inputVideo}},
		Timeline: []TimelineItem{
			{Type: "clip", Asset: "clip1", Start: 0, End: 3},
			{Type: "subtitle", Start: 0, End: 3, Text: "问题抛出，先抓注意力"},
			{Type: "clip", Asset: "clip1", Start: 3, End: 6},
			{Type: "subtitle", Start: 3, End: 6, Text: "第二段给结论或动作建议"},
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
			VoiceBoost:       1.0,
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
	b, _ := json.MarshalIndent(p, "", "  ")
	return os.WriteFile(projectPath, b, 0644)
}
