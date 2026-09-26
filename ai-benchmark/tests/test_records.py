import hashlib
import json
import tempfile
import unittest
from pathlib import Path

from ai_benchmark.config import load_experiment, skill_digest
from ai_benchmark.records import prepare
from test_config import fixture


class RecordTests(unittest.TestCase):
    def test_prepare_freezes_inputs_and_leaves_observations_unknown(self):
        with tempfile.TemporaryDirectory() as directory:
            path, data, prompt = fixture(directory)
            output = Path(directory) / "attempt-1"
            record = prepare(load_experiment(path), output)
            self.assertEqual(record.get("record_status"), "prepared")
            self.assertEqual((output / "prompt.md").read_bytes(), prompt)
            self.assertEqual((output / "task-prompt.md").read_bytes(), prompt)
            self.assertEqual((output / "experiment.json").read_bytes(), path.read_bytes())
            self.assertEqual(record["experiment"]["sha256"], hashlib.sha256(path.read_bytes()).hexdigest())
            self.assertEqual(record["requested"]["model"], data["model"])
            self.assertEqual(record["requested"]["tools"], [])
            self.assertEqual(record["run_id"], "attempt-1")
            self.assertEqual(record["execution"]["status"], "not_started")
            self.assertIsNone(record["reported"]["model"])
            self.assertIsNone(record["cli"]["version"])
            self.assertEqual(record["context"]["isolation_status"], "unverified")
            self.assertEqual(record["capture"]["response"]["status"], "not_collected")
            self.assertEqual((output / "response.txt").read_bytes(), b"")
            self.assertEqual((output / "errors.log").read_bytes(), b"")
            self.assertFalse((output / "treatment.md").exists())
            self.assertEqual(json.loads((output / "run.json").read_bytes()), record)
            self.assertFalse(json.loads((output / "preview.json").read_bytes())["launch_ready"])

    def test_existing_archive_cannot_be_overwritten(self):
        with tempfile.TemporaryDirectory() as directory:
            path, _, _ = fixture(directory)
            output = Path(directory) / "attempt-1"
            experiment = load_experiment(path)
            prepare(experiment, output)
            before = {p.name: p.read_bytes() for p in output.iterdir()}
            with self.assertRaises(FileExistsError):
                prepare(experiment, output)
            self.assertEqual(before, {p.name: p.read_bytes() for p in output.iterdir()})

    def test_skill_treatment_is_preserved_separately(self):
        with tempfile.TemporaryDirectory() as directory:
            path, data, task = fixture(directory)
            skill = path.parent / "skill"
            skill.mkdir()
            (skill / "SKILL.md").write_text("Fixture")
            entry = {"name": "scene-skill", "path": "skill", "sha256": skill_digest(skill), "invocation": "explicit"}
            path.write_text(json.dumps({**data, "capabilities": {**data["capabilities"], "skills": [entry]}}))
            output = Path(directory) / "attempt-1"
            record = prepare(load_experiment(path), output)
            self.assertTrue((output / "treatment.md").exists())
            self.assertEqual((output / "treatment.md").read_bytes(), b"$scene-skill")
            self.assertEqual((output / "prompt.md").read_bytes(), b"$scene-skill\n\n" + task)
            self.assertEqual(record["prompt"]["treatment_sha256"], hashlib.sha256(b"$scene-skill").hexdigest())
