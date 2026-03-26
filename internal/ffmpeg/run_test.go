package ffmpeg

import (
	"strings"
	"testing"
)

func TestBuildAudioMixFilterUsesSidechainWhenAvailable(t *testing.T) {
	filter, result := BuildAudioMixFilter(true, AudioFilterSupport{
		Sidechain:  true,
		Highpass:   true,
		Lowpass:    true,
		Compressor: true,
		Limiter:    true,
	}, AudioMixOptions{
		Volume:           0.14,
		FadeOutSeconds:   1.2,
		Ducking:          true,
		DuckingThreshold: 0.035,
		DuckingRatio:     10,
		DuckingAttackMs:  20,
		DuckingReleaseMs: 350,
		VoiceEnhance:     true,
		VoiceHighpassHz:  120,
		VoiceLowpassHz:   9000,
		VoiceBoost:       1.0,
	}, 6.0)

	if !strings.Contains(filter, "sidechaincompress=") {
		t.Fatalf("expected sidechaincompress in filter, got %s", filter)
	}
	if !strings.Contains(filter, "highpass=f=120") {
		t.Fatalf("expected highpass in filter, got %s", filter)
	}
	if !strings.Contains(filter, "acompressor=") {
		t.Fatalf("expected compressor in filter, got %s", filter)
	}
	if !strings.Contains(filter, "afade=t=out:st=4.800:d=1.200") {
		t.Fatalf("expected fade out in filter, got %s", filter)
	}
	if !result.DuckingApplied {
		t.Fatalf("expected ducking to be applied, got %+v", result)
	}
	if !result.VoiceEnhanceApplied {
		t.Fatalf("expected voice enhancement to be applied, got %+v", result)
	}
}

func TestBuildAudioMixFilterFallsBackWithoutVoice(t *testing.T) {
	filter, result := BuildAudioMixFilter(false, AudioFilterSupport{}, AudioMixOptions{
		Volume:         0.14,
		FadeOutSeconds: 1.0,
	}, 5.0)

	if strings.Contains(filter, "sidechaincompress=") {
		t.Fatalf("did not expect sidechaincompress in filter, got %s", filter)
	}
	if !strings.Contains(filter, "[bgm]anull[aout]") {
		t.Fatalf("expected bgm passthrough filter, got %s", filter)
	}
	if result.HasVoice {
		t.Fatalf("did not expect voice in result, got %+v", result)
	}
}

func TestBuildWatermarkFilterUsesOpacityAndPosition(t *testing.T) {
	filter, result := BuildWatermarkFilter(1080, 1920, true, OverlayOptions{
		Position:   "bottom-right",
		WidthRatio: 0.2,
		Opacity:    0.65,
		MarginX:    40,
		MarginY:    60,
		Start:      1.5,
		End:        6.0,
	})

	if !strings.Contains(filter, "scale=216:-1") {
		t.Fatalf("expected scaled watermark width, got %s", filter)
	}
	if !strings.Contains(filter, "colorchannelmixer=aa=0.650") {
		t.Fatalf("expected opacity chain, got %s", filter)
	}
	if !strings.Contains(filter, "overlay=W-w-40:H-h-60:enable='between(t,1.500,6.000)'") {
		t.Fatalf("expected positioned overlay with enable window, got %s", filter)
	}
	if !result.Applied || !result.OpacityApplied || result.Position != "bottom-right" {
		t.Fatalf("unexpected overlay result: %+v", result)
	}
}
