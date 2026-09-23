package integration

import (
	"context"
	"io"
	"path/filepath"

	"github.com/wonjinsin/ai-toybox/whisper/internal/adapter/out/whisper"
	"github.com/wonjinsin/ai-toybox/whisper/internal/bootstrap"
	portin "github.com/wonjinsin/ai-toybox/whisper/internal/port/in"
	"github.com/wonjinsin/ai-toybox/whisper/internal/usecase"
)

func execute(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	return bootstrap.Run(ctx, args, stdout, stderr)
}

func transcribe(ctx context.Context, request portin.Request) (string, error) {
	return usecase.NewService(nil, bootstrap.NewFactory(), nil).Transcribe(ctx, request)
}

func transcribeWithRunner(ctx context.Context, request portin.Request, progress portin.ProgressObserver, runner *whisper.CommandRunner) (string, error) {
	return usecase.NewService(nil, bootstrap.NewFactoryWithRunner(runner), progress).Transcribe(ctx, request)
}

func artifactPaths(outputBase string) (string, string) {
	return outputBase + ".partial.srt", filepath.Join(filepath.Dir(outputBase), "."+filepath.Base(outputBase)+".srt.whisper-local-checkpoint.json")
}
