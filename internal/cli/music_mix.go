package cli

import (
	"clawcut/internal/ffmpeg"
)

func MusicMix(video, audio, output string, bgmVolume string) error {
	profile := ffmpeg.VideoProfile{
		VideoCodec:   "libx264",
		AudioCodec:   "aac",
		AudioBitrate: "160k",
		Preset:       "medium",
		CRF:          21,
	}
	opts := ffmpeg.AudioMixOptions{
		Volume:           0.14,
		Loop:             true,
		Ducking:          true,
		DuckingThreshold: 0.035,
		DuckingRatio:     10,
		DuckingAttackMs:  20,
		DuckingReleaseMs: 350,
		VoiceEnhance:     true,
		VoiceHighpassHz:  120,
		VoiceLowpassHz:   9000,
		VoiceBoost:       1.0,
	}
	if bgmVolume != "" {
		if err := opts.SetVolume(bgmVolume); err != nil {
			return err
		}
	}
	_, err := ffmpeg.MixBackgroundMusic(video, audio, output, profile, opts)
	return err
}
