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
}
