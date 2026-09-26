"""Prepare explicit CLI invocations and gate execution on reviewed adapters."""

import hashlib
import json
from pathlib import Path


class UnsupportedCondition(ValueError):
    """No reviewed adapter can enforce the declared experiment condition."""


def run(experiment, output):
    raise UnsupportedCondition(preview(experiment, output)["blockers"][0])


def preview(experiment, output):
    data = experiment.data
    output = Path(output).resolve()
    treatment = "\n".join("$" + skill["name"] for skill in data["capabilities"]["skills"])
    delivered = treatment.encode("utf-8") + b"\n\n" + experiment.prompt if treatment else experiment.prompt
    return {
        "configuration_valid": True,
        "launch_ready": False,
        "blockers": [f"No reviewed isolation adapter is installed for {data['isolation_profile']!r}."],
        "cli_version_required": data["cli"]["required_version"],
        "model_support": "not_checked",
        "capability_controls": "unverified",
        "candidate_argv": [
            "codex", "exec", "--ephemeral", "--json", "--model", data["model"],
            "--config", "model_reasoning_effort=" + json.dumps(data["effort"]),
            "--output-last-message", str(output / "response.txt"), "-",
        ],
        "task_prompt_sha256": hashlib.sha256(experiment.prompt).hexdigest(),
        "treatment": treatment or None,
        "treatment_sha256": hashlib.sha256(treatment.encode("utf-8")).hexdigest() if treatment else None,
        "prompt_sha256": hashlib.sha256(delivered).hexdigest(),
        "delivered_prompt": delivered.decode("utf-8"),
        "output": str(output),
    }
