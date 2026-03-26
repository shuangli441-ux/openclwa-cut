package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPrepareAIEditTimelineBuildsFromScriptLines(t *testing.T) {
	tmpDir := t.TempDir()
	videoPath := filepath.Join(tmpDir, "input.mp4")
	if err := os.WriteFile(videoPath, []byte("video"), 0644); err != nil {
		t.Fatal(err)
	}

	project := &Project{
		Project: "ai-script-lines",
		Assets: []Asset{
			{ID: "main", Type: "video", Path: videoPath},
		},
		AIEdit: AIEditSettings{
			Enabled:            true,
			Mode:               "smart",
			TemplateKind:       TemplateDouyinAds,
			ScriptLines:        []string{"3 步看懂投放逻辑", "先讲痛点，再给方案", "最后引导私信领取试用"},
			MaxDurationSeconds: 18,
			HookSeconds:        3,
			CTASeconds:         4,
		},
		Output: Output{Path: filepath.Join(tmpDir, "output", "video", "final.mp4")},
	}
	project.ApplyDefaults()

	if err := project.PrepareAIEditTimeline(); err != nil {
		t.Fatal(err)
	}
	if len(project.Timeline) != 6 {
		t.Fatalf("expected 3 clip + 3 subtitle items, got %d", len(project.Timeline))
	}
	if project.Timeline[0].Type != "clip" || project.Timeline[0].Asset != "main" {
		t.Fatalf("expected generated clip on main asset, got %+v", project.Timeline[0])
	}
	if project.Timeline[1].Type != "subtitle" || project.Timeline[1].Text != "3 步看懂投放逻辑" {
		t.Fatalf("expected first subtitle from script lines, got %+v", project.Timeline[1])
	}
	totalClipDuration := 0.0
	lastSubtitleEnd := 0.0
	for _, item := range project.Timeline {
		if item.Type == "clip" {
			totalClipDuration += item.End - item.Start
		}
		if item.Type == "subtitle" {
			lastSubtitleEnd = item.End
		}
	}
	if lastSubtitleEnd > totalClipDuration+0.05 {
		t.Fatalf("expected subtitle timeline within composition, got subtitle end %.2f > total %.2f", lastSubtitleEnd, totalClipDuration)
	}
}

func TestPrepareAIEditTimelineBuildsFromSubtitleOnlyTimeline(t *testing.T) {
	tmpDir := t.TempDir()
	videoPath := filepath.Join(tmpDir, "input.mp4")
	if err := os.WriteFile(videoPath, []byte("video"), 0644); err != nil {
		t.Fatal(err)
	}

	project := &Project{
		Project: "ai-subtitle-only",
		Assets: []Asset{
			{ID: "clip1", Type: "video", Path: videoPath},
		},
		Timeline: []TimelineItem{
			{Type: "subtitle", Text: "第一句先抛问题"},
			{Type: "subtitle", Text: "第二句直接给答案"},
			{Type: "subtitle", Text: "第三句引导收藏"},
		},
		AIEdit: AIEditSettings{
			Enabled:            true,
			Mode:               "smart",
			TemplateKind:       TemplateDouyinQA,
			MaxDurationSeconds: 12,
			HookSeconds:        2.5,
			CTASeconds:         3,
		},
		Output: Output{Path: filepath.Join(tmpDir, "output", "video", "final.mp4")},
	}
	project.ApplyDefaults()

	if err := project.Validate(); err != nil {
		t.Fatal(err)
	}
	if len(project.Timeline) != 6 {
		t.Fatalf("expected generated clip timeline, got %d", len(project.Timeline))
	}
	if project.Timeline[0].Type != "clip" || project.Timeline[0].Asset != "clip1" {
		t.Fatalf("expected first generated clip on clip1, got %+v", project.Timeline[0])
	}
	if project.Timeline[5].Type != "subtitle" || project.Timeline[5].Text != "第三句引导收藏" {
		t.Fatalf("expected subtitle preserved, got %+v", project.Timeline[5])
	}
}

func TestResolveAIEditTemplateKindUsesProjectSignals(t *testing.T) {
	project := &Project{
		Project: "秒账投放广告",
		Music:   MusicSettings{Style: "douyin-ads"},
		Publish: PublishSettings{
			Title:       "库存混乱怎么处理",
			Description: "投放版短视频",
		},
	}

	if got := project.ResolveAIEditTemplateKind(); got != TemplateDouyinAds {
		t.Fatalf("expected ads template kind, got %q", got)
	}
}
