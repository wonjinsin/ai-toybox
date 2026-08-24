package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

type TranscriptionResult struct {
	InputPath  string
	OutputPath string
	Err        error
}

type transcribeFunc func(context.Context, Options) (string, error)

func transcribeAll(ctx context.Context, options Options, transcribe transcribeFunc) ([]TranscriptionResult, error) {
	if options.Parallel < 1 {
		return nil, errors.New("parallel must be at least 1")
	}
	inputPaths, err := discoverInputPaths(options.InputPath)
	if err != nil {
		return nil, err
	}
	if err := validateUniqueOutputNames(inputPaths, options.Format); err != nil {
		return nil, err
	}

	workerCount := min(options.Parallel, len(inputPaths))
	tasks := make(chan string, len(inputPaths))
	results := make(chan TranscriptionResult, len(inputPaths))
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
					results <- TranscriptionResult{InputPath: inputPath, Err: err}
					continue
				}
				taskOptions := options
				taskOptions.InputPath = inputPath
				outputPath, err := transcribe(ctx, taskOptions)
				results <- TranscriptionResult{
					InputPath:  inputPath,
					OutputPath: outputPath,
					Err:        err,
				}
			}
		}()
	}
	workers.Wait()
	close(results)

	orderedResults := make([]TranscriptionResult, 0, len(inputPaths))
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

var supportedMediaExtensions = map[string]struct{}{
	".3g2": {}, ".3gp": {}, ".aac": {}, ".ac3": {}, ".aif": {},
	".aiff": {}, ".alac": {}, ".amr": {}, ".ape": {}, ".au": {},
	".avi": {}, ".caf": {}, ".dts": {}, ".flac": {}, ".m2ts": {},
	".m4a": {}, ".m4v": {}, ".mka": {}, ".mkv": {}, ".mov": {},
	".mp3": {}, ".mp4": {}, ".mpeg": {}, ".mpg": {}, ".mts": {},
	".mxf": {}, ".oga": {}, ".ogg": {}, ".ogv": {}, ".opus": {},
	".ts": {}, ".vob": {}, ".wav": {}, ".webm": {}, ".wma": {},
	".wmv": {},
}

func discoverInputPaths(inputPath string) ([]string, error) {
	info, err := os.Stat(inputPath)
	if err != nil {
		return nil, err
	}
	if info.Mode().IsRegular() {
		return []string{inputPath}, nil
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("input path %q is not a regular file or directory", inputPath)
	}

	entries, err := os.ReadDir(inputPath)
	if err != nil {
		return nil, err
	}

	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.Type().IsRegular() && !strings.HasPrefix(entry.Name(), "._") {
			if _, supported := supportedMediaExtensions[strings.ToLower(filepath.Ext(entry.Name()))]; supported {
				paths = append(paths, filepath.Join(inputPath, entry.Name()))
			}
		}
	}
	sort.Strings(paths)
	if len(paths) == 0 {
		return nil, fmt.Errorf("no supported media files found in %q", inputPath)
	}
	return paths, nil
}
