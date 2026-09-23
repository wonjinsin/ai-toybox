// Package port contains contracts shared by input and output boundaries.
package port

import (
	"errors"
	"fmt"
)

type OutputExistsError struct{ OutputPath string }

func (err *OutputExistsError) Error() string {
	return fmt.Sprintf("output file %q already exists; use --force to overwrite it", err.OutputPath)
}
func NewOutputExistsError(path string) error { return &OutputExistsError{OutputPath: path} }
func IsOutputExistsError(err error) bool {
	var target *OutputExistsError
	return errors.As(err, &target)
}
