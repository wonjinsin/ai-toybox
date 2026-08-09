package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Transcribe(ctx context.Context, options Options) (string, error) {
	if err := validateLanguage(options.Language); err != nil {
		return "", err
	}
	if err := validateFormat(options.Format); err != nil {
		return "", err
	}

	inputPath, err := requireRegularFile(options.InputPath, "input")
	if err != nil {
		return "", err
	}
	modelPath, err := resolveModelPath(options.ModelPath)
	if err != nil {
		return "", err
	}

	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		return "", errors.New("ffmpeg not found: install it with 'brew install ffmpeg'")
	}
	whisperPath, err := exec.LookPath("whisper-cli")
	if err != nil {
		return "", errors.New("whisper-cli not found: install it with 'brew install whisper-cpp'")
	}

	outputDir := options.OutputDir
	if outputDir == "" {
		outputDir = filepath.Dir(inputPath)
	}
	outputDir, err = filepath.Abs(outputDir)
	if err != nil {
		return "", fmt.Errorf("resolve output directory: %w", err)
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", fmt.Errorf("create output directory %q: %w", outputDir, err)
	}

	baseName := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	outputBase := filepath.Join(outputDir, baseName)
	outputPath := outputBase + "." + options.Format
	if !options.Force {
		if _, err := os.Stat(outputPath); err == nil {
			return "", fmt.Errorf("output file %q already exists; use --force to overwrite it", outputPath)
		} else if !os.IsNotExist(err) {
			return "", fmt.Errorf("inspect output file %q: %w", outputPath, err)
		}
	}
	outputReservation, err := os.CreateTemp(outputDir, ".whisper-local-output-*")
	if err != nil {
		return "", fmt.Errorf("reserve temporary transcript: %w", err)
	}
	temporaryOutputBase := outputReservation.Name()
	if err := outputReservation.Close(); err != nil {
		os.Remove(temporaryOutputBase)
		return "", fmt.Errorf("close temporary transcript reservation: %w", err)
	}
	defer os.Remove(temporaryOutputBase)
	temporaryOutputPath := temporaryOutputBase + "." + options.Format
	defer os.Remove(temporaryOutputPath)

	tempFile, err := os.CreateTemp("", "whisper-local-*.wav")
	if err != nil {
		return "", fmt.Errorf("create temporary audio file: %w", err)
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)
	if err := tempFile.Close(); err != nil {
		return "", fmt.Errorf("close temporary audio file: %w", err)
	}

	ffmpegArgs := []string{
		"-nostdin", "-hide_banner", "-loglevel", "error", "-y",
		"-i", inputPath,
		"-vn", "-ar", "16000", "-ac", "1", "-c:a", "pcm_s16le",
		tempPath,
	}
	if err := runCommand(ctx, ffmpegPath, ffmpegArgs); err != nil {
		return "", fmt.Errorf("extract audio with ffmpeg: %w", err)
	}

	formatFlag := map[string]string{
		"srt": "-osrt",
		"txt": "-otxt",
		"vtt": "-ovtt",
	}[options.Format]
	whisperArgs := []string{
		"-m", modelPath,
		"-f", tempPath,
		"-l", options.Language,
		"-of", temporaryOutputBase,
		formatFlag,
		// Contain hallucination on non-speech audio: -mc 0 stops it from
		// propagating to later windows, -sns suppresses non-speech tokens.
		"-mc", "0",
		"-sns",
	}
	if vadModelPath := resolveVADModelPath(modelPath); vadModelPath != "" {
		// Default threshold (0.5) keeps quiet speech; 30s max speech duration
		// prevents hallucinated subtitles from spanning minutes.
		whisperArgs = append(whisperArgs,
			"--vad", "-vm", vadModelPath,
			"-vmsd", "30",
		)
	}
	if err := runCommand(ctx, whisperPath, whisperArgs); err != nil {
		return "", fmt.Errorf("transcribe audio with whisper.cpp: %w", err)
	}
	if _, err := requireRegularFile(temporaryOutputPath, "transcript"); err != nil {
		return "", fmt.Errorf("whisper.cpp did not create expected output: %w", err)
	}
	if options.Format == "srt" || options.Format == "vtt" {
		if err := postProcessTranscript(temporaryOutputPath); err != nil {
			return "", err
		}
	}
	if options.Force {
		if err := os.Rename(temporaryOutputPath, outputPath); err != nil {
			return "", fmt.Errorf("replace output file %q: %w", outputPath, err)
		}
	} else if err := os.Link(temporaryOutputPath, outputPath); err != nil {
		if os.IsExist(err) {
			return "", fmt.Errorf("output file %q already exists; use --force to overwrite it", outputPath)
		}
		return "", fmt.Errorf("publish output file %q: %w", outputPath, err)
	}

	return outputPath, nil
}

func resolveModelPath(explicitPath string) (string, error) {
	if explicitPath != "" {
		return requireRegularFile(explicitPath, "model")
	}
	if environmentPath := os.Getenv("WHISPER_MODEL"); environmentPath != "" {
		return requireRegularFile(environmentPath, "model from WHISPER_MODEL")
	}

	candidates := make([]string, 0, 3)
	if workingDirectory, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(workingDirectory, "models", "ggml-large-v3.bin"))
	}
	if executablePath, err := os.Executable(); err == nil {
		if resolvedPath, resolveErr := filepath.EvalSymlinks(executablePath); resolveErr == nil {
			executablePath = resolvedPath
		}
		candidates = append(candidates, filepath.Join(filepath.Dir(executablePath), "..", "models", "ggml-large-v3.bin"))
	}
	if cacheDirectory, err := os.UserCacheDir(); err == nil {
		candidates = append(candidates, filepath.Join(cacheDirectory, "whisper-local", "models", "ggml-large-v3.bin"))
	}

	for _, candidate := range candidates {
		if modelPath, err := requireRegularFile(candidate, "model"); err == nil {
			return modelPath, nil
		}
	}
	return "", fmt.Errorf("Whisper model not found; tried: %s; use --model or WHISPER_MODEL", strings.Join(candidates, ", "))
}

func requireRegularFile(path, label string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("%s file path is required", label)
	}
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve %s file %q: %w", label, path, err)
	}
	info, err := os.Stat(absolutePath)
	if err != nil {
		return "", fmt.Errorf("open %s file %q: %w", label, absolutePath, err)
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("%s path %q is not a regular file", label, absolutePath)
	}
	return absolutePath, nil
}

// resolveVADModelPath looks for the silero VAD model next to the Whisper model.
// Returns "" when absent so transcription still works without VAD.
func resolveVADModelPath(modelPath string) string {
	candidate := filepath.Join(filepath.Dir(modelPath), "ggml-silero-v5.1.2.bin")
	if vadPath, err := requireRegularFile(candidate, "VAD model"); err == nil {
		return vadPath
	}
	return ""
}

func runCommand(ctx context.Context, executable string, args []string) error {
	var stderr bytes.Buffer
	command := exec.CommandContext(ctx, executable, args...)
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			return err
		}
		return fmt.Errorf("%w: %s", err, message)
	}
	return nil
}
