import hashlib
import json
import tempfile
import unittest
from pathlib import Path

from ai_benchmark.config import load_experiment
from ai_benchmark.launcher import preview, run, UnsupportedCondition
from test_config import fixture


class PreviewTests(unittest.TestCase):
    def test_preview_discloses_unverified_isolation_without_creating_a_run(self):
        with tempfile.TemporaryDirectory() as directory:
            path, data, prompt = fixture(directory)
            output = Path(directory) / "archive" / "attempt-1"
            result = preview(load_experiment(path), output)
            self.assertEqual(result.get("launch_ready"), False)
            self.assertTrue(result.get("blockers"))
            self.assertEqual(result.get("prompt_sha256"), hashlib.sha256(prompt).hexdigest())
            self.assertEqual(result.get("delivered_prompt"), prompt.decode())
            self.assertIn(data["model"], result.get("candidate_argv", []))
            self.assertIn('model_reasoning_effort="high"', result.get("candidate_argv", []))
            self.assertFalse(output.exists())

    def test_explicit_skill_text_is_separate_from_shared_task(self):
        with tempfile.TemporaryDirectory() as directory:
            path, data, task = fixture(directory)
            skill = path.parent / "skill"
            skill.mkdir()
            content = b"---\nname: scene-skill\ndescription: Fixture\n---\nInstructions.\n"
            (skill / "SKILL.md").write_bytes(content)
            digest = hashlib.sha256(b"SKILL.md\x00" + hashlib.sha256(content).digest()).hexdigest()
            entry = {"name": "scene-skill", "path": "skill", "sha256": digest, "invocation": "explicit"}
            path.write_text(json.dumps({**data, "capabilities": {**data["capabilities"], "skills": [entry]}}))
            result = preview(load_experiment(path), path.parent / "output")
            self.assertEqual(result.get("treatment"), "$scene-skill")
            self.assertEqual(result["delivered_prompt"], "$scene-skill\n\n" + task.decode())
            self.assertEqual(result["task_prompt_sha256"], hashlib.sha256(task).hexdigest())
            self.assertNotEqual(result["task_prompt_sha256"], result["prompt_sha256"])

    def test_unreviewed_profile_is_refused_before_any_output_is_created(self):
        with tempfile.TemporaryDirectory() as directory:
            path, _, _ = fixture(directory)
            output = Path(directory) / "archive" / "attempt-1"
            with self.assertRaisesRegex(UnsupportedCondition, "isolation"):
                run(load_experiment(path), output)
            self.assertFalse(output.exists())
