package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"clawcut/internal/ffmpeg"
)

func TestWriteRenderReport(t *testing.T) {
	tmpDir := t.TempDir()
	reportPath := filepath.Join(tmpDir, "output", "final.render.json")
	report := RenderReport{
		Project:    "demo",
		OutputPath: filepath.Join(tmpDir, "output", "final.mp4"),
		ReportPath: reportPath,
		Cover: CoverRenderReport{
			Enabled: true,
			Path:    filepath.Join(tmpDir, "output", "final_cover.jpg"),
		},
		Audio: AudioRenderReport{
			Enabled:        true,
			DuckingApplied: true,
			VoiceBoost:     1.1,
			HasVoice:       true,
			MusicPath:      filepath.Join(tmpDir, "audio", "bgm.m4a"),
		},
		StartedAt:  time.Now().Format(time.RFC3339),
		FinishedAt: time.Now().Format(time.RFC3339),
	}

	if err := WriteRenderReport(reportPath, report); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatal(err)
	}
	var decoded RenderReport
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Cover.Enabled {
		t.Fatalf("expected cover enabled in decoded report: %+v", decoded)
	}
	if !decoded.Audio.DuckingApplied {
		t.Fatalf("expected ducking info in decoded report: %+v", decoded)
	}
}

func TestBuildRenderReportUsesMediaMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	videoPath := filepath.Join(tmpDir, "final.mp4")
	if err := ffmpeg.Run(
		"ffmpeg",
		"-y",
		"-f", "lavfi",
		"-i", "color=c=black:s=320x240:d=1",
		"-f", "lavfi",
		"-i", "anullsrc=channel_layout=stereo:sample_rate=48000",
		"-shortest",
		"-c:v", "libx264",
		"-c:a", "aac",
		videoPath,
	); err != nil {
		t.Fatal(err)
	}

	project := &Project{
		Project:    "demo",
		Format:     "douyin-vertical",
		FPS:        30,
		Resolution: "1080x1920",
		Output: Output{
			Path: videoPath,
		},
	}

	report, err := BuildRenderReport(
		project,
		videoPath,
		"none",
		"",
		"",
		ffmpeg.AudioMixResult{},
		ffmpeg.OverlayResult{},
		"",
		0,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		t.Fatal(err)
	}
	if report.DurationSeconds <= 0 {
		t.Fatalf("expected positive duration, got %+v", report)
	}
	if report.FileSizeBytes <= 0 {
		t.Fatalf("expected positive file size, got %+v", report)
	}
}
