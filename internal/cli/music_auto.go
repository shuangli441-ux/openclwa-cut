package cli

import (
	"fmt"
)

func PickMusicForStyle(libraryPath, style string) (string, error) {
	lib, err := LoadMusicLibrary(libraryPath)
	if err != nil {
		return "", err
	}
	if track := lib.FindByStyle(style); track != nil {
		return track.Path, nil
	}
	return "", fmt.Errorf("no music matched style %s", style)
}
