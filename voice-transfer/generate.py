import argparse
import os
from pathlib import Path
import subprocess
import sys

root = Path(__file__).resolve().parent
language_profiles = {
    "ja": {"voice": "ja-JP-KeitaNeural", "index_rate": 0.25},
    "en": {"voice": "en-US-EricNeural", "index_rate": 1.0},
}
parser = argparse.ArgumentParser(
    description="Generate Japanese or English speech with the downloaded Brad Pitt RVC model.",
    epilog=(
        "Japanese uses native Keita with index rate 0.25 to prioritize pronunciation "
        "over speaker similarity. Detected unvoiced frames retain zero pitch for synthesis and consonant protection. "
        "Bundled Japanese samples use conversational wording. Supplied text is spoken as provided, not rewritten. "
        "English uses Eric with index rate 1.0, selected for the bundled long script; "
        "similarity gains do not generalize to every sentence. "
        "Text is sent to Microsoft Edge TTS. Voice conversion runs locally. "
        "Existing output and source files are replaced."
    ),
)
parser.add_argument("text", nargs="?", help="Text to speak. Defaults to the selected language's sample file.")
parser.add_argument("--language", choices=language_profiles, default="ja")
parser.add_argument("--text-file", type=Path, help="UTF-8 file containing a longer script.")
parser.add_argument("--output", type=Path, help="Destination WAV file. Defaults to outputs/brad-pitt-LANGUAGE.wav.")
args = parser.parse_args()
if args.text is not None and args.text_file is not None:
    parser.error("Use either positional text or --text-file, not both.")
profile = language_profiles[args.language]
text_file = args.text_file.expanduser() if args.text_file is not None else root / f"sample-{args.language}.txt"
text = args.text if args.text is not None else text_file.read_text(encoding="utf-8")
if not text.strip():
    parser.error("Text must not be empty.")
output = (args.output or root / "outputs" / f"brad-pitt-{args.language}.wav").expanduser().resolve()
if output.suffix.lower() != ".wav":
    parser.error("Output must have a .wav extension.")
output.parent.mkdir(parents=True, exist_ok=True)
source_mp3 = output.with_suffix(".source.mp3")
source_wav = output.with_suffix(".source.wav")
transcript = output.with_suffix(".txt")
transcript.write_text(text.strip() + "\n", encoding="utf-8")

subprocess.run(
    [sys.executable, "-m", "edge_tts", "--voice", profile["voice"], "--file", str(transcript), "--write-media", str(source_mp3)],
    check=True,
)
subprocess.run(
    ["ffmpeg", "-nostdin", "-hide_banner", "-loglevel", "error", "-y", "-i", str(source_mp3), "-ar", "40000", "-ac", "1", str(source_wav)],
    check=True,
)
environment = os.environ.copy()
environment.update(TORCH_FORCE_WEIGHTS_ONLY_LOAD="1", OMP_NUM_THREADS="4", OPENBLAS_NUM_THREADS="1", HF_HUB_OFFLINE="1")
subprocess.run(
    [
        sys.executable, "-m", "infer.cli",
        "--model", str(root / "models" / "brad-pitt" / "BradPitt.pth"),
        "--index", str(root / "models" / "brad-pitt" / "added_IVF840_Flat_nprobe_1_BradPitt_v2.index"),
        "--input", str(source_wav),
        "--output", str(output),
        "--f0-method", "rmvpe",
        "--pitch", "0",
        "--index-rate", str(profile["index_rate"]),
        "--overwrite",
    ],
    cwd=root / "vendor" / "rvc",
    env=environment,
    check=True,
)
print(f"Converted audio: {output}")
print(f"Source audio: {source_wav}")
