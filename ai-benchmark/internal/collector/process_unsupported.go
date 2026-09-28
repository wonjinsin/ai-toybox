//go:build !darwin && !linux

package collector

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func platformSupport() error                     { return fmt.Errorf("collector is unsupported on %s", runtime.GOOS) }
func configureProcess(command *exec.Cmd)         {}
func processExitCode(state *os.ProcessState) int { return state.ExitCode() }
func groupRemains(pid int) (bool, error)         { return false, platformSupport() }
func stopGroup(pid int) error                    { return platformSupport() }
