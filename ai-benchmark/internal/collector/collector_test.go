package collector

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCollectorHelper(t *testing.T) {
	marker := -1
	for i, arg := range os.Args {
		if arg == "--" {
			marker = i
			break
		}
	}
	if marker < 0 {
		t.Skip("subprocess helper")
	}
	args := os.Args[marker+1:]
	if args[0] == "symlink" {
		if err := os.Symlink(args[2], args[1]); err != nil {
			os.Exit(96)
		}
		os.Exit(0)
	}
	if args[0] == "ignore" {
		ignoreTermination()
		os.WriteFile(args[2], []byte("ready"), 0600)
		helperWaitForFile(args[2] + ".release")
		os.WriteFile(args[1], []byte("late modification"), 0600)
		os.Exit(0)
	}
	if args[0] == "tree-ignore" || args[0] == "parent-ignore" {
		ignoreTermination()
		os.WriteFile(args[1], []byte("original response"), 0600)
		child := exec.Command(os.Args[0], "-test.run=^TestCollectorHelper$", "--", "ignore", args[1], args[2])
		child.Stdout, child.Stderr = os.Stdout, os.Stderr
		if err := child.Start(); err != nil {
			os.Exit(94)
		}
		helperWaitForFile(args[2])
		if args[0] == "tree-ignore" {
			time.Sleep(2 * time.Second)
		}
		os.Exit(7)
	}
	if args[0] == "late" {
		os.WriteFile(args[2]+".ready", []byte("ready"), 0600)
		helperWaitForFile(args[2])
		os.WriteFile(args[1], []byte("late modification"), 0600)
		os.Stdout.Write([]byte("late event\n"))
		os.Exit(0)
	}
	if args[0] == "parent" {
		if err := os.WriteFile(args[1], []byte("original response"), 0600); err != nil {
			os.Exit(93)
		}
		child := exec.Command(os.Args[0], "-test.run=^TestCollectorHelper$", "--", "late", args[1], args[2])
		child.Stdout, child.Stderr = os.Stdout, os.Stderr
		if err := child.Start(); err != nil {
			os.Exit(94)
		}
		helperWaitForFile(args[2] + ".ready")
		os.Exit(0)
	}
	if args[0] == "sleep" {
		os.Stdout.Write([]byte("started\n"))
		os.Stderr.Write([]byte("diagnostic\n"))
		time.Sleep(2 * time.Second)
		os.Exit(0)
	}
	if args[0] != "success" && args[0] != "nonzero" && args[0] != "missing" {
		os.Exit(90)
	}
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		os.Exit(91)
	}
	if err := os.WriteFile(args[1], input, 0600); args[0] != "missing" && err != nil {
		os.Exit(92)
	}
	if args[0] == "missing" {
		os.Remove(args[1])
	}
	os.Stdout.Write([]byte("raw\x00events\n"))
	os.Stderr.Write([]byte("raw\xffdiagnostic\n"))
	if args[0] == "nonzero" {
		os.Exit(7)
	}
	os.Exit(0)
}

func helperWaitForFile(path string) {
	deadline := time.Now().Add(5 * time.Second)
	for {
		if _, err := os.Stat(path); err == nil {
			return
		}
		if !time.Now().Before(deadline) {
			os.Exit(95)
		}
		time.Sleep(time.Millisecond)
	}
}

func fixture(t *testing.T, mode string) (Plan, string) {
	t.Helper()
	t.Setenv("GORACE", os.Getenv("GORACE")+" atexit_sleep_ms=0")
	root := t.TempDir()
	workspace := filepath.Join(root, "workspace")
	if err := os.Mkdir(workspace, 0700); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(root, "archive", "attempt")
	return Plan{
		Argv:   []string{os.Args[0], "-test.run=^TestCollectorHelper$", "--", mode, filepath.Join(output, "response.txt")},
		Prompt: []byte("exact prompt\r\n세계\n"), Experiment: []byte("{\"model\":\"fixture\"}\n"),
		Cwd: workspace, Timeout: 5 * time.Second,
	}, output
}

