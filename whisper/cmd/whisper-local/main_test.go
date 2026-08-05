package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"testing"
)

func TestSignalContextHandlesSIGTERM(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("SIGTERM behavior is Unix-specific")
	}

	command := exec.Command(os.Args[0], "-test.run=TestSignalContextHelper")
	command.Env = append(os.Environ(), "WHISPER_SIGNAL_HELPER=1")
	stdout, err := command.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}

	scanner := bufio.NewScanner(stdout)
	if !scanner.Scan() || scanner.Text() != "ready" {
		t.Fatalf("helper did not become ready: %q", scanner.Text())
	}
	if err := command.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	if !scanner.Scan() || scanner.Text() != "canceled" {
		t.Fatalf("context was not canceled: %q", scanner.Text())
	}
	if err := command.Wait(); err != nil {
		t.Fatalf("helper exit error = %v", err)
	}
}

func TestSignalContextHelper(t *testing.T) {
	if os.Getenv("WHISPER_SIGNAL_HELPER") != "1" {
		return
	}

	ctx, stop := signalContext()
	defer stop()
	fmt.Println("ready")
	<-ctx.Done()
	fmt.Println("canceled")
}
