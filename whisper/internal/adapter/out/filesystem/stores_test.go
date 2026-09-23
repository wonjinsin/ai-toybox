package filesystem

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/wonjinsin/ai-toybox/whisper/internal/domain"
	"github.com/wonjinsin/ai-toybox/whisper/internal/port"
	portout "github.com/wonjinsin/ai-toybox/whisper/internal/port/out"
)

func TestCheckpointStoreLoadsExistingSchemaAndPreservesRawCues(t *testing.T) {
	directory := t.TempDir()
	config := fingerprintFiles(t, directory)
	store, err := NewCheckpointStore(filepath.Join(directory, "input"), config)
	if err != nil {
		t.Fatal(err)
	}
	fingerprintJSON, err := json.Marshal(store.fingerprint)
	if err != nil {
		t.Fatal(err)
	}
	// This fixture uses the pre-refactor JSON schema, including nanosecond times.
	fixture := fmt.Sprintf(`{"version":1,"fingerprint":%s,"stage":"retrying","batch_size":128,"media_duration":90000000000,"completed_chunks":1,"retry_cursor":0,"chunks":[{"start":10000000000,"end":25000000000}],"cues":[{"start":11000000000,"end":12000000000,"text":"raw cue","probability":0.05,"tokens":[{"start":11000000000,"end":12000000000,"text":"raw cue","probability":0.04,"origin":{"index":37,"start":10000000000,"end":25000000000}}],"origin":{"index":37,"start":10000000000,"end":25000000000}}]}`, fingerprintJSON)
	if err := os.WriteFile(store.path, []byte(fixture), 0o600); err != nil {
		t.Fatal(err)
	}
	origin := domain.Origin{Index: 37, Start: 10 * time.Second, End: 25 * time.Second}
	want := portout.ResumeState{
		Stage: portout.StageRetrying, MediaDuration: 90 * time.Second, CompletedChunks: 1,
		Chunks: []domain.Chunk{{Start: 10 * time.Second, End: 25 * time.Second}},
		Cues: []domain.Cue{{Start: 11 * time.Second, End: 12 * time.Second, Text: "raw cue", Probability: 0.05, Origin: origin,
			Tokens: []domain.Token{{Start: 11 * time.Second, End: 12 * time.Second, Text: "raw cue", Probability: 0.04, Origin: origin}}}},
	}
	loaded, found, err := store.Load()
	if err != nil || !found || !reflect.DeepEqual(loaded, want) {
		t.Fatalf("Load() = %#v, %t, %v; want %#v", loaded, found, err, want)
	}
	if err := store.Save(loaded); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(store.path)
	if err != nil {
		t.Fatal(err)
	}
	var original, saved any
	if err := json.Unmarshal([]byte(fixture), &original); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(content, &saved); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(saved, original) {
		t.Fatalf("saved checkpoint = %s, want original schema %s", content, fixture)
	}
	info, err := os.Stat(store.path)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("checkpoint mode = %v, error = %v; want 0600", info, err)
	}
}

func TestCheckpointStoreDetectsChangedFilesAndResolvesSymlinks(t *testing.T) {
	directory := t.TempDir()
	config := fingerprintFiles(t, directory)
	link := filepath.Join(directory, "linked-input.wav")
	if err := os.Symlink(config.InputPath, link); err != nil {
		t.Fatal(err)
	}
	linkedConfig := config
	linkedConfig.InputPath = link
	store, err := NewCheckpointStore(filepath.Join(directory, "output"), linkedConfig)
	if err != nil {
		t.Fatal(err)
	}
	resolvedInput, err := filepath.EvalSymlinks(config.InputPath)
	if err != nil {
		t.Fatal(err)
	}
	if store.fingerprint.Input.Path != resolvedInput {
		t.Fatalf("input identity path = %q, want resolved %q", store.fingerprint.Input.Path, resolvedInput)
	}
	if _, found, err := store.Load(); err != nil || found {
		t.Fatalf("missing checkpoint Load() = %t, %v", found, err)
	}
	state := portout.ResumeState{Stage: portout.StageTranscribing, MediaDuration: time.Minute}
	if err := store.Save(state); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config.ModelPath, []byte("changed model contents"), 0o600); err != nil {
		t.Fatal(err)
	}
	changed, err := NewCheckpointStore(filepath.Join(directory, "output"), config)
	if err != nil {
		t.Fatal(err)
	}
	if _, found, err := changed.Load(); err != nil || found {
		t.Fatalf("changed model Load() = %t, %v; want false, nil", found, err)
	}
}

func TestCheckpointStoreErrorsRetainCheckpointContext(t *testing.T) {
	directory := t.TempDir()
	store, err := NewCheckpointStore(filepath.Join(directory, "input"), fingerprintFiles(t, directory))
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Save(portout.ResumeState{Stage: "unsupported", MediaDuration: time.Minute}); err != nil {
		t.Fatal(err)
	}
	if _, found, err := store.Load(); err == nil || found || !strings.Contains(err.Error(), store.path) || !strings.Contains(err.Error(), "unsupported stage") {
		t.Fatalf("invalid state Load() = %t, %v; want checkpoint path and validation error", found, err)
	}
	config := fingerprintFiles(t, t.TempDir())
	config.FFprobePath = directory
	if _, err := NewCheckpointStore(filepath.Join(directory, "other"), config); err == nil || !strings.Contains(err.Error(), "identify ffprobe for checkpoint") {
		t.Fatalf("NewCheckpointStore() error = %v; want invalid ffprobe context", err)
	}
}

