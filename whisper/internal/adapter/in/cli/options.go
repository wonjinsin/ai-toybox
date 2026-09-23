package cli

import (
	"errors"
	"flag"
	"io"

	portin "github.com/wonjinsin/ai-toybox/whisper/internal/port/in"
)

func ParseOptions(args []string) (portin.Request, error) {
	flags := flag.NewFlagSet("whisper-local", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	language := flags.String("language", "auto", "transcription language")
	format := flags.String("format", "txt", "output format")
	modelPath := flags.String("model", "", "Whisper model path")
	vadModelPath := flags.String("vad-model", "", "Silero VAD model path")
	correctionsPath := flags.String("corrections", "", "JSON text corrections path")
	outputDir := flags.String("output", "", "output directory")
	force := flags.Bool("force", false, "overwrite an existing transcript")
	parallel := flags.Int("parallel", 1, "maximum number of concurrent transcriptions")

	if err := flags.Parse(args); err != nil {
		return portin.Request{}, err
	}
	if flags.NArg() != 1 {
		return portin.Request{}, errors.New("input file is required")
	}
	if *parallel < 1 {
		return portin.Request{}, errors.New("parallel must be at least 1")
	}
	if err := portin.ValidateLanguage(*language); err != nil {
		return portin.Request{}, err
	}
	if err := portin.ValidateFormat(*format); err != nil {
		return portin.Request{}, err
	}

	return portin.Request{
		InputPath:       flags.Arg(0),
		Language:        *language,
		Format:          *format,
		ModelPath:       *modelPath,
		VADModelPath:    *vadModelPath,
		CorrectionsPath: *correctionsPath,
		OutputDir:       *outputDir,
		Force:           *force,
		Parallel:        *parallel,
	}, nil
}
