package filesystem

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestDiscoverInputPathsReturnsSupportedFilesInNameOrder(t *testing.T) {
	t.Parallel()

	inputDir := t.TempDir()
	for _, name := range []string{"voice.amr", "audio.mka", "track.ac3", "b.wav", "a.mp4", "notes.txt"} {
		if err := os.WriteFile(filepath.Join(inputDir, name), []byte("input"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(inputDir, "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inputDir, "nested", "ignored.mp3"), []byte("input"), 0o644); err != nil {
		t.Fatal(err)
	}

	inputs, err := (InputCatalog{}).Discover(inputDir)
	if err != nil {
		t.Fatalf("discoverInputPaths() error = %v", err)
	}

	want := []string{
		filepath.Join(inputDir, "a.mp4"),
		filepath.Join(inputDir, "audio.mka"),
		filepath.Join(inputDir, "b.wav"),
		filepath.Join(inputDir, "track.ac3"),
		filepath.Join(inputDir, "voice.amr"),
	}
	if !reflect.DeepEqual(inputs.Paths, want) {
		t.Errorf("discoverInputPaths() = %q, want %q", inputs.Paths, want)
	}
}

func TestDiscoverInputPathsIgnoresAppleDoubleMediaFiles(t *testing.T) {
	t.Parallel()

	inputDir := t.TempDir()
	for _, name := range []string{"._recording.mp4", "recording.mp4"} {
		if err := os.WriteFile(filepath.Join(inputDir, name), []byte("input"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	inputs, err := (InputCatalog{}).Discover(inputDir)
	if err != nil {
		t.Fatalf("discoverInputPaths() error = %v", err)
	}
	want := []string{filepath.Join(inputDir, "recording.mp4")}
	if !reflect.DeepEqual(inputs.Paths, want) {
		t.Errorf("discoverInputPaths() = %q, want %q", inputs.Paths, want)
	}
}

func TestDiscoverInputPathsRejectsDirectoryWithoutMedia(t *testing.T) {
	t.Parallel()

	_, err := (InputCatalog{}).Discover(t.TempDir())
	if err == nil {
		t.Fatal("discoverInputPaths() error = nil, want empty directory error")
	}
	if !strings.Contains(err.Error(), "no supported media files") {
		t.Errorf("discoverInputPaths() error = %q, want supported media context", err)
	}
}

func TestDiscoverInputPathsKeepsSingleFileInput(t *testing.T) {
	t.Parallel()

	inputPath := filepath.Join(t.TempDir(), "recording.custom")
	if err := os.WriteFile(inputPath, []byte("input"), 0o644); err != nil {
		t.Fatal(err)
	}

	inputs, err := (InputCatalog{}).Discover(inputPath)
	if err != nil {
		t.Fatalf("discoverInputPaths() error = %v", err)
	}
	if !reflect.DeepEqual(inputs.Paths, []string{inputPath}) {
		t.Errorf("discoverInputPaths() = %q, want single input %q", inputs.Paths, inputPath)
	}
}
