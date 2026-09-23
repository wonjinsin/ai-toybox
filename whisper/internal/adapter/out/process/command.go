package process

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func Run(ctx context.Context, executable string, args []string) error {
	_, err := RunCaptureStderr(ctx, executable, args)
	return err
}

func RunCaptureStderr(ctx context.Context, executable string, args []string) (string, error) {
	var stderr bytes.Buffer
	command := exec.CommandContext(ctx, executable, args...)
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			return "", err
		}
		return "", fmt.Errorf("%w: %s", err, message)
	}
	return stderr.String(), nil
}
