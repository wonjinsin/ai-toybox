package main

import (
	"fmt"
	"os"

	"github.com/wonjinsin/ledger/internal/adapter/in/cli"
)

func main() {
	if err := cli.NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
