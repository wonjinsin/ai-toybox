package runner

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"
)

const usage = `Usage: ai-benchmark run [--output DIRECTORY] EXPERIMENT

Read one experiment file and snapshot its prompt and settings.
Output defaults to a new timestamped directory under runs/.
Strict Skills/Tools enforcement is required. Codex execution is currently blocked.
`

const blockedReason = "cannot enforce the declared Skills/Tools/MCP/Agents allowlists with the current Codex CLI integration; no model was started"

func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Fprint(stdout, usage)
		return 0
	}
	if len(args) == 0 || args[0] != "run" {
		fmt.Fprint(stderr, usage)
		return 2
	}
	flags := flag.NewFlagSet("run", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() { fmt.Fprint(stdout, usage) }
	output := flags.String("output", "", "new run directory")
	for index, arg := range args {
		if (arg == "--output" || arg == "-output") && index+1 < len(args) && strings.HasPrefix(args[index+1], "-") {
			fmt.Fprintln(stderr, "error: --output requires a directory; use --output=VALUE for a name starting with '-'")
			return 2
		}
	}
	if err := flags.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if flags.NArg() != 1 {
		fmt.Fprint(stderr, usage)
		return 2
	}
	in, err := loadInput(flags.Arg(0))
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 2
	}
	archive := *output
	if archive == "" {
		archive = filepath.Join("runs", time.Now().UTC().Format("20060102T150405.000000000Z"))
	}
	archive, err = filepath.Abs(archive)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 2
	}
	record, err := prepareRecord(in, archive)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 2
	}
	// Never replace an unverified capability policy with ordinary CLI defaults.
	record = overlay(record, map[string]any{
		"cli":       map[string]any{"name": "codex", "version": nil, "mode": "non_interactive"},
		"preflight": map[string]any{"status": "blocked", "reason": blockedReason},
	})
	if err := saveRecord(archive, record); err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 2
	}
	fmt.Fprintf(stdout, "Inputs saved: %s\n", archive)
	fmt.Fprintf(stderr, "blocked: %s\n", blockedReason)
	return 3
}
