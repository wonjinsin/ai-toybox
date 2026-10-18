#!/usr/bin/env python3
"""Local Japanese speech conversion and comparison."""

import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import sys
import tempfile
import time
from datetime import datetime, timezone


ROOT = Path(__file__).resolve().parent


def pitch_argument(value):
    if value == "auto":
        return value
    try:
        return int(value)
    except ValueError:
        raise argparse.ArgumentTypeError("pitch must be 'auto' or an integer") from None


def parser():
    command = argparse.ArgumentParser(description=__doc__)
    command.add_argument("text", nargs="?", help="Japanese text to synthesize")
    command.add_argument("--runtime-root", type=Path, default=ROOT)
    modes = command.add_mutually_exclusive_group()
    modes.add_argument("--check", action="store_true", help="Verify both backends without modifying installation")
    modes.add_argument("--compare", action="store_true", help="Compare two sources at three index rates")
    modes.add_argument("--listen", type=Path, metavar="RUN", help="Interactively play an existing run")
    command.add_argument("--play", action="store_true", help="Open playback after generation")
    command.add_argument("--index-rate", type=float, default=None)
    command.add_argument("--pitch", type=pitch_argument, default=None)
    command.add_argument("--source", choices=("say", "kokoro"), default=None)
    return command


def run_command(arguments, *, env=None):
    result = subprocess.run(arguments, env=env, capture_output=True, text=True)
    if result.returncode:
        raise RuntimeError(
            f"{Path(str(arguments[0])).name} exited {result.returncode}: "
            f"{result.stderr.strip() or result.stdout.strip()}"
        )
    return result.stdout


def reed_voice():
    inventory = run_command(["say", "-v", "?"])
    for line in inventory.splitlines():
        match = re.match(r"^(.*?)\s+ja_JP\s+#", line)
        if match and re.match(r"^Reed(?:\s|\(|$)", match[1]):
            return match[1].strip()
    raise RuntimeError("missing Japanese Reed system voice; install it in macOS voice settings")


def convert(root, source, output, pitch, index_rate):
    from runtime import rvc_environment

    run_command([
        str(root / ".venv/bin/python"), "-B", "-m", "infer.cli",
        "--model", str(root / "models/BradPitt.pth"),
        "--input", str(source), "--output", str(output),
        "--index", str(root / "models/added_IVF840_Flat_nprobe_1_BradPitt_v2.index"),
        "--speaker-id", "0", "--f0-method", "rmvpe",
        "--pitch", str(pitch), "--index-rate", str(index_rate),
    ], env=rvc_environment(root))


def file_hash(path):
    with path.open("rb") as stream:
        return hashlib.file_digest(stream, "sha256").hexdigest()


def save_json(path, value):
    temporary = path.with_suffix(path.suffix + ".tmp")
    temporary.write_text(json.dumps(value, ensure_ascii=False, indent=2, allow_nan=False) + "\n",
                         encoding="utf-8")
    temporary.replace(path)


def synthesize(root, directory, text_file, backend, stem="source"):
    if backend == "say":
        voice = reed_voice()
        native = directory / f"{stem}.aiff"
        run_command(["say", "-v", voice, "-r", "165", "-o", str(native), "-f", str(text_file)])
        settings = {"voice": voice, "rate": 165}
    else:
        native = directory / f"{stem}-native.wav"
        from runtime import rvc_environment

        env = rvc_environment(root)
        # Keep RVC's Python module path and OpenMP library selection out of MLX.
        env.pop("PYTHONPATH", None)
        env.pop("DYLD_LIBRARY_PATH", None)
        run_command([
            str(root / ".venv-kokoro/bin/python"), "-B", str(ROOT / "kokoro_source.py"),
            "--runtime-root", str(root), "--text-file", str(text_file),
            "--output", str(native), "--voice", "jm_kumo", "--speed", "1.0",
        ], env=env)
        settings = {"voice": "jm_kumo", "language": "j", "speed": 1.0}
    source = directory / f"{stem}.wav"
    run_command(["ffmpeg", "-nostdin", "-hide_banner", "-loglevel", "error", "-n",
                 "-i", str(native), "-ar", "16000", "-ac", "1", str(source)])
    return source, settings


