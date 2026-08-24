package app

import (
	"errors"
	"fmt"
)

type outputExistsError struct {
	outputPath string
}

func (err *outputExistsError) Error() string {
	return fmt.Sprintf("output file %q already exists; use --force to overwrite it", err.outputPath)
}

func newOutputExistsError(outputPath string) error {
	return &outputExistsError{outputPath: outputPath}
}

func isOutputExistsError(err error) bool {
	var target *outputExistsError
	return errors.As(err, &target)
}
