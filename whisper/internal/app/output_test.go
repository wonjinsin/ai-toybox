package app

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func TestPublishTranscriptCopiesExclusivelyWhenHardLinksAreUnsupported(t *testing.T) {
	directory := t.TempDir()
	temporaryPath := filepath.Join(directory, ".temporary.srt")
	outputPath := filepath.Join(directory, "result.srt")
	if err := os.WriteFile(temporaryPath, []byte("subtitle\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := publishTranscript(temporaryPath, outputPath, false, func(string, string) error {
		return &os.LinkError{Op: "link", Old: temporaryPath, New: outputPath, Err: syscall.EOPNOTSUPP}
	})
	if err != nil {
		t.Fatalf("publishTranscript() error = %v", err)
	}
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "subtitle\n" {
		t.Errorf("published transcript = %q, want subtitle", content)
	}
	if _, err := os.Stat(temporaryPath); err != nil {
		t.Errorf("temporary transcript should remain for caller cleanup: %v", err)
	}
}

func TestPublishTranscriptDoesNotOverwriteDuringUnsupportedLinkFallback(t *testing.T) {
	directory := t.TempDir()
	temporaryPath := filepath.Join(directory, ".temporary.srt")
	outputPath := filepath.Join(directory, "result.srt")
	if err := os.WriteFile(temporaryPath, []byte("new subtitle\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(outputPath, []byte("existing subtitle\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := publishTranscript(temporaryPath, outputPath, false, func(string, string) error {
		return &os.LinkError{Op: "link", Old: temporaryPath, New: outputPath, Err: syscall.EOPNOTSUPP}
	})
	if !isOutputExistsError(err) {
		t.Fatalf("publishTranscript() error = %v, want output-exists error", err)
	}
	content, readErr := os.ReadFile(outputPath)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(content) != "existing subtitle\n" {
		t.Errorf("existing transcript = %q, want unchanged content", content)
	}
}

func TestWriteAndPublishTranscriptPreservesRecoveryFileAfterUnexpectedPublishFailure(t *testing.T) {
	directory := t.TempDir()
	temporaryPath := filepath.Join(directory, ".temporary.srt")
	outputPath := filepath.Join(directory, "result.srt")

	err := writeAndPublishTranscript(temporaryPath, outputPath, []byte("recoverable subtitle\n"), false, func(string, string) error {
		return &os.LinkError{Op: "link", Old: temporaryPath, New: outputPath, Err: syscall.EACCES}
	})
	if err == nil {
		t.Fatal("writeAndPublishTranscript() error = nil, want publish failure")
	}
	if !strings.Contains(err.Error(), temporaryPath) {
		t.Errorf("error = %q, want recovery path %q", err, temporaryPath)
	}
	content, readErr := os.ReadFile(temporaryPath)
	if readErr != nil {
		t.Fatalf("read recovery transcript: %v", readErr)
	}
	if string(content) != "recoverable subtitle\n" {
		t.Errorf("recovery transcript = %q, want completed content", content)
	}
	if _, statErr := os.Stat(outputPath); !os.IsNotExist(statErr) {
		t.Errorf("final output exists after publish failure; stat error = %v", statErr)
	}
}

func TestWriteAndPublishTranscriptRemovesTemporaryFileAfterSuccess(t *testing.T) {
	directory := t.TempDir()
	temporaryPath := filepath.Join(directory, ".temporary.srt")
	outputPath := filepath.Join(directory, "result.srt")

	if err := writeAndPublishTranscript(temporaryPath, outputPath, []byte("subtitle\n"), false, os.Link); err != nil {
		t.Fatalf("writeAndPublishTranscript() error = %v", err)
	}
	if _, err := os.Stat(temporaryPath); !os.IsNotExist(err) {
		t.Errorf("temporary transcript still exists; stat error = %v", err)
	}
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "subtitle\n" {
		t.Errorf("published transcript = %q, want subtitle", content)
	}
}

func TestWriteAndPublishTranscriptCopiesWhenHardLinksAreUnsupported(t *testing.T) {
	directory := t.TempDir()
	temporaryPath := filepath.Join(directory, ".temporary.srt")
	outputPath := filepath.Join(directory, "result.srt")

	err := writeAndPublishTranscript(temporaryPath, outputPath, []byte("subtitle\n"), false, func(string, string) error {
		return &os.LinkError{Op: "link", Old: temporaryPath, New: outputPath, Err: syscall.EOPNOTSUPP}
	})
	if err != nil {
		t.Fatalf("writeAndPublishTranscript() error = %v", err)
	}
	if _, statErr := os.Stat(temporaryPath); !os.IsNotExist(statErr) {
		t.Errorf("temporary transcript still exists; stat error = %v", statErr)
	}
	content, readErr := os.ReadFile(outputPath)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(content) != "subtitle\n" {
		t.Errorf("published transcript = %q, want subtitle", content)
	}
}

func TestWriteAndPublishTranscriptPreservesRecoveryFileWhenOutputAppearsDuringPublish(t *testing.T) {
	directory := t.TempDir()
	temporaryPath := filepath.Join(directory, ".temporary.srt")
	outputPath := filepath.Join(directory, "result.srt")
	if err := os.WriteFile(outputPath, []byte("existing subtitle\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := writeAndPublishTranscript(temporaryPath, outputPath, []byte("new subtitle\n"), false, os.Link)
	if !isOutputExistsError(err) {
		t.Fatalf("writeAndPublishTranscript() error = %v, want output-exists error", err)
	}
	if !strings.Contains(err.Error(), temporaryPath) {
		t.Errorf("error = %q, want recovery path %q", err, temporaryPath)
	}
	recoveryContent, recoveryErr := os.ReadFile(temporaryPath)
	if recoveryErr != nil {
		t.Fatalf("read recovery transcript: %v", recoveryErr)
	}
	if string(recoveryContent) != "new subtitle\n" {
		t.Errorf("recovery transcript = %q, want new subtitle", recoveryContent)
	}
	content, readErr := os.ReadFile(outputPath)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(content) != "existing subtitle\n" {
		t.Errorf("existing transcript = %q, want unchanged content", content)
	}
}

func TestWriteAndPublishTranscriptPreservesRecoveryFileAfterForceRenameFailure(t *testing.T) {
	directory := t.TempDir()
	temporaryPath := filepath.Join(directory, ".temporary.srt")
	outputPath := filepath.Join(directory, "missing", "result.srt")

	err := writeAndPublishTranscript(temporaryPath, outputPath, []byte("recoverable subtitle\n"), true, os.Link)
	if err == nil {
		t.Fatal("writeAndPublishTranscript() error = nil, want rename failure")
	}
	if !strings.Contains(err.Error(), temporaryPath) {
		t.Errorf("error = %q, want recovery path %q", err, temporaryPath)
	}
	content, readErr := os.ReadFile(temporaryPath)
	if readErr != nil {
		t.Fatalf("read recovery transcript: %v", readErr)
	}
	if string(content) != "recoverable subtitle\n" {
		t.Errorf("recovery transcript = %q, want completed content", content)
	}
}
