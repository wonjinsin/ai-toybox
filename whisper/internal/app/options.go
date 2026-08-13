package app

import "errors"
import "flag"
import "fmt"
import "io"

type Options struct {
	InputPath string
	Language  string
	Format    string
	ModelPath string
	OutputDir string
	Force     bool
	Parallel  int
}

func ParseOptions(args []string) (Options, error) {
	flags := flag.NewFlagSet("whisper-local", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	language := flags.String("language", "auto", "transcription language")
	format := flags.String("format", "txt", "output format")
	modelPath := flags.String("model", "", "Whisper model path")
	outputDir := flags.String("output", "", "output directory")
	force := flags.Bool("force", false, "overwrite an existing transcript")
	parallel := flags.Int("parallel", 1, "maximum number of concurrent transcriptions")

	if err := flags.Parse(args); err != nil {
		return Options{}, err
	}
	if flags.NArg() != 1 {
		return Options{}, errors.New("input file is required")
	}
	if *parallel < 1 {
		return Options{}, errors.New("parallel must be at least 1")
	}
	if err := validateLanguage(*language); err != nil {
		return Options{}, err
	}
	if err := validateFormat(*format); err != nil {
		return Options{}, err
	}

	return Options{
		InputPath: flags.Arg(0),
		Language:  *language,
		Format:    *format,
		ModelPath: *modelPath,
		OutputDir: *outputDir,
		Force:     *force,
		Parallel:  *parallel,
	}, nil
}

func validateLanguage(language string) error {
	switch language {
	case "auto", "en", "ja", "ko", "zh":
		return nil
	default:
		return fmt.Errorf("unsupported language %q: use auto, ko, ja, zh, or en", language)
	}
}

func validateFormat(format string) error {
	switch format {
	case "srt", "txt", "vtt":
		return nil
	default:
		return fmt.Errorf("unsupported format %q: use txt, srt, or vtt", format)
	}
}
