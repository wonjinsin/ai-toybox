from pathlib import Path
import sys
from types import SimpleNamespace
import unittest

import numpy as np
import torch

sys.path.insert(0, str(Path(__file__).resolve().parent / "vendor" / "rvc"))

from infer.module.models import SineGen
from infer.vc.pipeline import Pipeline


class VoicingTests(unittest.TestCase):
    def test_pitch_shift_preserves_unvoiced_synthesis_frames(self):
        frequencies = np.array([0, 100, 200, 0, 0, 150, 0], dtype=np.float32)
        config = SimpleNamespace(
            x_pad=1, x_query=1, x_center=1, x_max=1,
            is_half=False, device="cpu",
        )
        pipeline = Pipeline(32000, config)
        pipeline.model_rmvpe = SimpleNamespace(
            infer_from_audio=lambda audio, thred: frequencies.copy(),
        )
        synthesizer = SineGen(32000)
        for shift, expected in (
            (0, [0, 100, 200, 0, 0, 150, 0]),
            (12, [0, 200, 400, 0, 0, 300, 0]),
            (-12, [0, 50, 100, 0, 0, 75, 0]),
        ):
            with self.subTest(semitones=shift):
                coarse, continuous = pipeline.get_f0(
                    np.zeros(1120, dtype=np.float32), 7, shift, "rmvpe",
                )
                np.testing.assert_allclose(continuous, expected)
                np.testing.assert_array_equal(coarse[[0, 3, 4, 6]], [1, 1, 1, 1])
                _, voiced, _ = synthesizer(torch.from_numpy(continuous)[None], upp=2)
                np.testing.assert_array_equal(
                    voiced.numpy().reshape(-1),
                    [0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 1, 1, 0, 0],
                )


if __name__ == "__main__":
    unittest.main()