func TestExecutePreservesRawInputsAndOutputs(t *testing.T) {
	plan, output := fixture(t, "success")
	record, err := Execute(context.Background(), plan, output)
	if err != nil {
		t.Fatal(err)
	}
	execution, _ := record["execution"].(map[string]any)
	if execution["status"] != "completed" {
		t.Fatalf("execution status = %v, want completed", execution["status"])
	}
	for name, want := range map[string]string{
		"prompt.md": string(plan.Prompt), "experiment.json": string(plan.Experiment),
		"response.txt": string(plan.Prompt), "events.jsonl": "raw\x00events\n", "errors.log": "raw\xffdiagnostic\n",
	} {
		got, err := os.ReadFile(filepath.Join(output, name))
		if err != nil || string(got) != want {
			t.Errorf("%s = %q (%v), want %q", name, got, err, want)
		}
	}
}

func TestExecuteFinalizesMeasuredRecord(t *testing.T) {
	plan, output := fixture(t, "success")
	record, err := Execute(context.Background(), plan, output)
	if err != nil {
		t.Fatal(err)
	}
	if record["record_status"] != "finalized" {
		t.Fatalf("record status = %v, want finalized", record["record_status"])
	}
	execution := record["execution"].(map[string]any)
	if execution["exit_code"] != 0 || execution["timing_method"] != "monotonic_clock" || execution["timing_scope"] != "cli_process" {
		t.Fatalf("unexpected execution measurements: %v", execution)
	}
	if duration, ok := execution["duration_ms"].(float64); !ok || duration < 0 {
		t.Fatalf("invalid duration: %v", execution["duration_ms"])
	}
	for _, key := range []string{"started_at", "finished_at"} {
		value, err := time.Parse(time.RFC3339Nano, execution[key].(string))
		if err != nil || value.Location() != time.UTC {
			t.Errorf("invalid UTC %s: %v (%v)", key, execution[key], err)
		}
	}
	for _, item := range record["capture"].(map[string]any) {
		if item.(map[string]any)["status"] != "captured" {
			t.Errorf("capture should be complete: %v", item)
		}
	}
	saved, err := os.ReadFile(filepath.Join(output, "run.json"))
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(saved, &decoded); err != nil {
		t.Fatal(err)
	}
	want, _ := json.Marshal(record)
	got, _ := json.Marshal(decoded)
	if string(want) != string(got) {
		t.Fatalf("saved record differs: %s != %s", got, want)
	}
}

func TestExecutePreservesNonzeroAttempt(t *testing.T) {
	plan, output := fixture(t, "nonzero")
	record, err := Execute(context.Background(), plan, output)
	if err != nil {
		t.Fatalf("failed attempt must remain evidence, got error: %v", err)
	}
	execution := record["execution"].(map[string]any)
	if execution["status"] != "failed" || execution["exit_code"] != 7 || execution["reason"] != "nonzero_exit" {
		t.Fatalf("unexpected failed execution: %v", execution)
	}
	response := record["capture"].(map[string]any)["response"].(map[string]any)
	if response["status"] != "partial" || response["reason"] != "nonzero_exit" {
		t.Fatalf("unexpected response capture: %v", response)
	}
	data, err := os.ReadFile(filepath.Join(output, "response.txt"))
	if err != nil || string(data) != string(plan.Prompt) {
		t.Fatalf("partial response lost: %q (%v)", data, err)
	}
}

func TestExecuteMarksMissingFinalResponse(t *testing.T) {
	plan, output := fixture(t, "missing")
	record, err := Execute(context.Background(), plan, output)
	if err != nil {
		t.Fatal(err)
	}
	execution := record["execution"].(map[string]any)
	if execution["status"] != "failed" || execution["exit_code"] != 0 || execution["reason"] != "missing_final_response" {
		t.Fatalf("missing response must fail an exit-zero attempt: %v", execution)
	}
	response := record["capture"].(map[string]any)["response"].(map[string]any)
	if response["status"] != "unavailable" || response["reason"] != "missing_final_response" {
		t.Fatalf("missing response capture: %v", response)
	}
}

