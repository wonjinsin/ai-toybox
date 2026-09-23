package bootstrap

import (
	"context"
	"io"

	"github.com/wonjinsin/ai-toybox/whisper/internal/adapter/in/cli"
	"github.com/wonjinsin/ai-toybox/whisper/internal/adapter/out/filesystem"
	portin "github.com/wonjinsin/ai-toybox/whisper/internal/port/in"
	"github.com/wonjinsin/ai-toybox/whisper/internal/usecase"
)

func Run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	return cli.Execute(ctx, args, stdout, stderr, func(progress portin.ProgressObserver) portin.BatchRunner {
		return usecase.NewService(filesystem.InputCatalog{}, NewFactory(), progress)
	})
}
