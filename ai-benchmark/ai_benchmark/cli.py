"""Local experiment preparation commands."""

import argparse
import hashlib
import json
import sys
from pathlib import Path

from .config import ConfigurationError, _read, load_experiment
from .evidence import summarize_events
from .launcher import UnsupportedCondition, preview, run
from .records import prepare


def main(argv=None):
    parser = argparse.ArgumentParser(description="Prepare benchmark inputs; no reviewed live adapter is installed.")
    commands = parser.add_subparsers(dest="command", required=True)
    for name in ("validate", "preview", "prepare", "run"):
        command = commands.add_parser(name)
        command.add_argument("experiment")
        if name != "validate":
            command.add_argument("--output", required=True)
    inspect = commands.add_parser("inspect", help="Summarize existing JSONL without modifying its source")
    inspect.add_argument("events")
    args = parser.parse_args(argv)
    try:
        result = _dispatch(args)
    except UnsupportedCondition as error:
        print(f"unsupported condition: {error}", file=sys.stderr)
        return 3
    except (ConfigurationError, OSError) as error:
        print(f"error: {error}", file=sys.stderr)
        return 2
    print(json.dumps(result, indent=2, ensure_ascii=True))
    return 0


def _dispatch(args):
    if args.command == "inspect":
        path = Path(args.events).resolve()
        raw = _read(path, "events", 64 * 1024 * 1024)
        return {
            "source": {"path": str(path), "sha256": hashlib.sha256(raw).hexdigest()},
            "summary": summarize_events(raw),
        }
    experiment = load_experiment(args.experiment)
    if args.command == "validate":
        return {"configuration_valid": True, "launch_ready": False, "model_support": "not_checked"}
    if args.command == "preview":
        return preview(experiment, args.output)
    if args.command == "run":
        return run(experiment, args.output)
    return prepare(experiment, args.output)
