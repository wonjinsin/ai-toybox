package in

import "fmt"

func ValidateRequest(request Request) error {
	if err := ValidateLanguage(request.Language); err != nil {
		return err
	}
	return ValidateFormat(request.Format)
}

func ValidateLanguage(language string) error {
	switch language {
	case "auto", "en", "ja", "ko", "zh":
		return nil
	default:
		return fmt.Errorf("unsupported language %q: use auto, ko, ja, zh, or en", language)
	}
}

func ValidateFormat(format string) error {
	switch format {
	case "srt", "txt", "vtt":
		return nil
	default:
		return fmt.Errorf("unsupported format %q: use txt, srt, or vtt", format)
	}
}
