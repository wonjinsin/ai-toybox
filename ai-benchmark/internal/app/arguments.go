package app

import (
	"fmt"
	"strings"
)

type options struct {
	command, input, output string
	help                   bool
}

func parse(args []string) (options, error) {
	if len(args) == 0 {
		return options{}, fmt.Errorf("a command is required; use --help")
	}
	if args[0] == "--help" || args[0] == "-h" {
		return options{help: true}, nil
	}
	command := args[0]
	needsOutput := command == "preview" || command == "prepare" || command == "run"
	if !needsOutput && command != "validate" && command != "inspect" {
		return options{}, fmt.Errorf("unknown command %q", command)
	}
	var positionals []string
	var output string
	for index := 1; index < len(args); index++ {
		arg := args[index]
		switch {
		case arg == "--help" || arg == "-h":
			return options{help: true}, nil
		case arg == "--":
			positionals = append(positionals, args[index+1:]...)
			index = len(args)
		case arg == "--output" || strings.HasPrefix(arg, "--output="):
			if !needsOutput || output != "" {
				return options{}, fmt.Errorf("unexpected or repeated --output")
			}
			if arg == "--output" {
				index++
				if index == len(args) {
					return options{}, fmt.Errorf("--output requires a directory")
				}
				output = args[index]
				if strings.HasPrefix(output, "-") && output != "-" {
					return options{}, fmt.Errorf("--output requires a directory; use --output=VALUE for a name beginning with '-'")
				}
			} else {
				output = strings.TrimPrefix(arg, "--output=")
			}
			if output == "" {
				return options{}, fmt.Errorf("--output requires a directory")
			}
		case strings.HasPrefix(arg, "-"):
			return options{}, fmt.Errorf("unknown option %q", arg)
		default:
			positionals = append(positionals, arg)
		}
	}
	if len(positionals) != 1 || positionals[0] == "" {
		return options{}, fmt.Errorf("%s requires exactly one input path", command)
	}
	if needsOutput && output == "" {
		return options{}, fmt.Errorf("%s requires --output", command)
	}
	return options{command: command, input: positionals[0], output: output}, nil
}
