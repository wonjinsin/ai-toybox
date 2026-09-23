package main

import (
	"context"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/wonjinsin/ai-toybox/whisper/internal/bootstrap"
)

func main() {
	ctx, stop := signalContext()
	defer stop()

	os.Exit(run(ctx, os.Args[1:], os.Stdout, os.Stderr))
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	return bootstrap.Run(ctx, args, stdout, stderr)
}

func signalContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
}
