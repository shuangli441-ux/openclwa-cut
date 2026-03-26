package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shuangli441-ux/openclwa-cut/internal/ffmpeg"
)

func TestParseCodexScriptResponseSupportsCodeFence(t *testing.T) {
	lines, err := parseCodexScriptResponse([]byte("```json\n{\"scriptLines\":[\"第一句\",\"第二句\",\"第三句\"]}\n```"))
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %+v", lines)
	}
}

func TestBuildCodexScriptPromptIncludesProjectContext(t *testing.T) {
	project := &Project{
		Project: "demo",
		Subtitles: SubtitleSettings{
			MaxCharsPerLine: 16,
		},
		Publish: PublishSettings{
			Title:       "库存混乱怎么处理",
			Description: "适合企业服务投放",
			Hashtags:    []string{"#抖音广告", "#企业数字化"},
		},
		AIEdit: AIEditSettings{
			TemplateKind: TemplateDouyinAds,
			PromptHint:   "品牌语气要稳重",
		},
	}
	prompt := buildCodexScriptPrompt(project, Asset{Path: "/tmp/input.mp4"}, 22)
	for _, expected := range []string{"库存混乱怎么处理", "适合企业服务投放", "#抖音广告", "品牌语气要稳重", "目标句数：4 句"} {
		if !strings.Contains(prompt, expected) {
			t.Fatalf("expected prompt to contain %q, got %s", expected, prompt)
		}
	}
}

func TestGenerateAIScriptProjectUsesConfiguredCommand(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "input.mp4")
	if err := ffmpeg.Run(
		"ffmpeg",
		"-y",
		"-f", "lavfi",
		"-i", "color=c=black:s=320x240:d=8",
		"-f", "lavfi",
		"-i", "anullsrc=channel_layout=stereo:sample_rate=48000",
		"-shortest",
		"-c:v", "libx264",
		"-c:a", "aac",
		inputPath,
	); err != nil {
		t.Fatal(err)
	}

	fakeCodex := filepath.Join(tmpDir, "fake-codex")
	script := `#!/bin/sh
out=""
while [ "$#" -gt 0 ]; do
  if [ "$1" = "-o" ]; then
    out="$2"
    shift 2
    continue
  fi
  shift
done
cat >/dev/null
printf '%s\n' '{"scriptLines":["第一句钩子","第二句方案","第三句CTA"]}' > "$out"
`
	if err := os.WriteFile(fakeCodex, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}

	projectPath := filepath.Join(tmpDir, "project.json")
	project := Project{
		Project: "codex-project",
		Assets: []Asset{
			{ID: "main", Type: "video", Path: inputPath},
		},
		AIEdit: AIEditSettings{
			Enabled:      true,
			Mode:         "smart",
			Provider:     AIProviderCodex,
			Command:      fakeCodex,
			TemplateKind: TemplateDouyinAds,
		},
		Cover: CoverSettings{
			Enabled: true,
			Title:   "库存混乱怎么处理",
		},
		Publish: PublishSettings{
			Title: "库存混乱怎么处理",
		},
		Output: Output{
			Path:     filepath.Join(tmpDir, "output", "video", "final.mp4"),
			Platform: "douyin",
		},
	}
	project.ApplyDefaults()
	if err := writeProjectJSON(projectPath, project); err != nil {
		t.Fatal(err)
	}

	if err := GenerateAIScriptProject(projectPath, AIScriptOptions{Force: true}); err != nil {
		t.Fatal(err)
	}

	updated, err := LoadProject(projectPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.AIEdit.ScriptLines) != 3 {
		t.Fatalf("expected generated script lines, got %+v", updated.AIEdit.ScriptLines)
	}
	if len(updated.Timeline) != 6 {
		t.Fatalf("expected regenerated timeline, got %d items", len(updated.Timeline))
	}
}
