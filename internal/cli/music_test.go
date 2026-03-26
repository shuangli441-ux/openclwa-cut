package cli

import "testing"

func TestTrackMatchesStyleUsesAliases(t *testing.T) {
	track := MusicTrack{
		Tags:   []string{"electronic"},
		UseFor: []string{"general-short"},
	}
	if !trackMatchesStyle(track, "qa-short") {
		t.Fatalf("expected qa-short to match electronic alias")
	}
	if !trackMatchesStyle(track, "tutorial-short") {
		t.Fatalf("expected tutorial-short to match electronic alias")
	}
}

func TestInferMusicUseForUsesFolderNames(t *testing.T) {
	useFor := inferMusicUseFor("upbeat_001", "upbeat/upbeat_001.mp3", 30)
	if !containsNormalized(useFor, "goods-recommend") {
		t.Fatalf("expected upbeat folder to infer goods-recommend, got %+v", useFor)
	}
	if !containsNormalized(useFor, "promo-short") {
		t.Fatalf("expected upbeat folder to infer promo-short, got %+v", useFor)
	}
}

func TestBuildSmartTemplateTimelineCapsLongVideo(t *testing.T) {
	settings := AIEditSettings{
		Enabled:            true,
		Mode:               "smart",
		MaxDurationSeconds: 28,
		HookSeconds:        3,
		CTASeconds:         5,
	}
	timeline := BuildSmartTemplateTimeline(90, "main", []string{"钩子", "痛点", "卖点", "CTA"}, TemplateDouyinAds, settings)
	if len(timeline) != 8 {
		t.Fatalf("expected 8 timeline items, got %d", len(timeline))
	}
	total := 0.0
	for _, item := range timeline {
		if item.Type != "clip" {
			continue
		}
		total += item.End - item.Start
	}
	if total > 28.1 {
		t.Fatalf("expected capped smart duration, got %.2f", total)
	}
}
