import json
from pathlib import Path
import subprocess
import sys

work = Path(__file__).resolve().parent
plan = json.loads((work / "plan.json").read_text())
sources = work / "sources"
sources.mkdir(exist_ok=True)
for prompt_id, text in plan["prompts"].items():
    transcript = sources / f"{prompt_id}.txt"
    transcript.write_text(text + "\n", encoding="utf-8")
    for voice in plan["source_voices"]:
        stem = sources / f"{prompt_id}-{voice}"
        subprocess.run([sys.executable, "-m", "edge_tts", "--voice", voice, "--file", str(transcript), "--write-media", str(stem.with_suffix(".mp3"))], check=True)
        subprocess.run(["ffmpeg", "-nostdin", "-hide_banner", "-loglevel", "error", "-y", "-i", str(stem.with_suffix(".mp3")), "-ar", "40000", "-ac", "1", str(stem.with_suffix(".wav"))], check=True)
        print(stem.with_suffix(".wav"), flush=True)
