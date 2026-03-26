package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type MusicLibrary struct {
	Provider string       `json:"provider"`
	Tracks   []MusicTrack `json:"tracks"`
}

type MusicTrack struct {
	ID     string   `json:"id"`
	Title  string   `json:"title"`
	Path   string   `json:"path"`
	Tags   []string `json:"tags"`
	Mood   string   `json:"mood"`
	UseFor []string `json:"useFor"`
}

func InitMusicLibrary(path string) error {
	lib := MusicLibrary{Provider: "local"}
	b, _ := json.MarshalIndent(lib, "", "  ")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, b, 0644)
}

func MatchMusic(libraryPath, style string) error {
	lib, err := LoadMusicLibrary(libraryPath)
	if err != nil {
		return err
	}
	fmt.Println("music match style:", style)
	for _, t := range lib.Tracks {
		for _, u := range t.UseFor {
			if u == style {
				fmt.Printf("MATCH %s | %s | %s\n", t.ID, t.Title, t.Path)
			}
		}
	}
	return nil
}

func LoadMusicLibrary(path string) (*MusicLibrary, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var lib MusicLibrary
	if err := json.Unmarshal(b, &lib); err != nil {
		return nil, err
	}
	baseDir := filepath.Dir(path)
	for i := range lib.Tracks {
		lib.Tracks[i].Path = resolvePath(baseDir, lib.Tracks[i].Path)
	}
	return &lib, nil
}

func (lib *MusicLibrary) FindByStyle(style string) *MusicTrack {
	style = strings.TrimSpace(style)
	if style == "" {
		return nil
	}
	for i := range lib.Tracks {
		for _, u := range lib.Tracks[i].UseFor {
			if u == style {
				return &lib.Tracks[i]
			}
		}
	}
	return nil
}
