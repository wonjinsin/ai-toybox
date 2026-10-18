import json
from pathlib import Path
import tempfile
from types import SimpleNamespace
import unittest
from unittest.mock import patch

import numpy as np
import soundfile as sf

import harness


class RunStateTests(unittest.TestCase):
    def test_unvoiced_calibration_failure_retains_measured_source(self):
        def silent_source(root, directory, text_file, backend, stem):
            path = directory / f"{stem}.wav"
            sf.write(path, np.zeros(1600), 16000)
            return path, {"voice": "test-source"}

        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            args = SimpleNamespace(source="say", compare=False, index_rate=0.75,
                                   pitch="auto", text="。")
            with patch.object(harness, "ROOT", root), patch("runtime.check_runtime", return_value={}), \
                    patch.object(harness, "synthesize", side_effect=silent_source):
                with self.assertRaises(RuntimeError):
                    harness.generate(args, root)
            manifest_path, = (root / "outputs").glob("*/manifest.json")
            manifest = json.loads(manifest_path.read_text())
            self.assertEqual(manifest["status"], "failed")
            source = manifest["sources"][0]
            self.assertIn("metrics", source)
            self.assertIsNone(source["metrics"]["median_pitch_hz"])
            self.assertTrue((manifest_path.parent / source["path"]).is_file())
            self.assertEqual(manifest["candidates"][0]["status"], "pending")
            self.assertFalse((manifest_path.parent / manifest["candidates"][0]["path"]).exists())


if __name__ == "__main__":
    unittest.main()