def generate(args, root):
    from audio_metrics import measure, pitch_settings
    from runtime import check_runtime

    runtime_info = check_runtime(root, backend="kokoro" if args.compare else args.source)
    backends = ("say", "kokoro") if args.compare else (args.source,)
    rates = (0.25, 0.5, 0.75) if args.compare else (args.index_rate,)
    outputs = ROOT / "outputs"
    outputs.mkdir(exist_ok=True)
    directory = Path(tempfile.mkdtemp(prefix="run.", dir=outputs))
    text_file = directory / "input.txt"
    text_file.write_text(args.text, encoding="utf-8")
    manifest_path = directory / "manifest.json"
    sources = [{"id": backend, "backend": backend, "status": "pending"} for backend in backends]
    candidates = [
        {"id": f"{backend}-index-{rate:g}", "source_id": backend, "index_rate": rate,
         "path": f"{backend}-index-{rate:g}.wav" if args.compare else "brad-pitt-ja.wav",
         "status": "pending"}
        for backend in backends for rate in rates
    ]
    manifest = {
        "schema_version": 1, "created_at": datetime.now(timezone.utc).isoformat(),
        "text": args.text, "status": "running", "runtime": runtime_info,
        "sources": sources, "candidates": candidates,
    }
    save_json(manifest_path, manifest)
    started = time.monotonic()
    active = sources[0]
    try:
        print(f"Run: {directory}", flush=True)
        for source_record in sources:
            active = source_record
            backend = source_record["backend"]
            source_record["status"] = "running"
            save_json(manifest_path, manifest)
            source_start = time.monotonic()
            print(f"Synthesizing {backend} source...", flush=True)
            stem = f"{backend}-source" if args.compare else "source"
            source, settings = synthesize(root, directory, text_file, backend, stem)
            source_metrics = measure(source)
            source_record.update(settings, path=source.name, sha256=file_hash(source),
                                 metrics=source_metrics,
                                 elapsed_seconds=time.monotonic() - source_start)
            pitch = pitch_settings(source_metrics, args.pitch, backend)
            source_record.update(pitch=pitch, status="complete")
            for candidate in candidates:
                if candidate["source_id"] != backend:
                    continue
                active = candidate
                index_rate = candidate["index_rate"]
                candidate.update(
                    status="running", source_sha256=source_record["sha256"],
                    settings={"pitch": pitch["semitones"], "index_rate": index_rate,
                              "f0_method": "rmvpe", "speaker_id": 0, "rms_mix_rate": 1.0,
                              "protect": 0.33, "resample_sr": 0, "device": "cpu", "precision": "float32"},
                )
                save_json(manifest_path, manifest)
                conversion_start = time.monotonic()
                output = directory / candidate["path"]
                print(f"Converting {candidate['id']} (pitch {pitch['semitones']:+d})...", flush=True)
                convert(root, source, output, pitch["semitones"], index_rate)
                output_metrics = measure(output)
                candidate.update(status="complete", metrics=output_metrics,
                                 elapsed_seconds=time.monotonic() - conversion_start,
                                 duration_difference_seconds=output_metrics["duration_seconds"] - source_metrics["duration_seconds"])
                save_json(manifest_path, manifest)
        manifest["status"] = "complete"
        print(f"\n{'Candidate':24} {'Seconds':>8} {'Pitch Hz':>9} {'Clipped':>8} {'Delta s':>9}")
        for candidate in candidates:
            metrics = candidate["metrics"]
            pitch_hz = metrics["median_pitch_hz"]
            pitch_label = "n/a" if pitch_hz is None else f"{pitch_hz:.2f}"
            print(f"{candidate['id']:24} {metrics['duration_seconds']:8.3f} {pitch_label:>9} "
                  f"{metrics['clipped_samples']:8d} {candidate['duration_difference_seconds']:+9.3f}")
        for source_record in sources:
            print(f"Source: {directory / source_record['path']}")
        for candidate in candidates:
            print(f"Converted: {directory / candidate['path']}")
        print(f"Manifest: {manifest_path}")
        return directory
    except (Exception, KeyboardInterrupt) as error:
        active["status"] = "interrupted" if isinstance(error, KeyboardInterrupt) else "failed"
        active["error"] = str(error) or type(error).__name__
        manifest["status"] = active["status"]
        raise
    finally:
        manifest["elapsed_seconds"] = time.monotonic() - started
        save_json(manifest_path, manifest)


