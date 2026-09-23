package filesystem

import (
	"errors"
	"fmt"
	"io"
	"os"
	"syscall"

	"github.com/wonjinsin/ai-toybox/whisper/internal/port"
)

type linkFileFunc func(string, string) error

func publishTranscript(temporaryPath, outputPath string, force bool, linkFile linkFileFunc) error {
	if force {
		if err := os.Rename(temporaryPath, outputPath); err != nil {
			return fmt.Errorf("replace output file %q: %w", outputPath, err)
		}
		return nil
	}
	if err := linkFile(temporaryPath, outputPath); err != nil {
		if os.IsExist(err) {
			return port.NewOutputExistsError(outputPath)
		}
		if errors.Is(err, syscall.EOPNOTSUPP) || errors.Is(err, syscall.ENOTSUP) {
			return copyTranscriptExclusively(temporaryPath, outputPath)
		}
		return fmt.Errorf("publish output file %q: %w", outputPath, err)
	}
	return nil
}

func writeAndPublishTranscript(temporaryPath, outputPath string, content []byte, force bool, linkFile linkFileFunc) error {
	if err := os.WriteFile(temporaryPath, content, 0o644); err != nil {
		return fmt.Errorf("write temporary transcript %q: %w", temporaryPath, err)
	}
	if err := publishTranscript(temporaryPath, outputPath, force, linkFile); err != nil {
		return fmt.Errorf("%w; completed transcript kept at %q", err, temporaryPath)
	}
	if err := os.Remove(temporaryPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("output published to %q, but remove temporary transcript %q: %w", outputPath, temporaryPath, err)
	}
	return nil
}

func copyTranscriptExclusively(temporaryPath, outputPath string) error {
	source, err := os.Open(temporaryPath)
	if err != nil {
		return fmt.Errorf("open temporary transcript %q: %w", temporaryPath, err)
	}
	destination, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		source.Close()
		if os.IsExist(err) {
			return port.NewOutputExistsError(outputPath)
		}
		return fmt.Errorf("create output file %q: %w", outputPath, err)
	}

	_, copyErr := io.Copy(destination, source)
	var syncErr error
	if copyErr == nil {
		syncErr = destination.Sync()
	}
	closeDestinationErr := destination.Close()
	closeSourceErr := source.Close()
	if operationErr := errors.Join(copyErr, syncErr, closeDestinationErr, closeSourceErr); operationErr != nil {
		removeErr := os.Remove(outputPath)
		return fmt.Errorf("copy transcript to %q: %w", outputPath, errors.Join(operationErr, removeErr))
	}
	return nil
}
