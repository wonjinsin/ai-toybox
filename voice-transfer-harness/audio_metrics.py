"""Objective waveform measurements and reference-pitch calibration."""

import math
from pathlib import Path

import numpy as np
import parselmouth
import soundfile as sf


TARGET_PITCH_HZ = 129.63


def measure(path: Path) -> dict:
    audio, sample_rate = sf.read(path, dtype="float64", always_2d=True)
    if audio.shape[0] == 0:
        raise RuntimeError(f"empty generated audio: {path}")
    if not np.isfinite(audio).all():
        raise RuntimeError(f"non-finite generated audio: {path}")
    flat = audio.reshape(-1)
    duration = len(audio) / sample_rate
    pitch = np.empty(0)
    if duration >= 3 / 60:
        mono = audio[:, 0] if audio.shape[1] == 1 else audio.mean(axis=1)
        pitch = parselmouth.Sound(mono, sampling_frequency=sample_rate).to_pitch_ac(
            time_step=0.01, pitch_floor=60, pitch_ceiling=400,
        ).selected_array["frequency"]
    voiced = pitch[pitch > 0]
    return {
        "duration_seconds": duration,
        "sample_rate": sample_rate,
        "channels": audio.shape[1],
        "finite_samples": True,
        "rms": float(np.sqrt(np.dot(flat, flat) / flat.size)),
        "peak": float(max(-flat.min(), flat.max())),
        "clipped_samples": int(np.count_nonzero((flat <= -1) | (flat >= 1))),
        "voiced_frame_fraction": float(voiced.size / pitch.size) if pitch.size else 0.0,
        "median_pitch_hz": float(np.median(voiced)) if voiced.size else None,
    }


def automatic_pitch(source_hz: float | None) -> int:
    if source_hz is None or not math.isfinite(source_hz) or source_hz <= 0:
        raise RuntimeError("automatic pitch requires a usable voiced source pitch; supply an integer --pitch")
    semitones = 12 * math.log2(TARGET_PITCH_HZ / source_hz)
    magnitude = abs(semitones)
    integer = math.floor(magnitude)
    fraction = magnitude - integer
    # Log-ratio evaluation can put an exact half-semitone one ULP below its tie.
    rounded = integer + int(fraction > 0.5 or math.isclose(fraction, 0.5, rel_tol=0, abs_tol=1e-12))
    return rounded if semitones >= 0 else -rounded


def pitch_settings(source_metrics: dict, requested, backend: str) -> dict:
    choice = requested if requested is not None else (-12 if backend == "say" else "auto")
    source_hz = source_metrics["median_pitch_hz"]
    return {
        "mode": "auto" if choice == "auto" else "fixed",
        "source_median_hz": source_hz,
        "target_hz": TARGET_PITCH_HZ if choice == "auto" else None,
        "semitones": automatic_pitch(source_hz) if choice == "auto" else choice,
    }
