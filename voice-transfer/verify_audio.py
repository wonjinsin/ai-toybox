import argparse
import json
from pathlib import Path

import numpy as np
import soundfile as sf
from faster_whisper import WhisperModel

root = Path(__file__).resolve().parent
parser = argparse.ArgumentParser(description="Check generated speech with independent language detection and transcription.")
parser.add_argument("audio", nargs="+", type=Path)
parser.add_argument("--report", type=Path, default=root / "outputs" / "verification.json")
parser.add_argument("--language", choices=("ja", "en"), default="ja", help="Expected language; not supplied to the recognizer.")
args = parser.parse_args()

model = WhisperModel(str(root / "models" / "whisper-small"), device="cpu", compute_type="int8", cpu_threads=4)
records = []
for path in args.audio:
    audio, sample_rate = sf.read(path, dtype="float32")
    if audio.size == 0 or not np.isfinite(audio).all():
        raise ValueError(f"Invalid audio samples in {path}")
    rms = float(np.sqrt(np.mean(audio.astype(np.float64) ** 2)))
    if rms < 0.0001:
        raise ValueError(f"Audio is silent or too quiet in {path}")
    segments, info = model.transcribe(str(path), beam_size=5, condition_on_previous_text=False)
    transcript = "".join(segment.text for segment in segments).strip()
    record = {
        "file": str(path.resolve()),
        "sample_rate": sample_rate,
        "duration_seconds": len(audio) / sample_rate,
        "peak": float(np.max(np.abs(audio))),
        "rms": rms,
        "detected_language": info.language,
        "language_probability": info.language_probability,
        "transcript": transcript,
    }
    records.append(record)
    print(json.dumps(record, ensure_ascii=False, indent=2), flush=True)

args.report.parent.mkdir(parents=True, exist_ok=True)
args.report.write_text(json.dumps(records, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
if any(record["detected_language"] != args.language or not record["transcript"] for record in records):
    raise SystemExit(f"Expected {args.language} speech was not recognized in every file. Inspect the report.")
