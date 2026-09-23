// Package bootstrap assembles use cases and local adapters.
package bootstrap

import (
	"context"
	"errors"
	"os/exec"

	"github.com/wonjinsin/ai-toybox/whisper/internal/adapter/out/ffmpeg"
	"github.com/wonjinsin/ai-toybox/whisper/internal/adapter/out/filesystem"
	"github.com/wonjinsin/ai-toybox/whisper/internal/adapter/out/whisper"
	portout "github.com/wonjinsin/ai-toybox/whisper/internal/port/out"
)

type Factory struct {
	runner *whisper.CommandRunner
}

// NewFactory shares one inference slot across all sessions and file workers.
func NewFactory() *Factory {
	return NewFactoryWithRunner(nil)
}

func NewFactoryWithRunner(runner *whisper.CommandRunner) *Factory {
	if runner == nil {
		runner = whisper.NewCommandRunner(1, nil)
	}
	return &Factory{runner: runner}
}

func (factory *Factory) Open(_ context.Context, request portout.SessionRequest) (portout.Session, error) {
	prepared, err := filesystem.Prepare(request)
	if err != nil {
		return portout.Session{}, err
	}
	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		return portout.Session{}, errors.New("ffmpeg not found: install it with 'brew install ffmpeg'")
	}
	ffprobePath, err := exec.LookPath("ffprobe")
	if err != nil {
		return portout.Session{}, errors.New("ffprobe not found: install it with 'brew install ffmpeg'")
	}
	whisperPath, err := exec.LookPath("whisper-cli")
	if err != nil {
		return portout.Session{}, errors.New("whisper-cli not found: install it with 'brew install whisper-cpp'")
	}
	vadToolPath, err := exec.LookPath("whisper-vad-speech-segments")
	if err != nil {
		return portout.Session{}, errors.New("whisper-vad-speech-segments not found: install it with 'brew install whisper-cpp'")
	}
	if err := prepared.CreateOutputDirectory(); err != nil {
		return portout.Session{}, err
	}
	var checkpoints portout.CheckpointStore
	if request.Format == "srt" {
		checkpoints, err = filesystem.NewCheckpointStore(prepared.OutputBase, filesystem.FingerprintConfig{
			InputPath: prepared.Job.InputPath, ModelPath: prepared.ModelPath, VADModelPath: prepared.VADModelPath,
			WhisperPath: whisperPath, VADToolPath: vadToolPath, FFmpegPath: ffmpegPath,
			FFprobePath: ffprobePath, Language: request.Language,
		})
		if err != nil {
			return portout.Session{}, err
		}
	}
	resources, err := prepared.CreateResources()
	if err != nil {
		return portout.Session{}, err
	}
	return portout.Session{
		Job:         prepared.Job,
		Audio:       ffmpeg.New(ffmpegPath, ffprobePath, prepared.Job.InputPath, resources.AudioPath, resources.ChunkDirectory),
		Detector:    whisper.NewDetector(ffmpegPath, vadToolPath, resources.AudioPath, prepared.VADModelPath, factory.runner.Execute),
		Recognizer:  whisper.NewRecognizer(whisperPath, prepared.ModelPath, request.Language, factory.runner),
		Checkpoints: checkpoints, Transcripts: filesystem.NewTranscriptStore(prepared, resources), Close: resources.Close,
	}, nil
}
