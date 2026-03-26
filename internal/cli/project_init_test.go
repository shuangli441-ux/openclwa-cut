package cli

import (
	"path/filepath"
	"testing"

	"github.com/shuangli441-ux/openclwa-cut/internal/ffmpeg"
)

func TestInitProjectWithOptionsBuildsAIScaffold(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "input.mp4")
	if err := ffmpeg.Run(
		"ffmpeg",
		"-y",
		"-f", "lavfi",
		"-i", "color=c=black:s=320x240:d=12",
		"-f", "lavfi",
		"-i", "anullsrc=channel_layout=stereo:sample_rate=48000",
		"-shortest",
		"-c:v", "libx264",
		"-c:a", "aac",
		inputPath,
	); err != nil {
		t.Fatal(err)
	}

	projectDir := filepath.Join(tmpDir, "project")
	err := InitProjectWithOptions(projectDir, "ai-demo", InitProjectOptions{
		InputVideo:   inputPath,
		Title:        "库存混乱怎么处理",
		TemplateKind: TemplateDouyinAds,
		BrandName:    "秒账",
		CTA:          "现在私信领取试用版",
		ScriptLines: []string{
			"前三秒先抛问题",
			"中段直接给方案",
			"结尾明确 CTA",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	project, err := LoadProject(filepath.Join(projectDir, "project.json"))
	if err != nil {
		t.Fatal(err)
	}
	if project.AIEdit.TemplateKind != TemplateDouyinAds {
		t.Fatalf("expected ads ai kind, got %q", project.AIEdit.TemplateKind)
	}
	if len(project.AIEdit.ScriptLines) != 3 {
		t.Fatalf("expected ai script scaffold, got %+v", project.AIEdit.ScriptLines)
	}
	if len(project.Timeline) != 6 {
		t.Fatalf("expected ai timeline, got %d items", len(project.Timeline))
	}
	if project.Music.Style != "douyin-ads" {
		t.Fatalf("expected ads music style, got %q", project.Music.Style)
	}
	if err := project.Validate(); err != nil {
		t.Fatalf("expected generated project valid, got %v", err)
	}
}

func TestInitProjectWithOptionsCanDisableAIScaffold(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "input.mp4")
	if err := ffmpeg.Run(
		"ffmpeg",
		"-y",
		"-f", "lavfi",
		"-i", "color=c=black:s=320x240:d=4",
		"-f", "lavfi",
		"-i", "anullsrc=channel_layout=stereo:sample_rate=48000",
		"-shortest",
		"-c:v", "libx264",
		"-c:a", "aac",
		inputPath,
	); err != nil {
		t.Fatal(err)
	}

	projectDir := filepath.Join(tmpDir, "project")
	err := InitProjectWithOptions(projectDir, "classic-demo", InitProjectOptions{
		InputVideo:        inputPath,
		DisableAIScaffold: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	project, err := LoadProject(filepath.Join(projectDir, "project.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(project.Timeline) != 1 {
		t.Fatalf("expected classic single clip timeline, got %d items", len(project.Timeline))
	}
	if project.Timeline[0].Type != "clip" || project.Timeline[0].Asset != "main" {
		t.Fatalf("expected classic clip timeline, got %+v", project.Timeline[0])
	}
}
