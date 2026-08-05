package app

import "testing"

func TestParseOptionsUsesSafeDefaults(t *testing.T) {
	t.Parallel()

	options, err := ParseOptions([]string{"meeting.mp4"})
	if err != nil {
		t.Fatalf("ParseOptions() error = %v", err)
	}

	if options.InputPath != "meeting.mp4" {
		t.Errorf("InputPath = %q, want %q", options.InputPath, "meeting.mp4")
	}
	if options.Language != "auto" {
		t.Errorf("Language = %q, want %q", options.Language, "auto")
	}
	if options.Format != "txt" {
		t.Errorf("Format = %q, want %q", options.Format, "txt")
	}
}

func TestParseOptionsRequiresInput(t *testing.T) {
	t.Parallel()

	_, err := ParseOptions(nil)
	if err == nil {
		t.Fatal("ParseOptions() error = nil, want input validation error")
	}
}

func TestParseOptionsReadsFlags(t *testing.T) {
	t.Parallel()

	options, err := ParseOptions([]string{
		"-language", "ja",
		"-format", "srt",
		"-model", "/models/ggml-small.bin",
		"-output", "/transcripts",
		"-force",
		"movie.mp4",
	})
	if err != nil {
		t.Fatalf("ParseOptions() error = %v", err)
	}

	if options.InputPath != "movie.mp4" {
		t.Errorf("InputPath = %q, want %q", options.InputPath, "movie.mp4")
	}
	if options.Language != "ja" {
		t.Errorf("Language = %q, want %q", options.Language, "ja")
	}
	if options.Format != "srt" {
		t.Errorf("Format = %q, want %q", options.Format, "srt")
	}
	if options.ModelPath != "/models/ggml-small.bin" {
		t.Errorf("ModelPath = %q, want %q", options.ModelPath, "/models/ggml-small.bin")
	}
	if options.OutputDir != "/transcripts" {
		t.Errorf("OutputDir = %q, want %q", options.OutputDir, "/transcripts")
	}
	if !options.Force {
		t.Error("Force = false, want true")
	}
}

func TestParseOptionsRejectsUnsupportedLanguage(t *testing.T) {
	t.Parallel()

	_, err := ParseOptions([]string{"-language", "invalid", "meeting.wav"})
	if err == nil {
		t.Fatal("ParseOptions() error = nil, want unsupported language error")
	}
}

func TestParseOptionsRejectsUnsupportedFormat(t *testing.T) {
	t.Parallel()

	_, err := ParseOptions([]string{"-format", "docx", "meeting.wav"})
	if err == nil {
		t.Fatal("ParseOptions() error = nil, want unsupported format error")
	}
}
