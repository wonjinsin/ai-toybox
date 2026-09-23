package filesystem

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wonjinsin/ai-toybox/whisper/internal/port"
	portout "github.com/wonjinsin/ai-toybox/whisper/internal/port/out"
)

// Prepared contains filesystem configuration resolved before engine discovery.
type Prepared struct {
	Job             portout.Job
	ModelPath       string
	VADModelPath    string
	OutputBase      string
	OutputDirectory string
	Force           bool
}

func Prepare(request portout.SessionRequest) (Prepared, error) {
	inputPath, err := requireRegularFile(request.InputPath, "input")
	if err != nil {
		return Prepared{}, err
	}
	outputDirectory := request.OutputDir
	if outputDirectory == "" {
		outputDirectory = filepath.Dir(inputPath)
	}
	outputDirectory, err = filepath.Abs(outputDirectory)
	if err != nil {
		return Prepared{}, fmt.Errorf("resolve output directory: %w", err)
	}
	baseName := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	outputBase := filepath.Join(outputDirectory, baseName)
	outputPath := outputBase + "." + request.Format
	if !request.Force {
		if _, err := os.Stat(outputPath); err == nil {
			return Prepared{}, port.NewOutputExistsError(outputPath)
		} else if !os.IsNotExist(err) {
			return Prepared{}, fmt.Errorf("inspect output file %q: %w", outputPath, err)
		}
	}
	modelPath, err := resolveModelPath(request.ModelPath)
	if err != nil {
		return Prepared{}, err
	}
	vadModelPath, err := resolveRequiredVADModelPath(request.VADModelPath, modelPath)
	if err != nil {
		return Prepared{}, err
	}
	corrections, err := loadCorrections(request.CorrectionsPath)
	if err != nil {
		return Prepared{}, err
	}
	return Prepared{
		Job: portout.Job{
			InputPath: inputPath, OutputPath: outputPath,
			Language: request.Language, Format: request.Format, Corrections: corrections,
		},
		ModelPath: modelPath, VADModelPath: vadModelPath,
		OutputBase: outputBase, OutputDirectory: outputDirectory, Force: request.Force,
	}, nil
}

// Resources owns disposable audio resources and the output reservation. A completed
// temporary transcript belongs to the caller for recovery after publication fails.
type Resources struct {
	AudioPath           string
	ChunkDirectory      string
	TemporaryOutputPath string
	reservationPath     string
}

func (prepared Prepared) CreateOutputDirectory() error {
	if err := os.MkdirAll(prepared.OutputDirectory, 0o755); err != nil {
		return fmt.Errorf("create output directory %q: %w", prepared.OutputDirectory, err)
	}
	return nil
}

func (prepared Prepared) CreateResources() (_ *Resources, returnErr error) {
	reservation, err := os.CreateTemp(prepared.OutputDirectory, ".whisper-local-output-*")
	if err != nil {
		return nil, fmt.Errorf("reserve temporary transcript: %w", err)
	}
	reservationPath := reservation.Name()
	if err := reservation.Close(); err != nil {
		return nil, errors.Join(fmt.Errorf("close temporary transcript reservation: %w", err), removeResource(reservationPath))
	}
	defer func() {
		if returnErr != nil {
			returnErr = errors.Join(returnErr, removeResource(reservationPath))
		}
	}()

	audio, err := os.CreateTemp("", "whisper-local-*.wav")
	if err != nil {
		return nil, fmt.Errorf("create temporary audio file: %w", err)
	}
	audioPath := audio.Name()
	defer func() {
		if returnErr != nil {
			returnErr = errors.Join(returnErr, removeResource(audioPath))
		}
	}()
	if err := audio.Close(); err != nil {
		return nil, fmt.Errorf("close temporary audio file: %w", err)
	}
	chunkDirectory, err := os.MkdirTemp("", "whisper-local-chunks-*")
	if err != nil {
		return nil, fmt.Errorf("create temporary chunk directory: %w", err)
	}
	return &Resources{
		AudioPath: audioPath, ChunkDirectory: chunkDirectory,
		TemporaryOutputPath: reservationPath + "." + prepared.Job.Format,
		reservationPath:     reservationPath,
	}, nil
}

func (resources *Resources) Close() error {
	var chunkErr error
	if err := os.RemoveAll(resources.ChunkDirectory); err != nil {
		chunkErr = fmt.Errorf("remove temporary chunk directory %q: %w", resources.ChunkDirectory, err)
	}
	return errors.Join(chunkErr, removeResource(resources.AudioPath), removeResource(resources.reservationPath))
}

func removeResource(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove temporary resource %q: %w", path, err)
	}
	return nil
}
