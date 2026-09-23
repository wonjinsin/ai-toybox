// Package in defines the operations and responses exposed to driving adapters.
package in

import "context"

type Request struct {
	InputPath       string
	Language        string
	Format          string
	ModelPath       string
	VADModelPath    string
	CorrectionsPath string
	OutputDir       string
	Force           bool
	Parallel        int
}

type Result struct {
	InputPath  string
	OutputPath string
	Err        error
}

type BatchResult struct {
	Results   []Result
	Directory bool
}

type Transcriber interface {
	Transcribe(context.Context, Request) (string, error)
}

type BatchRunner interface {
	Run(context.Context, Request) (BatchResult, error)
}
