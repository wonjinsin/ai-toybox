import hashlib
import json
import os
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

from ai_benchmark.config import ConfigurationError, load_experiment, skill_digest


def fixture(directory):
    root = Path(directory)
    prompt = b"Build the agreed scene.\r\n"
    (root / "task.md").write_bytes(prompt)
    data = {
        "format_version": "0.1", "status": "ready", "experiment_id": "solar-high",
        "task": {"id": "solar", "prompt_path": "task.md", "prompt_version": "v1",
                 "prompt_sha256": hashlib.sha256(prompt).hexdigest()},
        "cli": {"name": "codex", "required_version": "0.156.1", "mode": "non_interactive"},
        "model": "fixture-model", "effort": "high",
        "capabilities": {"skills": [], "tools": [], "mcp_servers": [], "agents": []},
        "artifact": {"delivery": "final_response"}, "isolation_profile": "pending-review"
    }
    path = root / "experiment.json"
    path.write_text(json.dumps(data))
    return path, data, prompt


class ExperimentValidationTests(unittest.TestCase):
    def test_non_unicode_scalar_settings_are_rejected(self):
        with tempfile.TemporaryDirectory() as directory:
            path, data, _ = fixture(directory)
            path.write_text(json.dumps({**data, "task": {**data["task"], "id": "\ud800"}}))
            with self.assertRaisesRegex(ConfigurationError, "task.id.*UTF-8"):
                load_experiment(path)

    @unittest.skipUnless(hasattr(os, "mkfifo"), "Named-pipe fixture requires POSIX")
    def test_named_pipe_is_rejected_without_waiting_for_a_writer(self):
        with tempfile.TemporaryDirectory() as directory:
            pipe = Path(directory) / "pipe.json"
            os.mkfifo(pipe)
            try:
                result = subprocess.run(
                    [sys.executable, "-m", "ai_benchmark", "validate", str(pipe)],
                    capture_output=True, timeout=1,
                )
            except subprocess.TimeoutExpired:
                self.fail("Validation waited for a pipe writer instead of rejecting a non-file input")
            self.assertEqual(result.returncode, 2)
            self.assertIn(b"regular file", result.stderr)

    def test_embedded_control_characters_in_settings_are_rejected(self):
        with tempfile.TemporaryDirectory() as directory:
            path, data, _ = fixture(directory)
            path.write_text(json.dumps({**data, "task": {**data["task"], "id": "solar\x00system"}}))
            with self.assertRaisesRegex(ConfigurationError, "task.id"):
                load_experiment(path)

    def test_draft_experiment_is_rejected(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "experiment.json"
            path.write_text('{"format_version":"0.1","status":"draft"}')
            with self.assertRaisesRegex(ConfigurationError, "status.*ready"):
                load_experiment(path)

    def test_preparation_preserves_exact_input_bytes(self):
        with tempfile.TemporaryDirectory() as directory:
            path, data, prompt = fixture(directory)
            experiment = load_experiment(path)
            self.assertEqual(getattr(experiment, "prompt", None), prompt)
            self.assertEqual(getattr(experiment, "raw", None), path.read_bytes())
            self.assertEqual(getattr(experiment, "model", None), data["model"])
            self.assertEqual(getattr(experiment, "prompt_path", None), (path.parent / "task.md").resolve())

    def test_missing_or_ambiguous_execution_settings_are_rejected(self):
        with tempfile.TemporaryDirectory() as directory:
            path, data, _ = fixture(directory)
            cases = (
                ({**data, "effort": None}, "effort"),
                ({**data, "model": "--unexpected-flag"}, "model"),
                ({**data, "reasoning": "high"}, "unknown"),
                ({**data, "format_version": "99"}, "format_version"),
                ({**data, "capabilities": {**data["capabilities"], "tools": None}}, "tools"),
                ({**data, "cli": {**data["cli"], "name": "unknown"}}, "cli.name"),
                ({**data, "isolation_profile": None}, "isolation_profile"),
                ({**data, "artifact": {"delivery": "workspace_files"}}, "file"),
            )
            for invalid, message in cases:
                with self.subTest(message=message):
                    path.write_text(json.dumps(invalid))
                    with self.assertRaisesRegex(ConfigurationError, message):
                        load_experiment(path)

    def test_invalid_input_files_have_actionable_configuration_errors(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "experiment.json"
            for content in (b"[]", b"null", b"not-json", b"\xff", b'{"status":"ready","status":"draft"}', b"[" * 2000 + b"]" * 2000):
                with self.subTest(prefix=content[:40], length=len(content)):
                    path.write_bytes(content)
                    try:
                        load_experiment(path)
                    except Exception as error:
                        self.assertIsInstance(error, ConfigurationError)
                    else:
                        self.fail("Malformed configuration was accepted")
            path.unlink()
            try:
                load_experiment(path)
            except Exception as error:
                self.assertIsInstance(error, ConfigurationError)
            else:
                self.fail("Missing configuration was accepted")

    def test_changed_empty_or_draft_prompt_cannot_be_submitted(self):
        with tempfile.TemporaryDirectory() as directory:
            path, data, _ = fixture(directory)
            for prompt, update_hash, message in (
                (b"changed", False, "hash"),
                (b"  \n", True, "empty"),
                (b"# Task\nStatus: Draft for review.\n", True, "draft"),
            ):
                with self.subTest(message=message):
                    (path.parent / "task.md").write_bytes(prompt)
                    task = {**data["task"], "prompt_sha256": hashlib.sha256(prompt).hexdigest()} if update_hash else data["task"]
                    path.write_text(json.dumps({**data, "task": task}))
                    with self.assertRaisesRegex(ConfigurationError, message):
                        load_experiment(path)

    def test_skill_bundle_changes_are_rejected(self):
        with tempfile.TemporaryDirectory() as directory:
            path, data, _ = fixture(directory)
            skill = path.parent / "skill"
            skill.mkdir()
            content = b"---\nname: scene-skill\ndescription: Fixture\n---\nInstructions.\n"
            (skill / "SKILL.md").write_bytes(content)
            digest = hashlib.sha256(b"SKILL.md\x00" + hashlib.sha256(content).digest()).hexdigest()
            selection = {"name": "scene-skill", "path": "skill", "sha256": digest, "invocation": "explicit"}
            selected = {**data, "capabilities": {**data["capabilities"], "skills": [selection]}}
            path.write_text(json.dumps(selected))
            load_experiment(path)
            (skill / "helper.txt").write_text("Unapproved additional instructions")
            with self.assertRaisesRegex(ConfigurationError, "skill.*hash"):
                load_experiment(path)

    def test_skill_symlinks_are_rejected(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            skill = root / "skill"
            skill.mkdir()
            (skill / "SKILL.md").write_text("Fixture")
            (skill / "linked.md").symlink_to(skill / "SKILL.md")
            with self.assertRaisesRegex(ConfigurationError, "symbolic links"):
                skill_digest(skill)

    def test_non_utf8_prompt_is_rejected_even_with_matching_hash(self):
        with tempfile.TemporaryDirectory() as directory:
            path, data, _ = fixture(directory)
            prompt = b"\xff"
            (path.parent / "task.md").write_bytes(prompt)
            path.write_text(json.dumps({**data, "task": {**data["task"], "prompt_sha256": hashlib.sha256(prompt).hexdigest()}}))
            with self.assertRaisesRegex(ConfigurationError, "UTF-8"):
                load_experiment(path)
