"""Prepare immutable input snapshots without invoking a model."""

import hashlib
import json
from pathlib import Path

from .launcher import preview


TEMPLATE = Path(__file__).resolve().parent.parent / "docs/benchmarking/templates/run/run.json"


def prepare(experiment, output):
    data = experiment.data
    output = Path(output).resolve()
    resolved = preview(experiment, output)
    template = json.loads(TEMPLATE.read_text(encoding="utf-8"))
    record = {
        **template,
        "record_status": "prepared",
        "task_id": data["task"]["id"],
        "run_id": output.name,
        "experiment": {"path": "experiment.json", "sha256": hashlib.sha256(experiment.raw).hexdigest()},
        "cli": {**template["cli"], "name": data["cli"]["name"]},
        "requested": {"model": data["model"], "effort": data["effort"], **data["capabilities"]},
        "prompt": {
            **template["prompt"],
            "version": data["task"]["prompt_version"],
            "sha256": resolved["prompt_sha256"],
            "task_snapshot_path": "task-prompt.md",
            "task_sha256": resolved["task_prompt_sha256"],
            "treatment_snapshot_path": "treatment.md" if resolved["treatment"] else None,
            "treatment_sha256": resolved["treatment_sha256"],
        },
    }
    output.mkdir(parents=True, exist_ok=False)
    snapshots = {
        "experiment.json": experiment.raw,
        "task-prompt.md": experiment.prompt,
        "prompt.md": resolved["delivered_prompt"].encode("utf-8"),
        "response.txt": b"",
        "errors.log": b"",
        "preview.json": (json.dumps(resolved, indent=2, ensure_ascii=False) + "\n").encode("utf-8"),
        "run.json": (json.dumps(record, indent=2, ensure_ascii=False) + "\n").encode("utf-8"),
        **({"treatment.md": resolved["treatment"].encode("utf-8")} if resolved["treatment"] else {}),
    }
    for name, content in snapshots.items():
        with (output / name).open("xb") as stream:
            stream.write(content)
    return record
