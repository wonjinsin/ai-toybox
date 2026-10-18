import math
from pathlib import Path
import tempfile
import unittest

import numpy as np
import soundfile as sf

from audio_metrics import TARGET_PITCH_HZ, automatic_pitch, measure, pitch_settings
from kokoro_source import join_audio_segments, split_phonemes


class PitchTests(unittest.TestCase):
    def test_signed_half_semitones_round_away_from_zero(self):
        for shift, expected in ((-12.5, -13), (-0.5, -1), (0.5, 1), (12.5, 13)):
            with self.subTest(shift=shift):
                source = TARGET_PITCH_HZ / 2 ** (shift / 12)
                self.assertEqual(automatic_pitch(source), expected)

    def test_silence_has_unavailable_pitch_and_requires_explicit_shift(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "silence.wav"
            sf.write(path, np.zeros(1600), 16000)
            metrics = measure(path)
        self.assertIsNone(metrics["median_pitch_hz"])
        self.assertEqual(metrics["voiced_frame_fraction"], 0)
        with self.assertRaises(RuntimeError):
            pitch_settings(metrics, "auto", "kokoro")
        self.assertEqual(pitch_settings(metrics, 0, "kokoro")["semitones"], 0)

    def test_nonfinite_audio_fails_instead_of_producing_a_quality_record(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "invalid.wav"
            sf.write(path, np.array([0.0, math.nan, 0.0]), 16000, subtype="FLOAT")
            with self.assertRaises(RuntimeError):
                measure(path)


class FullUtteranceTests(unittest.TestCase):
    def test_long_phoneme_stream_retains_order_and_every_nonspace_character(self):
        phonemes = ("ka" * 280 + ", " + "na" * 260 + ". " + "mi" * 310)
        chunks = list(split_phonemes(phonemes))
        self.assertEqual("".join(chunks).replace(" ", ""), phonemes.replace(" ", ""))
        self.assertTrue(all(0 < len(chunk) <= 510 for chunk in chunks))

    def test_join_retains_each_segment_in_order(self):
        joined, count = join_audio_segments(
            [np.array([0.1, 0.2], dtype=np.float32), np.array([-0.3], dtype=np.float32),
             np.array([0.4, 0.5], dtype=np.float32)]
        )
        np.testing.assert_array_equal(joined, np.array([0.1, 0.2, -0.3, 0.4, 0.5], dtype=np.float32))
        self.assertEqual(count, 3)


if __name__ == "__main__":
    unittest.main()
