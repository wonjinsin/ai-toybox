// Package collector records subprocess evidence after a caller validates its isolation profile.
package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Plan supplies an invocation already validated by the caller's isolation adapter.
type Plan struct {
	Argv       []string
	Prompt     []byte
	Experiment []byte
	Cwd        string
	Timeout    time.Duration
}

// Execute preserves one attempt without modifying argv or enforcing isolation.
// Argv must already export the final response to outputDir/response.txt.
// Context cancellation records an interrupted attempt with reason context_cancelled.
// Process cleanup covers descendants remaining in the original POSIX process group.
func Execute(ctx context.Context, plan Plan, outputDir string) (map[string]any, error) {
	if ctx == nil || plan.Timeout <= 0 || len(plan.Argv) == 0 || plan.Argv[0] == "" || plan.Cwd == "" || outputDir == "" {
		return nil, fmt.Errorf("collector requires context, positive timeout, argv, workspace, and archive")
	}
	if err := platformSupport(); err != nil {
		return nil, err
	}
	workspace, err := resolvePath(plan.Cwd)
	if err != nil {
		return nil, err
	}
	archive, err := resolvePath(outputDir)
	if err != nil {
		return nil, err
	}
	overlap, err := pathsOverlap(workspace, archive)
	if err != nil {
		return nil, err
	}
	if overlap {
		return nil, fmt.Errorf("execution workspace and archive must not overlap")
	}
	outputDir = archive
	if err := os.MkdirAll(filepath.Dir(outputDir), 0700); err != nil {
		return nil, err
	}
	if err := os.Mkdir(outputDir, 0700); err != nil {
		return nil, err
	}
	for name, data := range map[string][]byte{"prompt.md": plan.Prompt, "experiment.json": plan.Experiment} {
		if err := os.WriteFile(filepath.Join(outputDir, name), data, 0600); err != nil {
			return nil, err
		}
	}
	input, err := os.Open(filepath.Join(outputDir, "prompt.md"))
	if err != nil {
		return nil, err
	}
	defer input.Close()
	events, err := os.Create(filepath.Join(outputDir, "events.jsonl"))
	if err != nil {
		return nil, err
	}
	defer events.Close()
	errors, err := os.Create(filepath.Join(outputDir, "errors.log"))
	if err != nil {
		return nil, err
	}
	defer errors.Close()
	command := exec.Command(plan.Argv[0], plan.Argv[1:]...)
	command.Dir, command.Stdin, command.Stdout, command.Stderr = workspace, input, events, errors
	configureProcess(command)
	started := time.Now()
	result := runCommand(ctx, command, plan.Timeout)
	info, statErr := os.Lstat(filepath.Join(outputDir, "response.txt"))
	responseIssue := ""
	if os.IsNotExist(statErr) {
		responseIssue = "missing_final_response"
	} else if statErr != nil {
		responseIssue = "response_inspection_error"
		if result.detail == nil {
			result = outcome{result.exitCode, result.reason, statErr, result.spawned}
		}
	} else if !info.Mode().IsRegular() {
		responseIssue = "non_regular_final_response"
	}
	record := makeRecord(started, result, responseIssue)
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return record, err
	}
	return record, os.WriteFile(filepath.Join(outputDir, "run.json"), append(data, '\n'), 0600)
}

type outcome struct {
	exitCode any
	reason   string
	detail   error
	spawned  bool
}

func runCommand(ctx context.Context, command *exec.Cmd, timeout time.Duration) outcome {
	if ctx.Err() != nil {
		return outcome{nil, "context_cancelled", nil, false}
	}
	if err := command.Start(); err != nil {
		return outcome{nil, "spawn_error", err, false}
	}
	done := make(chan error, 1)
	go func() { done <- command.Wait() }()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case waitErr := <-done:
		remains, err := groupRemains(command.Process.Pid)
		if err != nil {
			return outcome{processExitCode(command.ProcessState), "process_error", err, true}
		}
		if remains {
			return outcome{processExitCode(command.ProcessState), "lingering_processes", stopGroup(command.Process.Pid), true}
		}
		if waitErr != nil {
			return outcome{processExitCode(command.ProcessState), "nonzero_exit", nil, true}
		}
		return outcome{0, "", nil, true}
	case <-timer.C:
		if err := stopGroup(command.Process.Pid); err != nil {
			return outcome{nil, "timeout", err, true}
		}
		<-done
		return outcome{processExitCode(command.ProcessState), "timeout", nil, true}
	case <-ctx.Done():
		if err := stopGroup(command.Process.Pid); err != nil {
			return outcome{nil, "context_cancelled", err, true}
		}
		<-done
		return outcome{processExitCode(command.ProcessState), "context_cancelled", nil, true}
	}
}

func makeRecord(started time.Time, result outcome, responseIssue string) map[string]any {
	status, responseStatus, streamStatus := "completed", "captured", "captured"
	failureReason := result.reason
	if failureReason == "" && responseIssue != "" {
		failureReason = responseIssue
	}
	if failureReason != "" {
		status, responseStatus = "failed", "partial"
	}
	if failureReason == "context_cancelled" {
		status = "interrupted"
	}
	responseReason := failureReason
	if responseIssue != "" {
		responseStatus, responseReason = "unavailable", responseIssue
	}
	streamReason := result.reason
	if streamReason == "nonzero_exit" {
		streamReason = ""
	}
	if !result.spawned {
		streamStatus = "unavailable"
	} else if streamReason != "" {
		streamStatus = "partial"
	}
	var errorDetail any
	if result.detail != nil {
		errorDetail = result.detail.Error()
	}
	return map[string]any{
		"record_status": "finalized",
		"execution": map[string]any{
			"status": status, "exit_code": result.exitCode, "reason": nullable(failureReason), "error": errorDetail,
			"started_at": started.UTC().Format(time.RFC3339Nano), "finished_at": time.Now().UTC().Format(time.RFC3339Nano),
			"duration_ms":   float64(time.Since(started)) / float64(time.Millisecond),
			"timing_method": "monotonic_clock", "timing_scope": "cli_process",
		},
		"capture": map[string]any{
			"response": capture("response.txt", responseStatus, nullable(responseReason)),
			"stderr":   capture("errors.log", streamStatus, nullable(streamReason)),
			"events":   capture("events.jsonl", streamStatus, nullable(streamReason)),
		},
	}
}

func capture(path, status string, reason any) map[string]any {
	return map[string]any{"path": path, "status": status, "reason": reason}
}

func nullable(value string) any {
	if value == "" {
		return nil
	}
	return value
}
