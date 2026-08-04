package out

import "context"

// AIRunner runs a prompt through an AI CLI backend and returns its text output.
type AIRunner interface {
	Run(ctx context.Context, prompt string) (string, error)
}
