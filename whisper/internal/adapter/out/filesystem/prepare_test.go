package filesystem

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wonjinsin/ai-toybox/whisper/internal/port"
	portout "github.com/wonjinsin/ai-toybox/whisper/internal/port/out"
)

func TestResolveModelPathUsesEnvironmentVariable(t *testing.T) {
	modelPath := filepath.Join(t.TempDir(), "custom-model.bin")
	if err := os.WriteFile(modelPath, []byte("model"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("WHISPER_MODEL", modelPath)

	resolved, err := resolveModelPath("")
	if err != nil {
		t.Fatalf("resolveModelPath() error = %v", err)
	}
	if resolved != modelPath {
		t.Errorf("resolveModelPath() = %q, want %q", resolved, modelPath)
	}
}

func TestResolveModelPathUsesWorkingDirectory(t *testing.T) {
	tempDir := t.TempDir()
	modelDir := filepath.Join(tempDir, "models")
	if err := os.Mkdir(modelDir, 0o755); err != nil {
		t.Fatal(err)
	}
	modelPath := filepath.Join(modelDir, "ggml-large-v3.bin")
	if err := os.WriteFile(modelPath, []byte("model"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("WHISPER_MODEL", "")
	t.Chdir(tempDir)

	resolved, err := resolveModelPath("")
	if err != nil {
		t.Fatalf("resolveModelPath() error = %v", err)
	}
	if resolved != modelPath {
		t.Errorf("resolveModelPath() = %q, want %q", resolved, modelPath)
	}
}

func TestResolveVADModelPathDoesNotUseLegacyModel(t *testing.T) {
	modelDir := t.TempDir()
	vadPath := filepath.Join(modelDir, "ggml-silero-v5.1.2.bin")
	if err := os.WriteFile(vadPath, []byte("vad"), 0o644); err != nil {
		t.Fatal(err)
	}

	if resolved := resolveVADModelPath(filepath.Join(modelDir, "ggml-large-v3.bin")); resolved != "" {
		t.Errorf("resolveVADModelPath() = %q, want empty", resolved)
	}
}

func TestResolveVADModelPathPrefersSileroV62(t *testing.T) {
	modelDir := t.TempDir()
	legacyPath := filepath.Join(modelDir, "ggml-silero-v5.1.2.bin")
	if err := os.WriteFile(legacyPath, []byte("legacy vad"), 0o644); err != nil {
		t.Fatal(err)
	}
	v62Path := filepath.Join(modelDir, "ggml-silero-v6.2.0.bin")
	if err := os.WriteFile(v62Path, []byte("v6.2 vad"), 0o644); err != nil {
		t.Fatal(err)
	}

	if resolved := resolveVADModelPath(filepath.Join(modelDir, "ggml-large-v3.bin")); resolved != v62Path {
		t.Errorf("resolveVADModelPath() = %q, want %q", resolved, v62Path)
	}
}

func TestResolveVADModelPathMissingReturnsEmpty(t *testing.T) {
	if resolved := resolveVADModelPath(filepath.Join(t.TempDir(), "ggml-large-v3.bin")); resolved != "" {
		t.Errorf("resolveVADModelPath() = %q, want empty", resolved)
	}
}

func TestResolveRequiredVADModelPathUsesExplicitPath(t *testing.T) {
	t.Parallel()

	modelDir := t.TempDir()
	modelPath := filepath.Join(modelDir, "model.bin")
	explicitPath := filepath.Join(modelDir, "custom-vad.bin")
	if err := os.WriteFile(explicitPath, []byte("vad"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := resolveRequiredVADModelPath(explicitPath, modelPath)
	if err != nil {
		t.Fatalf("resolveRequiredVADModelPath() error = %v", err)
	}
	if got != explicitPath {
		t.Errorf("resolveRequiredVADModelPath() = %q, want %q", got, explicitPath)
	}
}

func TestLoadCorrectionsReadsJSONMap(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "corrections.json")
	if err := os.WriteFile(path, []byte(`{"レッサン":"レッスン"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := loadCorrections(path)
	if err != nil {
		t.Fatalf("loadCorrections() error = %v", err)
	}
	if got["レッサン"] != "レッスン" {
		t.Errorf("loadCorrections() = %#v, want correction map", got)
	}
}

func TestLoadCorrectionsRejectsInvalidJSON(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "corrections.json")
	if err := os.WriteFile(path, []byte(`{"broken":`), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := loadCorrections(path)
	if err == nil || !strings.Contains(err.Error(), "decode corrections file") {
		t.Fatalf("loadCorrections() error = %v, want JSON decode context", err)
	}
}

func TestPrepareResolvesOutputAndSkipsBeforeModelDiscovery(t *testing.T) {
	directory := t.TempDir()
	config := fingerprintFiles(t, directory)
	request := portout.SessionRequest{InputPath: config.InputPath, ModelPath: config.ModelPath,
		VADModelPath: config.VADModelPath, Format: "srt", Language: "ko"}
	prepared, err := Prepare(request)
	if err != nil {
		t.Fatal(err)
	}
	wantOutput := filepath.Join(directory, "input.srt")
	if prepared.Job.InputPath != config.InputPath || prepared.Job.OutputPath != wantOutput || prepared.OutputDirectory != directory {
		t.Fatalf("prepared paths = %#v", prepared)
	}
	if err := os.WriteFile(wantOutput, []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}
	missingModel := request
	missingModel.ModelPath = filepath.Join(directory, "missing.bin")
	if _, err := Prepare(missingModel); !port.IsOutputExistsError(err) {
		t.Fatalf("Prepare() error = %v; want skip before model discovery", err)
	}
	forced := request
	forced.Force = true
	if _, err := Prepare(forced); err != nil {
		t.Fatalf("forced Prepare(): %v", err)
	}
	custom := request
	custom.OutputDir = filepath.Join(directory, "nested", "subtitles")
	prepared, err = Prepare(custom)
	if err != nil {
		t.Fatal(err)
	}
	if prepared.Job.OutputPath != filepath.Join(custom.OutputDir, "input.srt") {
		t.Fatalf("custom output path = %q", prepared.Job.OutputPath)
	}
	if err := prepared.CreateOutputDirectory(); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(custom.OutputDir); err != nil || !info.IsDir() {
		t.Fatalf("output directory = %v, %v", info, err)
	}
}

func TestPrepareRejectsInvalidFilesAndCorrections(t *testing.T) {
	directory := t.TempDir()
	config := fingerprintFiles(t, directory)
	corrections := filepath.Join(directory, "corrections.json")
	if err := os.WriteFile(corrections, []byte(`{"":"replacement"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	valid := portout.SessionRequest{InputPath: config.InputPath, ModelPath: config.ModelPath,
		VADModelPath: config.VADModelPath, Format: "srt", Language: "ko"}
	tests := []struct {
		name    string
		request portout.SessionRequest
		message string
	}{
		{name: "empty input", request: portout.SessionRequest{}, message: "input file path is required"},
		{name: "directory input", request: portout.SessionRequest{InputPath: directory}, message: "not a regular file"},
		{name: "missing input", request: portout.SessionRequest{InputPath: filepath.Join(directory, "missing")}, message: "open input file"},
		{name: "missing model", request: portout.SessionRequest{InputPath: valid.InputPath, ModelPath: filepath.Join(directory, "missing")}, message: "open model file"},
		{name: "missing VAD", request: portout.SessionRequest{InputPath: valid.InputPath, ModelPath: valid.ModelPath, VADModelPath: filepath.Join(directory, "missing")}, message: "open VAD model file"},
		{name: "missing corrections", request: portout.SessionRequest{InputPath: valid.InputPath, ModelPath: valid.ModelPath, VADModelPath: valid.VADModelPath, CorrectionsPath: filepath.Join(directory, "missing")}, message: "open corrections file"},
		{name: "empty correction source", request: portout.SessionRequest{InputPath: valid.InputPath, ModelPath: valid.ModelPath, VADModelPath: valid.VADModelPath, CorrectionsPath: corrections}, message: "contains an empty source"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := Prepare(test.request); err == nil || !strings.Contains(err.Error(), test.message) {
				t.Fatalf("Prepare() error = %v, want %q", err, test.message)
			}
		})
	}
}

func TestResourceCreationFailureRemovesOutputReservation(t *testing.T) {
	directory := t.TempDir()
	prepared := Prepared{OutputDirectory: directory, Job: portout.Job{Format: "srt"}}
	t.Setenv("TMPDIR", filepath.Join(directory, "missing"))
	if _, err := prepared.CreateResources(); err == nil || !strings.Contains(err.Error(), "create temporary audio file") {
		t.Fatalf("CreateResources() error = %v; want temporary audio failure", err)
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("failed resource creation leaked output reservations: %v", entries)
	}
}

func TestResolveRequiredVADModelPathUsesEnvironmentBeforeAdjacentModel(t *testing.T) {
	directory := t.TempDir()
	config := fingerprintFiles(t, directory)
	t.Setenv("WHISPER_VAD_MODEL", config.VADModelPath)
	if got, err := resolveRequiredVADModelPath("", config.ModelPath); err != nil || got != config.VADModelPath {
		t.Fatalf("resolveRequiredVADModelPath() = %q, %v", got, err)
	}
	t.Setenv("WHISPER_VAD_MODEL", "")
	if _, err := resolveRequiredVADModelPath("", config.ModelPath); err == nil || !strings.Contains(err.Error(), "VAD model not found") {
		t.Fatalf("missing VAD error = %v", err)
	}
}
