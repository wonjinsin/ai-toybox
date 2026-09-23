package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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

// resolveVADModelPath looks for the required Silero VAD model next to the Whisper model.
func resolveVADModelPath(modelPath string) string {
	candidate := filepath.Join(filepath.Dir(modelPath), "ggml-silero-v6.2.0.bin")
	if vadPath, err := requireRegularFile(candidate, "VAD model"); err == nil {
		return vadPath
	}
	return ""
}

func resolveRequiredVADModelPath(explicitPath, modelPath string) (string, error) {
	if explicitPath != "" {
		return requireRegularFile(explicitPath, "VAD model")
	}
	if environmentPath := os.Getenv("WHISPER_VAD_MODEL"); environmentPath != "" {
		return requireRegularFile(environmentPath, "VAD model from WHISPER_VAD_MODEL")
	}
	if resolved := resolveVADModelPath(modelPath); resolved != "" {
		return resolved, nil
	}
	return "", fmt.Errorf("VAD model not found next to Whisper model %q; use --vad-model or WHISPER_VAD_MODEL", modelPath)
}
