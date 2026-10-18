import argparse
import json
from pathlib import Path

from faster_whisper import WhisperModel

root = Path(__file__).resolve().parents[2]
parser = argparse.ArgumentParser()
parser.add_argument("audio", type=Path)
parser.add_argument("--output", type=Path, required=True)
args = parser.parse_args()
model = WhisperModel(str(root / "models/whisper-small"), device="cpu", compute_type="int8", cpu_threads=4)
segments, info = model.transcribe(str(args.audio), beam_size=5, condition_on_previous_text=False)
records = [{"start": s.start, "end": s.end, "text": s.text} for s in segments]
args.output.write_text(json.dumps({"audio": str(args.audio), "language": info.language, "segments": records}, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
print(f"Saved {len(records)} timed segments to {args.output}")
