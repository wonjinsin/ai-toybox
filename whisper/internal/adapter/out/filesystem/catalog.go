package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	portout "github.com/wonjinsin/ai-toybox/whisper/internal/port/out"
)

type InputCatalog struct{}

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

func (InputCatalog) Discover(inputPath string) (portout.InputSet, error) {
	info, err := os.Stat(inputPath)
	if err != nil {
		return portout.InputSet{}, err
	}
	if info.Mode().IsRegular() {
		return portout.InputSet{Paths: []string{inputPath}}, nil
	}
	if !info.IsDir() {
		return portout.InputSet{}, fmt.Errorf("input path %q is not a regular file or directory", inputPath)
	}

	entries, err := os.ReadDir(inputPath)
	if err != nil {
		return portout.InputSet{}, err
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
		return portout.InputSet{}, fmt.Errorf("no supported media files found in %q", inputPath)
	}
	return portout.InputSet{Paths: paths, Directory: true}, nil
}
