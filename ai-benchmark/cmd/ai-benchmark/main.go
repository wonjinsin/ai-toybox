package main

import (
	"ai-benchmark/internal/runner"
	"os"
)

func main() { os.Exit(runner.Run(os.Args[1:], os.Stdout, os.Stderr)) }
