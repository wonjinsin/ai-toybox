package usecase

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	portin "github.com/wonjinsin/ai-toybox/whisper/internal/port/in"
)

type transcribeFunc func(context.Context, portin.Request) (string, error)

func runBatch(ctx context.Context, options portin.Request, inputPaths []string, transcribe transcribeFunc) ([]portin.Result, error) {
	if options.Parallel < 1 {
		return nil, errors.New("parallel must be at least 1")
	}
	if err := validateUniqueOutputNames(inputPaths, options.Format); err != nil {
		return nil, err
	}

	workerCount := min(options.Parallel, len(inputPaths))
	tasks := make(chan string, len(inputPaths))
	results := make(chan portin.Result, len(inputPaths))
	for _, inputPath := range inputPaths {
		tasks <- inputPath
	}
	close(tasks)

	var workers sync.WaitGroup
	workers.Add(workerCount)
	for range workerCount {
		go func() {
			defer workers.Done()
			for inputPath := range tasks {
				if err := ctx.Err(); err != nil {
					results <- portin.Result{InputPath: inputPath, Err: err}
					continue
				}
				taskRequest := options
				taskRequest.InputPath = inputPath
				outputPath, err := transcribe(ctx, taskRequest)
				results <- portin.Result{
					InputPath:  inputPath,
					OutputPath: outputPath,
					Err:        err,
				}
			}
		}()
	}
	workers.Wait()
	close(results)

	orderedResults := make([]portin.Result, 0, len(inputPaths))
	for result := range results {
		orderedResults = append(orderedResults, result)
	}
	sort.Slice(orderedResults, func(first, second int) bool {
		return orderedResults[first].InputPath < orderedResults[second].InputPath
	})
	return orderedResults, nil
}

func validateUniqueOutputNames(inputPaths []string, format string) error {
	seen := make(map[string]string, len(inputPaths))
	for _, inputPath := range inputPaths {
		baseName := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
		outputName := strings.ToLower(baseName + "." + format)
		if previousPath, exists := seen[outputName]; exists {
			return fmt.Errorf("inputs %q and %q produce the same output %q", previousPath, inputPath, outputName)
		}
		seen[outputName] = inputPath
	}
	return nil
}

func (service *Service) Run(ctx context.Context, request portin.Request) (portin.BatchResult, error) {
	if err := portin.ValidateRequest(request); err != nil {
		return portin.BatchResult{}, err
	}
	if request.Parallel < 1 {
		return portin.BatchResult{}, errors.New("parallel must be at least 1")
	}
	inputs, err := service.catalog.Discover(request.InputPath)
	if err != nil {
		return portin.BatchResult{}, err
	}
	results, err := runBatch(ctx, request, inputs.Paths, service.Transcribe)
	return portin.BatchResult{Results: results, Directory: inputs.Directory}, err
}
