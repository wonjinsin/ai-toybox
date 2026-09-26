import sys
import json
import gc
import os
import signal
import subprocess
import tempfile
import time
import unittest
import warnings
from dataclasses import replace
from dataclasses import FrozenInstanceError
from datetime import datetime
from pathlib import Path
from unittest.mock import patch

from ai_benchmark.process import ExecutionPlan, execute


class ProcessTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        self.workspace = self.root / "workspace"
        self.workspace.mkdir()
        self.output = self.root / "archive" / "attempt"

    def plan(self, code, *, timeout=5.0, argv=None):
        return ExecutionPlan(
            argv=argv or (sys.executable, "-c", code, str(self.output / "response.txt")),
            prompt=b"exact prompt\r\n\xec\x84\xb8\xea\xb3\x84\n",
            experiment=b'{"model":"fixture"}\n',
            cwd=str(self.workspace),
            timeout_seconds=timeout,
        )

    def test_completed_process_preserves_input_and_raw_output(self):
        plan = self.plan(
            "import pathlib, sys; "
            "pathlib.Path(sys.argv[1]).write_bytes(sys.stdin.buffer.read()); "
            "sys.stdout.buffer.write(b'raw\\x00events\\n'); "
            "sys.stderr.buffer.write(b'raw\\xffdiagnostic\\n')"
        )
        record = execute(plan, self.output)
        self.assertEqual(record.get("execution", {}).get("status"), "completed")
        self.assertEqual((self.output / "response.txt").read_bytes(), plan.prompt)
        self.assertEqual((self.output / "prompt.md").read_bytes(), plan.prompt)
        self.assertEqual((self.output / "experiment.json").read_bytes(), plan.experiment)
        self.assertEqual((self.output / "events.jsonl").read_bytes(), b"raw\x00events\n")
        self.assertEqual((self.output / "errors.log").read_bytes(), b"raw\xffdiagnostic\n")
        self.assertEqual(record["execution"]["exit_code"], 0)

    def test_timeout_preserves_partial_output_and_finalizes(self):
        record = execute(self.plan(
            "import sys, time; print('started', flush=True); "
            "print('diagnostic', file=sys.stderr, flush=True); time.sleep(2)",
            timeout=0.1,
        ), self.output)
        self.assertEqual(record["execution"].get("reason"), "timeout")
        self.assertEqual(record["execution"]["status"], "failed")
        self.assertLess(record["execution"]["duration_ms"], 1500)
        self.assertEqual(record["capture"]["events"]["status"], "partial")
        self.assertIn(b"started", (self.output / "events.jsonl").read_bytes())
        self.assertIn(b"diagnostic", (self.output / "errors.log").read_bytes())
        self.assertEqual(json.loads((self.output / "run.json").read_text()), record)

    def test_timeout_requires_positive_finite_seconds(self):
        for invalid in (0, -1, float("inf"), float("nan"), True, None):
            with self.subTest(timeout=invalid):
                with self.assertRaisesRegex(ValueError, "timeout_seconds"):
                    execute(self.plan("pass", timeout=invalid), self.output)
                self.assertFalse(self.output.exists())

    def test_immediate_timeout_closes_process_input_stream(self):
        with warnings.catch_warnings(record=True) as caught:
            warnings.simplefilter("always", ResourceWarning)
            record = execute(self.plan("import time; time.sleep(2)", timeout=0.000001), self.output)
            gc.collect()
        self.assertEqual(record["execution"]["reason"], "timeout")
        self.assertFalse(tuple(w for w in caught if issubclass(w.category, ResourceWarning)))

    def test_workspace_and_archive_must_have_separate_roots(self):
        plan = self.plan("print('should not run')")
        with self.assertRaisesRegex(ValueError, "overlap"):
            execute(plan, self.workspace / "nested-archive")
        with self.assertRaisesRegex(ValueError, "overlap"):
            execute(replace(plan, cwd=str(self.output / "nested-workspace")), self.output)
        with self.assertRaisesRegex(ValueError, "overlap"):
            execute(plan, self.workspace)

    def test_keyboard_interrupt_finalizes_partial_attempt(self):
        original = subprocess.Popen.communicate
        processes = ()
        interrupted = False

        def communicate(process, *args, **kwargs):
            nonlocal processes, interrupted
            if not interrupted:
                interrupted = True
                processes = (*processes, process)
                raise KeyboardInterrupt
            return original(process, *args, **kwargs)

        try:
            with patch.object(subprocess.Popen, "communicate", communicate):
                try:
                    record = execute(self.plan("import time; time.sleep(2)"), self.output)
                except KeyboardInterrupt:
                    self.fail("KeyboardInterrupt escaped without finalizing the attempt")
        finally:
            for process in processes:
                if process.poll() is None:
                    process.kill()
                    original(process)
        self.assertEqual(record["execution"]["status"], "interrupted")
        self.assertEqual(record["execution"]["reason"], "keyboard_interrupt")
        self.assertEqual(record["capture"]["events"]["status"], "partial")
        self.assertEqual(json.loads((self.output / "run.json").read_text()), record)

    def test_communication_error_stops_process_and_preserves_evidence(self):
        original = subprocess.Popen.communicate
        processes = ()

        def communicate(process, *args, **kwargs):
            nonlocal processes
            if not processes:
                processes = (process,)
                raise OSError("communication fixture failure")
            return original(process, *args, **kwargs)

        try:
            with patch.object(subprocess.Popen, "communicate", communicate):
                record = execute(self.plan("import time; time.sleep(2)"), self.output)
            self.assertIsNotNone(processes[0].poll(), "Process survived communication failure")
            self.assertEqual(record["execution"]["status"], "failed")
            self.assertEqual(record["execution"]["reason"], "process_error")
            self.assertEqual(record["execution"]["error"], "communication fixture failure")
            self.assertEqual(record["capture"]["events"]["status"], "partial")
            self.assertEqual(json.loads((self.output / "run.json").read_text()), record)
        finally:
            for process in processes:
                if process.poll() is None:
                    process.kill()
                    original(process)

    def test_existing_archive_is_never_overwritten(self):
        self.output.mkdir(parents=True)
        original = self.output / "response.txt"
        original.write_bytes(b"original evidence")
        with self.assertRaises(FileExistsError):
            execute(self.plan("pass"), self.output)
        self.assertEqual(original.read_bytes(), b"original evidence")
        self.assertEqual(tuple(path.name for path in self.output.iterdir()), ("response.txt",))

    @unittest.skipUnless(os.name == "posix", "POSIX process-group behavior")
    def test_successful_parent_cannot_leave_descendants_modifying_finalized_evidence(self):
        child = (
            "import pathlib, sys, time; time.sleep(0.3); "
            "pathlib.Path(sys.argv[1]).write_text('late modification'); "
            "print('late event', flush=True)"
        )
        parent = (
            "import pathlib, subprocess, sys; "
            "pathlib.Path(sys.argv[1]).write_text('original response'); "
            f"subprocess.Popen([sys.executable, '-c', {child!r}, sys.argv[1]])"
        )
        record = execute(self.plan(parent), self.output)
        events = (self.output / "events.jsonl").read_bytes()
        time.sleep(0.5)
        self.assertEqual((self.output / "response.txt").read_text(), "original response")
        self.assertEqual((self.output / "events.jsonl").read_bytes(), events)
        self.assertEqual(record["execution"]["exit_code"], 0)
        self.assertEqual(record["execution"]["status"], "failed")
        self.assertEqual(record["execution"]["reason"], "lingering_processes")
        self.assertEqual(record["capture"]["response"]["status"], "partial")

    @unittest.skipUnless(os.name == "posix", "POSIX process-group behavior")
    def test_failed_parent_cleans_up_descendants_that_ignore_termination(self):
        ready = self.workspace / "ready"
        child = (
            "import pathlib, signal, sys, time; "
            "signal.signal(signal.SIGTERM, signal.SIG_IGN); "
            f"pathlib.Path({str(ready)!r}).write_text('ready'); "
            "time.sleep(0.6); pathlib.Path(sys.argv[1]).write_text('late modification')"
        )
        parent = (
            "import pathlib, subprocess, sys, time\n"
            "pathlib.Path(sys.argv[1]).write_text('original response')\n"
            f"subprocess.Popen([sys.executable, '-c', {child!r}, sys.argv[1]])\n"
            f"while not pathlib.Path({str(ready)!r}).exists(): time.sleep(0.001)\n"
            "sys.exit(7)"
        )
        record = execute(self.plan(parent), self.output)
        time.sleep(0.7)
        self.assertEqual((self.output / "response.txt").read_text(), "original response")
        self.assertEqual(record["execution"]["exit_code"], 7)
        self.assertEqual(record["execution"]["status"], "failed")
        self.assertEqual(record["execution"]["reason"], "lingering_processes")

    def test_execution_plan_fields_are_frozen(self):
        plan = self.plan("pass")
        with self.assertRaises(FrozenInstanceError):
            plan.cwd = "another-workspace"

    @unittest.skipUnless(os.name == "posix", "POSIX process-group behavior")
    def test_timeout_kills_descendants_that_ignore_termination(self):
        ready = self.workspace / "ready"
        survived = self.workspace / "survived"
        child = (
            "import os, pathlib, signal, time; "
            "signal.signal(signal.SIGTERM, signal.SIG_IGN); "
            f"pathlib.Path({str(ready)!r}).write_text(str(os.getpid())); "
            "time.sleep(0.8); "
            f"pathlib.Path({str(survived)!r}).write_text('escaped')"
        )
        parent = (
            "import signal, subprocess, sys, time; "
            "signal.signal(signal.SIGTERM, signal.SIG_IGN); "
            f"subprocess.Popen([sys.executable, '-c', {child!r}]); "
            "time.sleep(2)"
        )
        try:
            record = execute(self.plan(parent, timeout=0.2), self.output)
            self.assertEqual(record["execution"]["exit_code"], -signal.SIGKILL)
            self.assertTrue(ready.exists(), "Child must have started for this check")
            time.sleep(0.8)
            self.assertFalse(survived.exists(), "Descendant survived timeout cleanup")
        finally:
            if ready.exists():
                try:
                    os.kill(int(ready.read_text()), signal.SIGKILL)
                except ProcessLookupError:
                    pass

    def test_spawn_error_is_recorded_without_losing_snapshots(self):
        plan = self.plan("", argv=(str(self.root / "missing-cli"),))
        try:
            record = execute(plan, self.output)
        except OSError as error:
            self.fail(f"Spawn error escaped instead of becoming evidence: {error}")
        self.assertEqual(record["execution"]["status"], "failed")
        self.assertEqual(record["execution"]["reason"], "spawn_error")
        self.assertIn("missing-cli", record["execution"]["error"])
        self.assertIsNone(record["execution"]["exit_code"])
        self.assertEqual(record["capture"]["events"]["status"], "unavailable")
        self.assertEqual((self.output / "prompt.md").read_bytes(), plan.prompt)
        self.assertEqual(json.loads((self.output / "run.json").read_text()), record)

    def test_nonzero_exit_preserves_partial_evidence(self):
        record = execute(self.plan(
            "import pathlib, sys; pathlib.Path(sys.argv[1]).write_text('partial'); "
            "print('event before failure'); print('failure details', file=sys.stderr); sys.exit(7)"
        ), self.output)
        self.assertEqual(record["execution"]["status"], "failed")
        self.assertEqual(record["execution"]["exit_code"], 7)
        self.assertEqual((self.output / "response.txt").read_text(), "partial")
        self.assertIn(b"failure details", (self.output / "errors.log").read_bytes())
        self.assertEqual(record["capture"]["response"]["status"], "partial")
        self.assertEqual(record["capture"]["response"]["reason"], "nonzero_exit")

    def test_finalized_record_describes_capture_and_external_timing(self):
        record = execute(self.plan(
            "import pathlib, sys; pathlib.Path(sys.argv[1]).write_text('done')"
        ), self.output)
        self.assertEqual(record.get("record_status"), "finalized")
        execution = record["execution"]
        self.assertGreaterEqual(execution["duration_ms"], 0)
        self.assertEqual(execution["timing_method"], "monotonic_clock")
        self.assertEqual(execution["timing_scope"], "cli_process")
        started = datetime.fromisoformat(execution["started_at"])
        finished = datetime.fromisoformat(execution["finished_at"])
        self.assertIsNotNone(started.utcoffset())
        self.assertGreaterEqual(finished, started)
        self.assertEqual(record["capture"]["response"]["status"], "captured")
        self.assertEqual(record["capture"]["stderr"]["status"], "captured")
        self.assertEqual(record["capture"]["events"]["path"], "events.jsonl")
        self.assertEqual(json.loads((self.output / "run.json").read_text()), record)

    def test_missing_response_is_unavailable_even_after_exit_zero(self):
        record = execute(self.plan("print('finished without response')"), self.output)
        self.assertEqual(record["capture"]["response"]["status"], "unavailable")
        self.assertEqual(record["capture"]["response"]["reason"], "missing_final_response")
        self.assertEqual(record["execution"]["status"], "failed")
        self.assertEqual(record["execution"]["exit_code"], 0)


if __name__ == "__main__":
    unittest.main()
