"""Collect raw evidence from one externally validated execution plan."""

from dataclasses import dataclass
from datetime import datetime, timezone
from pathlib import Path
import json
import math
import os
import signal
import subprocess
import time


@dataclass(frozen=True)
class ExecutionPlan:
    argv: tuple[str, ...]
    prompt: bytes
    experiment: bytes
    cwd: str
    timeout_seconds: float


def _group_remains(process):
    if os.name != "posix":
        return False
    try:
        os.killpg(process.pid, 0)
    except ProcessLookupError:
        return False
    except PermissionError:
        return True
    return True


def _wait_for_group_exit(process, seconds):
    deadline = time.monotonic() + seconds
    while _group_remains(process):
        if time.monotonic() >= deadline:
            return False
        time.sleep(0.01)
    return True


def _stop(process):
    def send(kill=False):
        try:
            if os.name == "posix":
                os.killpg(process.pid, signal.SIGKILL if kill else signal.SIGTERM)
            elif kill:
                process.kill()
            else:
                process.terminate()
        except ProcessLookupError:
            pass
        except PermissionError:
            if os.name != "posix" or not _wait_for_group_exit(process, 0.05):
                raise

    deadline = time.monotonic() + 0.25
    send()
    try:
        process.communicate(timeout=max(0, deadline - time.monotonic()))
    except subprocess.TimeoutExpired:
        pass
    if os.name == "posix" and _wait_for_group_exit(process, max(0, deadline - time.monotonic())):
        return
    send(kill=True)
    process.communicate()
    if os.name == "posix":
        _wait_for_group_exit(process, 0.05)


def execute(plan: ExecutionPlan, output_dir):
    """Collect one attempt after the caller validates its isolation profile.

    The caller must already direct the CLI's final-response export to the
    absolute output_dir/response.txt path in argv. This primitive does not
    modify argv, configure credentials, or enforce benchmark isolation.
    Cleanup covers descendants remaining in the original POSIX process group.
    """
    timeout = plan.timeout_seconds
    if (
        isinstance(timeout, bool)
        or not isinstance(timeout, (int, float))
        or not math.isfinite(timeout)
        or timeout <= 0
    ):
        raise ValueError("timeout_seconds must be a positive finite number")
    output = Path(output_dir).resolve()
    workspace = Path(plan.cwd).resolve()
    if output == workspace or output in workspace.parents or workspace in output.parents:
        raise ValueError("Execution workspace and archive must not overlap")
    output.mkdir(parents=True, exist_ok=False)
    (output / "prompt.md").write_bytes(plan.prompt)
    (output / "experiment.json").write_bytes(plan.experiment)
    process = None
    reason = None
    error_detail = None
    with (output / "events.jsonl").open("wb") as events, (output / "errors.log").open("wb") as errors:
        started_at = datetime.now(timezone.utc).isoformat()
        started = time.monotonic()
        try:
            process = subprocess.Popen(
                plan.argv,
                cwd=plan.cwd,
                stdin=subprocess.PIPE,
                stdout=events,
                stderr=errors,
                shell=False,
                start_new_session=os.name == "posix",
            )
            process.communicate(input=plan.prompt, timeout=plan.timeout_seconds)
        except subprocess.TimeoutExpired:
            reason = "timeout"
            _stop(process)
        except KeyboardInterrupt:
            reason = "keyboard_interrupt"
            if process is not None:
                _stop(process)
        except OSError as error:
            reason = "spawn_error" if process is None else "process_error"
            error_detail = str(error)
            if process is not None:
                _stop(process)
        finally:
            if process is not None and process.stdin is not None:
                process.stdin.close()
            if process is not None and _group_remains(process):
                reason = reason or "lingering_processes"
                _stop(process)
        duration_ms = (time.monotonic() - started) * 1000
        finished_at = datetime.now(timezone.utc).isoformat()
    has_response = (output / "response.txt").is_file()
    exit_code = process.returncode if process is not None else None
    stream_status = "unavailable" if process is None else "partial" if reason else "captured"
    status = "completed" if exit_code == 0 and has_response and not reason else "failed"
    failure_reason = (
        reason
        or ("nonzero_exit" if exit_code else None)
        or (None if has_response else "missing_final_response")
    )
    response_status = "partial" if failure_reason else "captured"
    record = {
        "record_status": "finalized",
        "execution": {
            "status": "interrupted" if reason == "keyboard_interrupt" else status,
            "exit_code": exit_code,
            "reason": failure_reason,
            "error": error_detail,
            "started_at": started_at,
            "finished_at": finished_at,
            "duration_ms": duration_ms,
            "timing_method": "monotonic_clock",
            "timing_scope": "cli_process",
        },
        "capture": {
            "response": {
                "path": "response.txt",
                "status": response_status if has_response else "unavailable",
                "reason": failure_reason if has_response else "missing_final_response",
            },
            "stderr": {"path": "errors.log", "status": stream_status, "reason": reason},
            "events": {"path": "events.jsonl", "status": stream_status, "reason": reason},
        },
    }
    (output / "run.json").write_text(json.dumps(record, indent=2) + "\n", encoding="utf-8")
    return record
