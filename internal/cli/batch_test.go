package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindProjectFilesMatchesRecursivePattern(t *testing.T) {
	tmpDir := t.TempDir()
	files := []string{
		filepath.Join(tmpDir, "a", "project.json"),
		filepath.Join(tmpDir, "b", "nested", "promo.json"),
		filepath.Join(tmpDir, "notes.txt"),
	}
	for _, path := range files {
		if filepath.Ext(path) == ".txt" {
			if err := os.WriteFile(path, []byte("ignore"), 0644); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	matches, err := FindProjectFiles(tmpDir, "*.json", true)
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}
}