func TestExecutePreservesSpawnFailure(t *testing.T) {
	plan, output := fixture(t, "success")
	plan.Argv = []string{filepath.Join(plan.Cwd, "missing-cli")}
	record, err := Execute(context.Background(), plan, output)
	if err != nil {
		t.Fatalf("spawn failure should be finalized evidence: %v", err)
	}
	execution := record["execution"].(map[string]any)
	if execution["status"] != "failed" || execution["reason"] != "spawn_error" || execution["exit_code"] != nil {
		t.Fatalf("unexpected spawn failure: %v", execution)
	}
	if detail, ok := execution["error"].(string); !ok || !strings.Contains(detail, "missing-cli") {
		t.Fatalf("missing spawn diagnostic: %v", execution["error"])
	}
	for _, item := range record["capture"].(map[string]any) {
		if item.(map[string]any)["status"] != "unavailable" {
			t.Errorf("unstarted capture must be unavailable: %v", item)
		}
	}
	if _, err := os.Stat(filepath.Join(output, "run.json")); err != nil {
		t.Fatalf("failed record not saved: %v", err)
	}
}

func TestExecuteTimeoutPreservesPartialEvidence(t *testing.T) {
	plan, output := fixture(t, "sleep")
	plan.Timeout = 100 * time.Millisecond
	record, err := Execute(context.Background(), plan, output)
	if err != nil {
		t.Fatal(err)
	}
	execution := record["execution"].(map[string]any)
	if execution["reason"] != "timeout" || execution["status"] != "failed" {
		t.Fatalf("timeout was not recorded: %v", execution)
	}
	if execution["duration_ms"].(float64) > 1500 || execution["exit_code"].(int) >= 0 {
		t.Fatalf("process did not stop promptly: %v", execution)
	}
	events := record["capture"].(map[string]any)["events"].(map[string]any)
	if events["status"] != "partial" || events["reason"] != "timeout" {
		t.Fatalf("timeout capture: %v", events)
	}
	for name, want := range map[string]string{"events.jsonl": "started\n", "errors.log": "diagnostic\n"} {
		got, err := os.ReadFile(filepath.Join(output, name))
		if err != nil || string(got) != want {
			t.Errorf("missing partial %s: %q (%v)", name, got, err)
		}
	}
}

func TestExecuteContextCancellationFinalizesInterruptedAttempt(t *testing.T) {
	plan, output := fixture(t, "sleep")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	timer := time.AfterFunc(100*time.Millisecond, cancel)
	defer timer.Stop()
	record, err := Execute(ctx, plan, output)
	if err != nil {
		t.Fatal(err)
	}
	execution := record["execution"].(map[string]any)
	if execution["status"] != "interrupted" || execution["reason"] != "context_cancelled" {
		t.Fatalf("interruption not preserved: %v", execution)
	}
	if execution["duration_ms"].(float64) > 1500 {
		t.Fatalf("cancellation ignored: %v", execution)
	}
	events := record["capture"].(map[string]any)["events"].(map[string]any)
	if events["status"] != "partial" || events["reason"] != "context_cancelled" {
		t.Fatalf("interruption capture: %v", events)
	}
}

func TestExecuteCancelledContextDoesNotStartProcess(t *testing.T) {
	plan, output := fixture(t, "success")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	record, err := Execute(ctx, plan, output)
	if err != nil {
		t.Fatal(err)
	}
	execution := record["execution"].(map[string]any)
	if execution["status"] != "interrupted" || execution["exit_code"] != nil {
		t.Fatalf("already cancelled context started a process: %v", execution)
	}
	events := record["capture"].(map[string]any)["events"].(map[string]any)
	if events["status"] != "unavailable" {
		t.Fatalf("already cancelled capture: %v", events)
	}
}

func TestExecuteRejectsInvalidInputsBeforeCreatingArchive(t *testing.T) {
	for _, name := range []string{"zero timeout", "negative timeout", "empty argv", "empty command", "empty workspace", "empty archive", "nil context"} {
		t.Run(name, func(t *testing.T) {
			plan, output := fixture(t, "success")
			ctx := context.Background()
			switch name {
			case "zero timeout":
				plan.Timeout = 0
			case "negative timeout":
				plan.Timeout = -time.Second
			case "empty argv":
				plan.Argv = nil
			case "empty command":
				plan.Argv = []string{""}
			case "empty workspace":
				plan.Cwd = ""
			case "empty archive":
				output = ""
			case "nil context":
				ctx = nil
			}
			if _, err := Execute(ctx, plan, output); err == nil {
				t.Fatal("invalid inputs accepted")
			}
			if output != "" {
				if _, err := os.Stat(output); !os.IsNotExist(err) {
					t.Fatalf("archive created for invalid input: %v", err)
				}
			}
		})
	}
}

