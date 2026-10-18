#!/usr/bin/env python3
"""Synthesize complete Japanese utterances using the installed local Kokoro model."""

from __future__ import annotations

import argparse
import contextlib
import hashlib
import json
import math
import os
from pathlib import Path
import re
import sys
from typing import Iterable


DEFAULT_VOICE = "jm_kumo"
MODEL_REVISION = "a71e4d38b236d968966a2002c4c895dbd12b1c3c"


def offline_environment() -> None:
    os.environ["HF_HUB_OFFLINE"] = "1"
    os.environ["TRANSFORMERS_OFFLINE"] = "1"
    os.environ["HF_HUB_DISABLE_TELEMETRY"] = "1"
    os.environ["DO_NOT_TRACK"] = "1"
    os.environ["TOKENIZERS_PARALLELISM"] = "false"
    sys.dont_write_bytecode = True


def split_phonemes(phonemes: str, limit: int = 510) -> Iterable[str]:
    """Keep every phoneme while respecting Kokoro's context window."""
    remaining = phonemes.strip()
    while len(remaining) > limit:
        window = remaining[:limit]
        boundary = max(window.rfind(mark) + 1 for mark in ".!?;, ")
        if boundary <= 0:
            boundary = limit
        chunk = remaining[:boundary].strip()
        if chunk:
            yield chunk
        remaining = remaining[boundary:].lstrip()
    if remaining:
        yield remaining


def join_audio_segments(segments: Iterable[object]):
    import numpy as np

    arrays = []
    for segment in segments:
        audio = np.asarray(segment, dtype=np.float32)
        if audio.ndim == 2 and audio.shape[0] == 1:
            audio = audio[0]
        if audio.ndim != 1 or not audio.size or not np.isfinite(audio).all():
            raise RuntimeError("Kokoro produced empty, non-mono, or non-finite audio")
        arrays.append(audio)
    if not arrays:
        raise RuntimeError("Kokoro produced no speech segments for the supplied text")
    return np.concatenate(arrays), len(arrays)


def load_local_runtime(root: Path, voice: str = DEFAULT_VOICE):
    offline_environment()
    if voice != DEFAULT_VOICE:
        raise RuntimeError(f"Unsupported Kokoro voice {voice!r}; the installed Japanese voice is {DEFAULT_VOICE}")
    model_dir = root / ".runtime" / "kokoro"
    required = [model_dir / "config.json", model_dir / "kokoro-v1_0.safetensors", model_dir / "voices" / f"{voice}.safetensors"]
    for path in required:
        if not path.is_file():
            raise RuntimeError(f"Kokoro runtime missing asset: {path}; run bootstrap.py --runtime-root {root}")
    try:
        import unidic_lite
        from mlx_audio.tts.models.kokoro.pipeline import KokoroPipeline
        from mlx_audio.tts.utils import load_model
    except ImportError as exc:
        raise RuntimeError(f"Kokoro runtime missing dependency: {exc}; run bootstrap.py --runtime-root {root}") from exc

    manifest_path = Path(__file__).with_name("artifacts-kokoro.json")
    manifest = json.loads(manifest_path.read_text(encoding="utf-8"))
    dictionary_dir = Path(unidic_lite.DICDIR)
    dictionary = manifest["sources"]["japanese_processing"]
    for artifact in dictionary["dictionary_files"]:
        path = dictionary_dir / artifact["path"]
        if not path.is_file():
            raise RuntimeError(f"Kokoro runtime missing Japanese dictionary asset: {path}; run bootstrap.py --runtime-root {root}")
        with path.open("rb") as handle:
            observed = hashlib.file_digest(handle, "sha256").hexdigest()
        if observed != artifact["sha256"]:
            raise RuntimeError(f"Kokoro dictionary checksum mismatch: {path}; expected {artifact['sha256']}, observed {observed}; restore the pinned runtime")

    # Passing a Path and a voice filename bypasses Hugging Face repository lookup.
    model = load_model(model_dir, model_type="kokoro")
    pipeline = KokoroPipeline(lang_code="j", model=model, repo_id=str(model_dir))
    voice_path = str(model_dir / "voices" / f"{voice}.safetensors")
    pipeline.load_voice(voice_path)
    if not pipeline.g2p("日本語の音声を確認します。"):
        raise RuntimeError("Kokoro Japanese dictionary produced no phonemes; restore the pinned runtime")
    return model, pipeline, voice_path


def check_runtime(root: Path) -> dict:
    model, _, _ = load_local_runtime(root)
    return {"backend": "kokoro", "model_revision": MODEL_REVISION, "voice": DEFAULT_VOICE, "language": "j", "sample_rate": model.config.sample_rate, "inference_dtype": "float32", "offline": True}


def synthesize(root: Path, text: str, output: Path, voice: str = DEFAULT_VOICE, speed: float = 1.0) -> dict:
    if not text.strip():
        raise RuntimeError("Kokoro input text must not be empty")
    if not math.isfinite(speed) or speed <= 0:
        raise RuntimeError("Kokoro speed must be a finite positive number")
    if output.exists():
        raise RuntimeError(f"Refusing to overwrite existing Kokoro output: {output}")
    model, pipeline, voice_path = load_local_runtime(root, voice)

    def audio_segments():
        for paragraph in re.split(r"\n+", text):
            if not paragraph.strip():
                continue
            phonemes = pipeline.g2p(paragraph)
            if not phonemes:
                raise RuntimeError(f"Kokoro could not phonemize input paragraph: {paragraph!r}")
            # Upstream's text pipeline truncates long Japanese phonemes. Use its
            # bounded phoneme API instead, retaining every chunk in source order.
            for chunk in split_phonemes(phonemes):
                for result in pipeline.generate_from_tokens(chunk, voice=voice_path, speed=speed):
                    if result.audio is None:
                        raise RuntimeError("Kokoro produced a segment without audio")
                    yield result.audio

    audio, count = join_audio_segments(audio_segments())
    import soundfile as sf

    output.parent.mkdir(parents=True, exist_ok=True)
    sf.write(str(output), audio, model.config.sample_rate, subtype="FLOAT", format="WAV")
    return {"backend": "kokoro", "voice": voice, "language": "j", "speed": speed, "sample_rate": model.config.sample_rate, "segments": count, "samples": int(audio.size), "model_revision": MODEL_REVISION, "inference_dtype": "float32", "output": str(output), "offline": True}


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--runtime-root", required=True, type=Path)
    parser.add_argument("--check", action="store_true")
    parser.add_argument("--text-file", type=Path)
    parser.add_argument("--output", type=Path)
    parser.add_argument("--voice", default=DEFAULT_VOICE)
    parser.add_argument("--speed", type=float, default=1.0)
    args = parser.parse_args(argv)
    if not args.check and (args.text_file is None or args.output is None):
        parser.error("generation requires --text-file and --output")
    try:
        with contextlib.redirect_stdout(sys.stderr):
            if args.check:
                metadata = check_runtime(args.runtime_root.resolve())
            else:
                metadata = synthesize(args.runtime_root.resolve(), args.text_file.read_text(encoding="utf-8"), args.output.resolve(), args.voice, args.speed)
        print(json.dumps(metadata, ensure_ascii=False))
        return 0
    except (OSError, ValueError, RuntimeError, ImportError) as exc:
        print(f"Kokoro error: {exc}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
