import pathlib
import subprocess
import tempfile
import unittest


ROOT = pathlib.Path(__file__).resolve().parents[1]


class RuntimeCheckTests(unittest.TestCase):
    def test_check_reports_missing_runtime_without_creating_files(self):
        with tempfile.TemporaryDirectory() as directory:
            result = subprocess.run(
                ["bash", str(ROOT / "generate.sh"), "--check", "--runtime-root", directory],
                capture_output=True,
                text=True,
            )
            self.assertEqual(result.returncode, 1, result.stderr)
            self.assertEqual(list(pathlib.Path(directory).iterdir()), [])



if __name__ == "__main__":
    unittest.main()
