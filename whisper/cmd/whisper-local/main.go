package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

import "github.com/wonjinsin/ai-toybox/whisper/internal/app"

func main() {
	ctx, stop := signalContext()
	defer stop()

	os.Exit(app.Execute(ctx, os.Args[1:], os.Stdout, os.Stderr))
}

func signalContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
}
