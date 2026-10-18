import hashlib
import json
from pathlib import Path
import tempfile
import unittest
import zipfile

from bootstrap import extract_artifact
from harness import listen
from runtime import safe_path, verify_file


class IntegrityTests(unittest.TestCase):
    def test_artifact_paths_cannot_escape_through_parent_or_symlink(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "link").symlink_to(root.parent, target_is_directory=True)
            for value in ("../outside", "/absolute", "link/outside", "..\\outside"):
                with self.subTest(path=value), self.assertRaises(RuntimeError):
                    safe_path(root, value)

    def test_checksum_mismatch_preserves_existing_bytes(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "model"
            path.write_bytes(b"user-owned bytes")
            with self.assertRaises(RuntimeError):
                verify_file(path, hashlib.sha256(b"expected bytes").hexdigest())
            self.assertEqual(path.read_bytes(), b"user-owned bytes")

    def test_archive_rejects_traversal_before_promoting_valid_members(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            archive = root / "unsafe.zip"
            staging = root / "staging"
            staging.mkdir()
            with zipfile.ZipFile(archive, "w") as package:
                package.writestr("model", b"verified")
                package.writestr("../escape", b"untrusted")
            artifact = {"format": "zip", "members": [
                {"member": "model", "path": "installed/model",
                 "sha256": hashlib.sha256(b"verified").hexdigest()},
            ]}
            with self.assertRaises(RuntimeError):
                extract_artifact(root, artifact, archive, staging)
            self.assertFalse((root / "installed").exists())

    def test_playback_rejects_audio_outside_its_run(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            outside = root / "outside.wav"
            outside.write_bytes(b"not to be opened")
            run = root / "run"
            run.mkdir()
            (run / "manifest.json").write_text(json.dumps({
                "schema_version": 1, "sources": [], "candidates": [
                    {"id": "escape", "status": "complete", "path": "../outside.wav"},
                ],
            }))
            with self.assertRaises(RuntimeError):
                listen(run)
            self.assertFalse((run / "preference.json").exists())


if __name__ == "__main__":
    unittest.main()
