//go:build darwin || linux

package collector

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
	"time"
)

func platformSupport() error { return nil }

func configureProcess(command *exec.Cmd) {
	command.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}

func processExitCode(state *os.ProcessState) int {
	if status, ok := state.Sys().(syscall.WaitStatus); ok && status.Signaled() {
		return -int(status.Signal())
	}
	return state.ExitCode()
}

func groupRemains(pid int) (bool, error) {
	err := syscall.Kill(-pid, 0)
	if errors.Is(err, syscall.ESRCH) {
		return false, nil
	}
	if err == nil || errors.Is(err, syscall.EPERM) {
		return true, nil
	}
	return true, err
}

func waitGroupExit(pid int, duration time.Duration) (bool, error) {
	deadline := time.Now().Add(duration)
	for {
		remains, err := groupRemains(pid)
		if err != nil || !remains {
			return !remains, err
		}
		if !time.Now().Before(deadline) {
			return false, nil
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func signalGroup(pid int, signal syscall.Signal) error {
	err := syscall.Kill(-pid, signal)
	if errors.Is(err, syscall.ESRCH) {
		return nil
	}
	if errors.Is(err, syscall.EPERM) {
		gone, checkErr := waitGroupExit(pid, 50*time.Millisecond)
		if checkErr == nil && gone {
			return nil
		}
	}
	return err
}

func stopGroup(pid int) error {
	if err := signalGroup(pid, syscall.SIGTERM); err != nil {
		return err
	}
	gone, err := waitGroupExit(pid, 250*time.Millisecond)
	if err != nil || gone {
		return err
	}
	if err := signalGroup(pid, syscall.SIGKILL); err != nil {
		return err
	}
	_, err = waitGroupExit(pid, 50*time.Millisecond)
	return err
}
