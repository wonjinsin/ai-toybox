"""Compare local speech clips with ECAPA speaker embeddings.

Write a JSON array to --output and one JSON record per candidate to stdout.
Cosine scores are similarity proxies, not identity percentages or calibrated
probabilities. Use independent, same-speaker held-out clips as --audio to
calibrate the reference set. Do not compare a clip to itself for calibration.

Required files in models/speaker-ecapa: hyperparams.yaml, embedding_model.ckpt,
classifier.ckpt, mean_var_norm_emb.ckpt, and label_encoder.txt. Download these
separately from speechbrain/spkrec-ecapa-voxceleb before running this script.
Only trusted model files should be used; HyperPyYAML constructs Python objects.
"""

import argparse
from dataclasses import dataclass
import hashlib
from importlib.metadata import version
import json
import math
from pathlib import Path

import numpy as np
import soundfile as sf
import torch
from hyperpyyaml import load_hyperpyyaml
from speechbrain.inference.classifiers import EncoderClassifier
from torchaudio.functional import resample


ROOT = Path(__file__).resolve().parent
MODEL_DIR = ROOT / "models" / "speaker-ecapa"
MODEL_ID = "speechbrain/spkrec-ecapa-voxceleb"
SAMPLE_RATE = 16000
MIN_SECONDS = 3
MIN_RMS = 0.0001
ASSETS = {
    "embedding_model": "embedding_model.ckpt",
    "mean_var_norm_emb": "mean_var_norm_emb.ckpt",
    "classifier": "classifier.ckpt",
    "label_encoder": "label_encoder.txt",
}


@dataclass
class EmbeddedClip:
    metadata: dict
    embeddings: torch.Tensor


def load_model():
    paths = {name: MODEL_DIR / filename for name, filename in ASSETS.items()}
    for path in (MODEL_DIR / "hyperparams.yaml", *paths.values()):
        if not path.is_file():
            raise FileNotFoundError(f"Required local model asset is missing: {path}")
    with (MODEL_DIR / "hyperparams.yaml").open(encoding="utf-8") as stream:
        hparams = load_hyperpyyaml(stream, {"pretrained_path": str(MODEL_DIR)})
    pretrainer = hparams["pretrainer"]
    if set(pretrainer.loadables) != set(paths):
        raise ValueError("Local hyperparams.yaml does not match the expected ECAPA model assets")
    # Bypass collection and fetching entirely, including optional custom.py fetches.
    pretrainer.paths = {name: str(path) for name, path in paths.items()}
    pretrainer.is_local = list(paths)
    pretrainer.load_collected()
    model = EncoderClassifier(
        modules=hparams["modules"], hparams=hparams,
        run_opts={"device": "cpu"}, freeze_params=True,
    ).float().eval()
    hashes = {}
    for filename in ("hyperparams.yaml", "embedding_model.ckpt"):
        with (MODEL_DIR / filename).open("rb") as stream:
            hashes[filename] = hashlib.file_digest(stream, "sha256").hexdigest()
    identity = {
        "id": MODEL_ID,
        "local_directory": str(MODEL_DIR),
        "sha256": hashes,
        "speechbrain_version": version("speechbrain"),
        "torch_version": version("torch"),
        "torchaudio_version": version("torchaudio"),
        "device": "cpu",
        "dtype": "float32",
        "threads": 4,
        "sample_rate": SAMPLE_RATE,
        "embedding_normalization": "L2; encode_batch(normalize=False)",
    }
    return model, identity