func TestExecuteRejectsOverlappingWorkspaceAndArchive(t *testing.T) {
	for _, direction := range []string{"same", "archive inside workspace", "workspace inside archive", "symlink parent", "relative symlink", "dangling symlink", "symlink before dotdot"} {
		t.Run(direction, func(t *testing.T) {
			plan, output := fixture(t, "success")
			switch direction {
			case "same":
				output = plan.Cwd
			case "archive inside workspace":
				output = filepath.Join(plan.Cwd, "archive")
			case "workspace inside archive":
				plan.Cwd = filepath.Join(output, "workspace")
			case "symlink parent":
				link := filepath.Join(filepath.Dir(plan.Cwd), "alias")
				if err := os.Symlink(plan.Cwd, link); err != nil {
					t.Fatal(err)
				}
				output = filepath.Join(link, "missing", "archive")
			case "relative symlink":
				link := filepath.Join(filepath.Dir(plan.Cwd), "alias")
				if err := os.Symlink("workspace", link); err != nil {
					t.Fatal(err)
				}
				output = filepath.Join(link, "archive")
			case "dangling symlink":
				link := filepath.Join(filepath.Dir(plan.Cwd), "alias")
				if err := os.Symlink(filepath.Join(plan.Cwd, "missing"), link); err != nil {
					t.Fatal(err)
				}
				output = filepath.Join(link, "archive")
			case "symlink before dotdot":
				link := filepath.Join(filepath.Dir(plan.Cwd), "alias")
				if err := os.Symlink(filepath.Join(plan.Cwd, "nested"), link); err != nil {
					t.Fatal(err)
				}
				output = link + "/../archive"
			}
			if _, err := Execute(context.Background(), plan, output); err == nil || !strings.Contains(err.Error(), "overlap") {
				t.Fatalf("overlap not rejected: %v", err)
			}
		})
	}
}

func TestExecuteStopsLingeringChildrenBeforeFinalizing(t *testing.T) {
	plan, output := fixture(t, "parent")
	trigger := filepath.Join(plan.Cwd, "release")
	plan.Argv = append(plan.Argv, trigger)
	record, err := Execute(context.Background(), plan, output)
	if err != nil {
		t.Fatal(err)
	}
	events, err := os.ReadFile(filepath.Join(output, "events.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(trigger + ".ready"); err != nil {
		t.Fatalf("child never became ready: %v", err)
	}
	if err := os.WriteFile(trigger, []byte("release after finalization"), 0600); err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)
	response, err := os.ReadFile(filepath.Join(output, "response.txt"))
	if err != nil || string(response) != "original response" {
		t.Fatalf("finalized evidence modified by descendant: %q (%v)", response, err)
	}
	after, err := os.ReadFile(filepath.Join(output, "events.jsonl"))
	if err != nil || string(after) != string(events) {
		t.Fatalf("finalized events modified: %q (%v)", after, err)
	}
	execution := record["execution"].(map[string]any)
	if execution["status"] != "failed" || execution["exit_code"] != 0 || execution["reason"] != "lingering_processes" {
		t.Fatalf("lingering descendants not reported: %v", execution)
	}
}

func TestExecuteKillsChildrenIgnoringTermination(t *testing.T) {
	for _, mode := range []string{"tree-ignore", "parent-ignore"} {
		t.Run(mode, func(t *testing.T) {
			plan, output := fixture(t, mode)
			ready := filepath.Join(plan.Cwd, "ready")
			plan.Argv = append(plan.Argv, ready)
			if mode == "tree-ignore" {
				plan.Timeout = 300 * time.Millisecond
			}
			record, err := Execute(context.Background(), plan, output)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(ready); err != nil {
				t.Fatalf("child never became ready: %v", err)
			}
			if err := os.WriteFile(ready+".release", []byte("release after finalization"), 0600); err != nil {
				t.Fatal(err)
			}
			time.Sleep(200 * time.Millisecond)
			response, err := os.ReadFile(filepath.Join(output, "response.txt"))
			if err != nil || string(response) != "original response" {
				t.Fatalf("child survived cleanup: %q (%v)", response, err)
			}
			execution := record["execution"].(map[string]any)
			if mode == "tree-ignore" {
				if execution["reason"] != "timeout" || execution["exit_code"] != -9 {
					t.Fatalf("timeout group was not killed: %v", execution)
				}
			} else if execution["reason"] != "lingering_processes" || execution["exit_code"] != 7 {
				t.Fatalf("nonzero parent left a child: %v", execution)
			}
		})
	}
}