func TestTranscriptStorePublishesAndRemovesProgress(t *testing.T) {
	prepared, resources := transcriptResources(t, "srt", false)
	t.Cleanup(func() {
		if err := resources.Close(); err != nil {
			t.Error(err)
		}
	})
	store := NewTranscriptStore(prepared, resources)
	cues := []domain.Cue{{Start: time.Second, End: 2 * time.Second, Text: "ready", Probability: 0.01}}
	if err := store.WritePartial(cues); err != nil {
		t.Fatal(err)
	}
	partialPath, checkpointPath := incrementalArtifactPaths(prepared.OutputBase)
	partial, err := os.ReadFile(partialPath)
	if err != nil || string(partial) != "1\n00:00:01,000 --> 00:00:02,000\nready\n" {
		t.Fatalf("partial = %q, %v; want supplied cues without another cleaning pass", partial, err)
	}
	if err := os.WriteFile(checkpointPath, []byte("checkpoint"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := store.Publish(cues); err != nil {
		t.Fatal(err)
	}
	output, err := os.ReadFile(prepared.Job.OutputPath)
	if err != nil || string(output) != string(partial) {
		t.Fatalf("published = %q, %v; want %q", output, err, partial)
	}
	if err := store.RemoveProgress(); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{partialPath, checkpointPath, resources.TemporaryOutputPath} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("temporary artifact %q remains: %v", path, err)
		}
	}
	if err := store.RemoveProgress(); err != nil {
		t.Fatalf("repeated progress removal: %v", err)
	}
}

func TestResourcesClosePreservesCompletedTranscriptWhenPublicationFails(t *testing.T) {
	prepared, resources := transcriptResources(t, "srt", false)
	store := NewTranscriptStore(prepared, resources)
	cues := []domain.Cue{{Start: time.Second, End: 2 * time.Second, Text: "recoverable"}}
	if err := store.WritePartial(cues); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(prepared.Job.OutputPath, []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := store.Publish(cues); !port.IsOutputExistsError(err) || !strings.Contains(err.Error(), resources.TemporaryOutputPath) {
		t.Fatalf("Publish() error = %v; want collision with recovery path", err)
	}
	if err := resources.Close(); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{resources.AudioPath, resources.ChunkDirectory, resources.reservationPath} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("resource %q remains: %v", path, err)
		}
	}
	content, err := os.ReadFile(resources.TemporaryOutputPath)
	if err != nil || !strings.Contains(string(content), "recoverable") {
		t.Fatalf("recovery transcript = %q, %v", content, err)
	}
	partialPath, _ := incrementalArtifactPaths(prepared.OutputBase)
	if _, err := os.Stat(partialPath); err != nil {
		t.Fatalf("partial transcript removed on failed publication: %v", err)
	}
	if err := resources.Close(); err != nil {
		t.Fatalf("repeated resource cleanup: %v", err)
	}
}

func TestTranscriptStoreForceReplacesOutputAndReportsStorageFailures(t *testing.T) {
	prepared, resources := transcriptResources(t, "txt", true)
	t.Cleanup(func() { _ = resources.Close() })
	store := NewTranscriptStore(prepared, resources)
	if err := os.WriteFile(prepared.Job.OutputPath, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := store.Publish([]domain.Cue{{Text: "replacement"}}); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(prepared.Job.OutputPath)
	if err != nil || string(content) != "replacement\n" {
		t.Fatalf("forced output = %q, %v", content, err)
	}
	invalid := *store
	invalid.format = "invalid"
	if err := invalid.Publish(nil); err == nil {
		t.Fatal("invalid format accepted")
	}
	missing := *store
	missing.outputBase = filepath.Join(t.TempDir(), "missing", "output")
	if err := missing.WritePartial(nil); err == nil || !strings.Contains(err.Error(), "write partial transcript") {
		t.Fatalf("WritePartial() error = %v; want storage context", err)
	}
	_, checkpointPath := incrementalArtifactPaths(prepared.OutputBase)
	if err := os.Mkdir(checkpointPath, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(checkpointPath, "block-removal"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := store.RemoveProgress(); err == nil || !strings.Contains(err.Error(), checkpointPath) {
		t.Fatalf("RemoveProgress() error = %v; want artifact context", err)
	}
}

func fingerprintFiles(t *testing.T, directory string) FingerprintConfig {
	t.Helper()
	paths := make([]string, 7)
	for index, name := range []string{"input.wav", "model.bin", "vad.bin", "whisper-cli", "whisper-vad", "ffmpeg", "ffprobe"} {
		paths[index] = filepath.Join(directory, name)
		if err := os.WriteFile(paths[index], []byte(name), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return FingerprintConfig{InputPath: paths[0], ModelPath: paths[1], VADModelPath: paths[2], WhisperPath: paths[3],
		VADToolPath: paths[4], FFmpegPath: paths[5], FFprobePath: paths[6], Language: "ko"}
}

func transcriptResources(t *testing.T, format string, force bool) (Prepared, *Resources) {
	t.Helper()
	directory := t.TempDir()
	outputBase := filepath.Join(directory, "input")
	prepared := Prepared{OutputBase: outputBase, OutputDirectory: directory, Force: force,
		Job: portout.Job{OutputPath: outputBase + "." + format, Language: "ko", Format: format}}
	resources, err := prepared.CreateResources()
	if err != nil {
		t.Fatal(err)
	}
	return prepared, resources
}
