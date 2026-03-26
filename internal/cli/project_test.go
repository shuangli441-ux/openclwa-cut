package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadProjectResolvesRelativePathsAndDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	assetPath := filepath.Join(tmpDir, "assets", "clip.mp4")
	if err := os.MkdirAll(filepath.Dir(assetPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(assetPath, []byte("video"), 0644); err != nil {
		t.Fatal(err)
	}
	projectPath := filepath.Join(tmpDir, "project.json")
	projectJSON := `{
  "project": "demo",
  "assets": [{"id":"clip1","type":"video","path":"assets/clip.mp4"}],
  "timeline": [{"type":"clip","asset":"clip1","start":0,"end":2}],
  "output": {"path":"output/final.mp4"}
}`
	if err := os.WriteFile(projectPath, []byte(projectJSON), 0644); err != nil {
		t.Fatal(err)
	}

	p, err := LoadProject(projectPath)
	if err != nil {
		t.Fatal(err)
	}

	if p.Assets[0].Path != assetPath {
		t.Fatalf("expected resolved asset path %q, got %q", assetPath, p.Assets[0].Path)
	}
	if !strings.HasSuffix(p.Output.Path, filepath.Join("output", "final.mp4")) {
		t.Fatalf("expected resolved output path, got %q", p.Output.Path)
	}
	if p.FPS != 30 {
		t.Fatalf("expected default fps 30, got %d", p.FPS)
	}
	if p.Resolution != "1080x1920" {
		t.Fatalf("expected default resolution, got %q", p.Resolution)
	}
	if err := p.Validate(); err != nil {
		t.Fatalf("expected valid project, got %v", err)
	}
}

func TestValidateRejectsSubtitleBeyondComposition(t *testing.T) {
	tmpDir := t.TempDir()
	assetPath := filepath.Join(tmpDir, "clip.mp4")
	if err := os.WriteFile(assetPath, []byte("video"), 0644); err != nil {
		t.Fatal(err)
	}
	p := &Project{
		Project:    "bad-subtitle",
		Format:     "douyin-vertical",
		FPS:        30,
		Resolution: "1080x1920",
		Assets: []Asset{
			{ID: "clip1", Type: "video", Path: assetPath},
		},
		Timeline: []TimelineItem{
			{Type: "clip", Asset: "clip1", Start: 0, End: 2},
			{Type: "subtitle", Start: 0, End: 4, Text: "too long"},
		},
		Output: Output{Path: filepath.Join(tmpDir, "final.mp4")},
	}
	p.ApplyDefaults()
	if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "字幕超出了成片总时长") {
		t.Fatalf("expected duration validation error, got %v", err)
	}
}

func TestValidateInfersImageAssetType(t *testing.T) {
	tmpDir := t.TempDir()
	imagePath := filepath.Join(tmpDir, "cover.png")
	if err := os.WriteFile(imagePath, []byte("image"), 0644); err != nil {
		t.Fatal(err)
	}
	p := &Project{
		Project:    "image-project",
		Format:     "douyin-vertical",
		FPS:        30,
		Resolution: "1080x1920",
		Assets: []Asset{
			{ID: "cover", Path: imagePath},
		},
		Timeline: []TimelineItem{
			{Type: "clip", Asset: "cover", Start: 0, End: 2},
		},
		Output: Output{Path: filepath.Join(tmpDir, "final.mp4")},
	}
	p.ApplyDefaults()
	if err := p.Validate(); err != nil {
		t.Fatalf("expected valid image project, got %v", err)
	}
	if p.Assets[0].Type != "image" {
		t.Fatalf("expected inferred image type, got %q", p.Assets[0].Type)
	}
}

func TestLoadProjectResolvesBrandingAndCoverDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	assetPath := filepath.Join(tmpDir, "assets", "clip.mp4")
	watermarkPath := filepath.Join(tmpDir, "assets", "logo.png")
	if err := os.MkdirAll(filepath.Dir(assetPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(assetPath, []byte("video"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(watermarkPath, []byte("image"), 0644); err != nil {
		t.Fatal(err)
	}
	projectPath := filepath.Join(tmpDir, "project.json")
	projectJSON := `{
  "project": "brand-demo",
  "assets": [{"id":"clip1","path":"assets/clip.mp4"}],
  "timeline": [{"type":"clip","asset":"clip1","start":0,"end":3}],
  "branding": {"watermarkPath":"assets/logo.png"},
  "cover": {"enabled": true},
  "output": {"path":"output/final.mp4"}
}`
	if err := os.WriteFile(projectPath, []byte(projectJSON), 0644); err != nil {
		t.Fatal(err)
	}

	p, err := LoadProject(projectPath)
	if err != nil {
		t.Fatal(err)
	}

	if p.Branding.WatermarkPath != watermarkPath {
		t.Fatalf("expected resolved watermark path %q, got %q", watermarkPath, p.Branding.WatermarkPath)
	}
	if got := p.ResolveCoverPath(); !strings.HasSuffix(got, filepath.Join("output", "cover", "final_cover.jpg")) {
		t.Fatalf("expected default cover path, got %q", got)
	}
	if got := p.ResolveReportPath(); !strings.HasSuffix(got, filepath.Join("output", "report", "final.render.json")) {
		t.Fatalf("expected default report path, got %q", got)
	}
	if got := p.ResolvePublishPath(); !strings.HasSuffix(got, filepath.Join("output", "report", "final.publish.txt")) {
		t.Fatalf("expected default publish path, got %q", got)
	}
	if err := p.Validate(); err != nil {
		t.Fatalf("expected valid project, got %v", err)
	}
}
