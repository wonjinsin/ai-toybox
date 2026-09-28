//go:build darwin || linux

package collector

import (
	"os/signal"
	"syscall"
)

func ignoreTermination() { signal.Ignore(syscall.SIGTERM) }
