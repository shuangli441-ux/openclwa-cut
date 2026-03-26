package cli

import (
	"path/filepath"
	"testing"

	"github.com/shuangli441-ux/openclwa-cut/internal/ffmpeg"
)

func TestInitTemplateProjectCreatesDouyinGoodsProject(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "input.mp4")
	if err := ffmpeg.Run(
		"ffmpeg",
		"-y",
		"-f", "lavfi",
		"-i", "color=c=black:s=320x240:d=3",
		"-f", "lavfi",
		"-i", "anullsrc=channel_layout=stereo:sample_rate=48000",
		"-shortest",
		"-c:v", "libx264",
		"-c:a", "aac",
		inputPath,
	); err != nil {
		t.Fatal(err)
	}

	projectDir := filepath.Join(tmpDir, "goods-project")
	err := InitTemplateProject(projectDir, "goods-demo", TemplateInitOptions{
		Kind:       TemplateDouyinGoods,
		InputVideo: inputPath,
		Title:      "办公室高频好物推荐",
	})
	if err != nil {
		t.Fatal(err)
	}

	project, err := LoadProject(filepath.Join(projectDir, "project.json"))
	if err != nil {
		t.Fatal(err)
	}
	if project.Music.Style != "goods-recommend" {
		t.Fatalf("expected goods music style, got %q", project.Music.Style)
	}
	if project.Publish.Title != "办公室高频好物推荐" {
		t.Fatalf("expected publish title, got %q", project.Publish.Title)
	}
	if len(project.Publish.Hashtags) == 0 {
		t.Fatalf("expected default hashtags, got %+v", project.Publish)
	}
	if len(project.Timeline) != 6 {
		t.Fatalf("expected 3 clip + 3 subtitle items, got %d", len(project.Timeline))
	}
	if project.AIEdit.Mode != "smart" {
		t.Fatalf("expected smart ai edit mode, got %q", project.AIEdit.Mode)
	}
	if project.AIEdit.TemplateKind != TemplateDouyinGoods {
		t.Fatalf("expected goods ai template kind, got %q", project.AIEdit.TemplateKind)
	}
	if len(project.AIEdit.ScriptLines) != 3 {
		t.Fatalf("expected goods ai script lines, got %+v", project.AIEdit.ScriptLines)
	}
}

func TestInitTemplateProjectCreatesDouyinAdsProject(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "input.mp4")
	if err := ffmpeg.Run(
		"ffmpeg",
		"-y",
		"-f", "lavfi",
		"-i", "color=c=black:s=320x240:d=40",
		"-f", "lavfi",
		"-i", "anullsrc=channel_layout=stereo:sample_rate=48000",
		"-shortest",
		"-c:v", "libx264",
		"-c:a", "aac",
		inputPath,
	); err != nil {
		t.Fatal(err)
	}

	projectDir := filepath.Join(tmpDir, "ads-project")
	err := InitTemplateProject(projectDir, "ads-demo", TemplateInitOptions{
		Kind:       TemplateDouyinAds,
		InputVideo: inputPath,
		Title:      "这套方案能帮你少走弯路",
		BrandName:  "秒账",
		CTA:        "现在私信领取演示版",
	})
	if err != nil {
		t.Fatal(err)
	}

	project, err := LoadProject(filepath.Join(projectDir, "project.json"))
	if err != nil {
		t.Fatal(err)
	}
	if project.Music.Style != "douyin-ads" {
		t.Fatalf("expected ads music style, got %q", project.Music.Style)
	}
	if project.AIEdit.MaxDurationSeconds != 28 {
		t.Fatalf("expected ads max duration 28, got %.2f", project.AIEdit.MaxDurationSeconds)
	}
	if project.AIEdit.TemplateKind != TemplateDouyinAds {
		t.Fatalf("expected ads ai template kind, got %q", project.AIEdit.TemplateKind)
	}
	if len(project.AIEdit.ScriptLines) != 4 {
		t.Fatalf("expected ads ai script lines, got %+v", project.AIEdit.ScriptLines)
	}
	if len(project.Timeline) != 8 {
		t.Fatalf("expected 4 clip + 4 subtitle items, got %d", len(project.Timeline))
	}
	if got := project.Publish.Description; got == "" || got == "四段式投放模板：钩子、痛点、卖点、CTA，适合抖音信息流和本地推广短视频。" {
		t.Fatalf("expected richer publish description, got %q", got)
	}
}
