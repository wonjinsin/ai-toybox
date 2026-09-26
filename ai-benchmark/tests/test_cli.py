import contextlib
import hashlib
import io
import json
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

from ai_benchmark.cli import main
from test_config import fixture


def invoke(argv):
    out, err = io.StringIO(), io.StringIO()
    with contextlib.redirect_stdout(out), contextlib.redirect_stderr(err):
        code = main(argv)
    return code, out.getvalue(), err.getvalue()


class CliTests(unittest.TestCase):
    def test_inspect_can_report_escaped_surrogates_on_utf8_stdout(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "events.jsonl"
            raw = b'{"type":"error","message":"\\ud800"}\n'
            path.write_bytes(raw)
            result = subprocess.run(
                [sys.executable, "-X", "utf8", "-m", "ai_benchmark", "inspect", str(path)],
                capture_output=True, timeout=5,
            )
            self.assertEqual(result.returncode, 0, result.stderr)
            data = json.loads(result.stdout.decode("utf-8"))
            self.assertEqual(data["summary"]["diagnostics"][0]["message"], "\ud800")
            self.assertEqual(path.read_bytes(), raw)

    def test_inspect_summarizes_existing_events_without_modifying_them(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "captured.jsonl"
            raw = b'{"type":"turn.completed","usage":{"input_tokens":12,"cached_input_tokens":4,"output_tokens":5}}\n'
            path.write_bytes(raw)
            try:
                code, out, err = invoke(["inspect", str(path)])
            except SystemExit as error:
                self.fail(f"Inspection command is unavailable: {error.code}")
            self.assertEqual(code, 0)
            self.assertFalse(err)
            result = json.loads(out)
            self.assertEqual(result["source"]["sha256"], hashlib.sha256(raw).hexdigest())
            self.assertEqual(result["source"]["path"], str(path.resolve()))
            self.assertEqual(result["summary"]["usage"]["input_tokens"], 12)
            self.assertEqual(path.read_bytes(), raw)
            self.assertEqual(list(Path(directory).iterdir()), [path])

    def test_preparation_commands_work_without_a_model_call(self):
        with tempfile.TemporaryDirectory() as directory:
            path, _, _ = fixture(directory)
            output = Path(directory) / "attempt-1"
            for command in ("validate", "preview", "prepare"):
                with self.subTest(command=command):
                    argv = [command, str(path)]
                    if command != "validate":
                        argv += ["--output", str(output)]
                    code, out, err = invoke(argv)
                    self.assertEqual(code, 0)
                    self.assertFalse(err)
                    result = json.loads(out)
                    self.assertIsInstance(result, dict)
            self.assertTrue((output / "run.json").exists())

    def test_invalid_inputs_produce_an_error_without_a_traceback(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "bad.json"
            path.write_text('{"status":"draft"}')
            try:
                code, out, err = invoke(["validate", str(path)])
            except Exception as error:
                self.fail(f"CLI leaked {type(error).__name__} instead of returning a diagnostic")
            self.assertEqual(code, 2)
            self.assertFalse(out)
            self.assertIn("ready", err)
            self.assertNotIn("Traceback", err)

    def test_live_execution_is_refused_with_a_distinct_exit_code(self):
        with tempfile.TemporaryDirectory() as directory:
            path, _, _ = fixture(directory)
            output = Path(directory) / "attempt-1"
            try:
                code, out, err = invoke(["run", str(path), "--output", str(output)])
            except SystemExit as error:
                self.fail(f"Expected an isolation diagnostic, got parser exit {error.code}")
            self.assertEqual(code, 3)
            self.assertFalse(out)
            self.assertIn("isolation", err)
            self.assertFalse(output.exists())