func TestExecuteNeverOverwritesExistingArchive(t *testing.T) {
	plan, output := fixture(t, "success")
	if err := os.MkdirAll(output, 0700); err != nil {
		t.Fatal(err)
	}
	response := filepath.Join(output, "response.txt")
	if err := os.WriteFile(response, []byte("original evidence"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Execute(context.Background(), plan, output); !os.IsExist(err) {
		t.Fatalf("existing archive accepted: %v", err)
	}
	data, err := os.ReadFile(response)
	if err != nil || string(data) != "original evidence" {
		t.Fatalf("original evidence overwritten: %q (%v)", data, err)
	}
	entries, err := os.ReadDir(output)
	if err != nil || len(entries) != 1 {
		t.Fatalf("archive modified: %v (%v)", entries, err)
	}
}

func TestExecuteRejectsSymlinkLoops(t *testing.T) {
	plan, output := fixture(t, "success")
	link := filepath.Join(filepath.Dir(plan.Cwd), "loop")
	if err := os.Symlink("loop", link); err != nil {
		t.Fatal(err)
	}
	plan.Cwd = link
	if _, err := Execute(context.Background(), plan, output); err == nil || !strings.Contains(err.Error(), "symlinks") {
		t.Fatalf("symlink loop accepted: %v", err)
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatalf("loop created archive: %v", err)
	}
}

func TestExecutePreservesMissingWorkspaceAsSpawnFailure(t *testing.T) {
	plan, output := fixture(t, "success")
	plan.Cwd = filepath.Join(plan.Cwd, "missing")
	record, err := Execute(context.Background(), plan, output)
	if err != nil {
		t.Fatal(err)
	}
	if execution := record["execution"].(map[string]any); execution["reason"] != "spawn_error" {
		t.Fatalf("missing workspace outcome: %v", execution)
	}
}

func TestExecuteRejectsCaseAliasesOfWorkspace(t *testing.T) {
	for _, direction := range []string{"archive inside workspace", "workspace inside archive"} {
		t.Run(direction, func(t *testing.T) {
			plan, output := fixture(t, "success")
			alias := filepath.Join(filepath.Dir(plan.Cwd), strings.ToUpper(filepath.Base(plan.Cwd)))
			workspaceInfo, err := os.Stat(plan.Cwd)
			if err != nil {
				t.Fatal(err)
			}
			aliasInfo, err := os.Stat(alias)
			if err != nil || !os.SameFile(workspaceInfo, aliasInfo) {
				t.Skip("filesystem is case sensitive")
			}
			if direction == "archive inside workspace" {
				output = filepath.Join(alias, "attempt")
			} else {
				output, plan.Cwd = plan.Cwd, filepath.Join(alias, "nested-workspace")
			}
			if _, err := Execute(context.Background(), plan, output); err == nil || !strings.Contains(err.Error(), "overlap") {
				t.Fatalf("case alias overlap not rejected: %v", err)
			}
		})
	}
}

func TestExecuteDoesNotCaptureSymlinkResponse(t *testing.T) {
	plan, output := fixture(t, "symlink")
	target := filepath.Join(plan.Cwd, "mutable-response")
	if err := os.WriteFile(target, []byte("original response"), 0600); err != nil {
		t.Fatal(err)
	}
	plan.Argv = append(plan.Argv, target)
	record, err := Execute(context.Background(), plan, output)
	if err != nil {
		t.Fatal(err)
	}
	execution := record["execution"].(map[string]any)
	if execution["status"] != "failed" || execution["reason"] != "non_regular_final_response" {
		t.Fatalf("symlink response accepted: %v", execution)
	}
	response := record["capture"].(map[string]any)["response"].(map[string]any)
	if response["status"] != "unavailable" || response["reason"] != "non_regular_final_response" {
		t.Fatalf("symlink response claimed as collected: %v", response)
	}
	if err := os.WriteFile(target, []byte("changed later"), 0600); err != nil {
		t.Fatal(err)
	}
}