def listen(directory):
    directory = directory.expanduser().resolve()
    manifest_path = directory / "manifest.json"
    manifest = json.loads(manifest_path.read_text(encoding="utf-8"))
    if not isinstance(manifest, dict) or manifest.get("schema_version") != 1:
        raise RuntimeError("unsupported run manifest")
    entries = []
    for kind, key in (("source", "sources"), ("candidate", "candidates")):
        records = manifest.get(key)
        if not isinstance(records, list):
            raise RuntimeError(f"invalid manifest {key}")
        for record in records:
            if not isinstance(record, dict):
                raise RuntimeError(f"invalid manifest {key} entry")
            if record.get("status") != "complete":
                continue
            if not isinstance(record.get("path"), str) or not isinstance(record.get("id"), str):
                raise RuntimeError("invalid playable entry")
            path = (directory / record["path"]).resolve()
            if not path.is_relative_to(directory) or not path.is_file():
                raise RuntimeError(f"missing or unsafe audio path: {record['path']}")
            entries.append((kind, record, path))
    if not entries:
        raise RuntimeError("run has no complete audio to play")
    print(f"Run status: {manifest.get('status', 'unknown')}")
    for number, (kind, record, _) in enumerate(entries, 1):
        print(f"{number}: {kind} {record['id']}")
    print("Enter a number to play, r to replay, prefer NUMBER to record a candidate, or q to exit.")
    previous = None
    while True:
        try:
            choice = input("listen> ").strip()
        except EOFError:
            return
        if choice == "q":
            return
        if choice == "r":
            if previous is None:
                print("Play an entry before replaying.")
                continue
            number = previous
        else:
            selecting = choice.startswith("prefer ")
            token = choice[7:].strip() if selecting else choice
            if not token.isascii() or not token.isdecimal() or not 1 <= int(token) <= len(entries):
                print("Choose one of the listed numbers, r, prefer NUMBER, or q.")
                continue
            number = int(token)
            if selecting:
                kind, record, _ = entries[number - 1]
                if kind != "candidate":
                    print("A preference must select a converted candidate, not a source.")
                    continue
                save_json(directory / "preference.json", {
                    "schema_version": 1, "candidate_id": record["id"],
                    "selected_at": datetime.now(timezone.utc).isoformat(),
                    "manifest_sha256": file_hash(manifest_path),
                })
                print(f"Preference saved: {record['id']}. Generation defaults are unchanged.")
                continue
        _, record, path = entries[number - 1]
        print(f"Playing {record['id']}", flush=True)
        run_command(["afplay", str(path)])
        previous = number


def main(argv=None):
    command = parser()
    args = command.parse_args(argv)
    if args.index_rate is not None and not 0 <= args.index_rate <= 1:
        command.error("index rate must be between 0 and 1")
    if args.compare and (args.source is not None or args.index_rate is not None):
        command.error("comparison fixes both sources and index rates; omit --source and --index-rate")
    if (args.check or args.listen) and (
            args.text is not None or args.source is not None or args.pitch is not None
            or args.index_rate is not None or args.play):
        command.error("--check and --listen cannot be combined with generation settings")
    if (args.play or args.listen) and not sys.stdin.isatty():
        command.error("playback requires an interactive terminal")
    args.source = args.source or "say"
    args.index_rate = 0.75 if args.index_rate is None else args.index_rate
    if not args.check and not args.listen and (args.text is None or not args.text.strip()):
        command.error("Japanese text must not be empty")
    root = args.runtime_root.expanduser().resolve()
    try:
        if args.listen:
            listen(args.listen)
            return 0
        if not (root / ".venv/bin/python").is_file():
            raise RuntimeError(f"missing RVC environment: {root / '.venv'}; run bootstrap.py --runtime-root {root}")
        if args.check:
            from runtime import check_runtime
            print(json.dumps(check_runtime(root, backend="kokoro"), ensure_ascii=False, indent=2))
            return 0
        from runtime import rvc_environment
        if (Path(sys.prefix).resolve() != (root / ".venv").resolve()
                or os.environ.get("VOICE_HARNESS_RUNTIME") != str(root)):
            env = rvc_environment(root)
            env["VOICE_HARNESS_RUNTIME"] = str(root)
            python = str(root / ".venv/bin/python")
            os.execve(python, [python, "-B", str(Path(__file__).resolve()),
                              *(sys.argv[1:] if argv is None else argv)], env)
        directory = generate(args, root)
        if args.play:
            listen(directory)
        return 0
    except (RuntimeError, OSError, ValueError) as error:
        print(f"voice-harness: {error}", file=sys.stderr)
        return 1
    except KeyboardInterrupt:
        print("voice-harness: interrupted", file=sys.stderr)
        return 130


if __name__ == "__main__":
    raise SystemExit(main())