@torch.inference_mode()
def embed_clip(path: Path, model: EncoderClassifier, window_samples: int) -> EmbeddedClip:
    if not path.is_file():
        raise FileNotFoundError(f"Audio file is missing: {path}")
    audio, sample_rate = sf.read(path, dtype="float32", always_2d=True)
    if sample_rate <= 0 or audio.size == 0 or not np.isfinite(audio).all():
        raise ValueError(f"Invalid or empty audio samples in {path}")
    duration = len(audio) / sample_rate
    if duration < MIN_SECONDS:
        raise ValueError(f"Audio must be at least {MIN_SECONDS} seconds: {path} ({duration:.3f}s)")
    waveform = torch.from_numpy(audio.mean(axis=1))
    if sample_rate != SAMPLE_RATE:
        waveform = resample(waveform, sample_rate, SAMPLE_RATE)
    if not torch.isfinite(waveform).all():
        raise ValueError(f"Non-finite samples after mono conversion/resampling in {path}")
    rms = float(waveform.square().mean().sqrt())
    if not math.isfinite(rms) or rms < MIN_RMS:
        raise ValueError(f"Audio is silent or too quiet in {path} (RMS {rms})")
    windows = []
    embeddings = []
    for start in range(0, waveform.numel(), window_samples):
        end = min(start + window_samples, waveform.numel())
        if end - start < MIN_SECONDS * SAMPLE_RATE:
            break
        samples = waveform[start:end]
        window_rms = float(samples.square().mean().sqrt())
        if not math.isfinite(window_rms) or window_rms < MIN_RMS:
            raise ValueError(
                f"Silent or too-quiet window in {path} at {start / SAMPLE_RATE:.3f}s "
                f"(RMS {window_rms})"
            )
        embedding = model.encode_batch(samples.unsqueeze(0), normalize=False).reshape(-1)
        norm = torch.linalg.vector_norm(embedding)
        if not torch.isfinite(embedding).all() or not torch.isfinite(norm) or norm <= 0:
            raise ValueError(f"Invalid speaker embedding for {path} at {start / SAMPLE_RATE:.3f}s")
        embeddings.append(embedding / norm)
        windows.append({
            "index": len(windows),
            "start_seconds": start / SAMPLE_RATE,
            "end_seconds": end / SAMPLE_RATE,
            "duration_seconds": (end - start) / SAMPLE_RATE,
            "rms": window_rms,
        })
    if not embeddings:
        raise ValueError(f"No windows of at least {MIN_SECONDS} seconds in {path}")
    metadata = {
        "file": str(path),
        "sample_rate": sample_rate,
        "channels": audio.shape[1],
        "duration_seconds": duration,
        "rms": rms,
        "number_of_windows": len(windows),
        "discarded_tail_seconds": (waveform.numel() / SAMPLE_RATE) - windows[-1]["end_seconds"],
        "windows": windows,
    }
    return EmbeddedClip(metadata, torch.stack(embeddings))


def score_candidate(candidate: EmbeddedClip, references: list[EmbeddedClip], identity: dict,
                    window_samples: int) -> dict:
    per_reference = []
    pooled = []
    for reference in references:
        scores = (reference.embeddings @ candidate.embeddings.T).clamp(-1, 1).numpy()
        pooled.append(scores.reshape(-1))
        per_reference.append({
            **reference.metadata,
            "same_file_as_candidate": reference.metadata["file"] == candidate.metadata["file"],
            "median_cosine_similarity": float(np.median(scores)),
            "cosine_similarity_range": [float(scores.min()), float(scores.max())],
            "number_of_window_pairs": int(scores.size),
            "window_pair_cosine_similarities": scores.tolist(),
        })
    all_scores = np.concatenate(pooled)
    return {
        **candidate.metadata,
        "model": identity,
        "score_interpretation": "Cosine similarity proxy, not an identity percentage or probability.",
        "aggregation": "Median over every reference-window/candidate-window pair; longer references have more pairs.",
        "window_pair_matrix_axes": ["reference_window_index", "candidate_window_index"],
        "window_seconds": window_samples / SAMPLE_RATE,
        "minimum_window_seconds": MIN_SECONDS,
        "median_cosine_similarity": float(np.median(all_scores)),
        "cosine_similarity_range": [float(all_scores.min()), float(all_scores.max())],
        "number_of_references": len(references),
        "number_of_reference_windows": sum(len(reference.embeddings) for reference in references),
        "number_of_window_pairs": int(all_scores.size),
        "per_reference": per_reference,
    }


def main():
    parser = argparse.ArgumentParser(
        description="Compare local speech with offline ECAPA cosine similarity proxies, not identity percentages.",
        epilog="For calibration, pass independent same-speaker held-out recordings as --audio. "
               "Final windows shorter than 3 seconds are excluded. Silent retained windows are errors.",
    )
    parser.add_argument("--references", nargs="+", type=Path, required=True)
    parser.add_argument("--audio", nargs="+", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True, help="JSON array report; also emits one record per candidate to stdout.")
    parser.add_argument("--window-seconds", type=float, default=6.0)
    args = parser.parse_args()
    if not math.isfinite(args.window_seconds) or args.window_seconds < MIN_SECONDS:
        parser.error("--window-seconds must be finite and at least 3")
    window_samples = round(args.window_seconds * SAMPLE_RATE)
    references = [path.resolve() for path in args.references]
    candidates = [path.resolve() for path in args.audio]
    output = args.output.resolve()
    if output in references + candidates or output.is_relative_to(MODEL_DIR.resolve()):
        parser.error("--output must not overwrite input audio or local model assets")
    torch.set_num_threads(4)
    model, identity = load_model()
    clips = {}
    for path in dict.fromkeys(references + candidates):
        clips[path] = embed_clip(path, model, window_samples)
    reference_clips = [clips[path] for path in references]
    records = [score_candidate(clips[path], reference_clips, identity, window_samples) for path in candidates]
    output.parent.mkdir(parents=True, exist_ok=True)
    output.write_text(json.dumps(records, ensure_ascii=False, indent=2, allow_nan=False) + "\n", encoding="utf-8")
    for record in records:
        print(json.dumps(record, ensure_ascii=False, allow_nan=False), flush=True)


if __name__ == "__main__":
    main()
