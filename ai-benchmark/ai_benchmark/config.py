"""Validate experiment inputs without starting a model."""

import hashlib
import json
import re
from dataclasses import dataclass
from pathlib import Path


class ConfigurationError(ValueError):
    """The experiment cannot be prepared as requested."""


@dataclass(frozen=True)
class Experiment:
    path: Path
    raw: bytes
    prompt_path: Path
    prompt: bytes
    model: str

    @property
    def data(self):
        return json.loads(self.raw)


def _object(value, keys, label):
    if not isinstance(value, dict):
        raise ConfigurationError(f"{label} must be an object")
    unknown = set(value) - set(keys)
    missing = set(keys) - set(value)
    if unknown or missing:
        raise ConfigurationError(
            f"{label}: unknown keys {sorted(unknown)}; missing keys {sorted(missing)}"
        )


def _text(value, label, pattern=None):
    if not isinstance(value, str) or not value.strip() or value != value.strip():
        raise ConfigurationError(f"{label} must be explicitly specified as nonempty text")
    if any(ord(character) < 32 or ord(character) == 127 for character in value):
        raise ConfigurationError(f"{label} must not contain control characters")
    try:
        value.encode("utf-8")
    except UnicodeError as error:
        raise ConfigurationError(f"{label} must be valid UTF-8 text") from error
    if pattern and re.fullmatch(pattern, value) is None:
        raise ConfigurationError(f"{label} has an invalid format")


def _validate_settings(data):
    _object(data, ("format_version", "status", "experiment_id", "task", "cli",
                   "model", "effort", "capabilities", "artifact", "isolation_profile"),
            "experiment")
    if data["format_version"] != "0.1":
        raise ConfigurationError("experiment.format_version must be '0.1'")
    for key in ("experiment_id", "isolation_profile"):
        _text(data[key], key, r"[A-Za-z0-9][A-Za-z0-9_.-]*")
    _text(data["model"], "model", r"[A-Za-z0-9][A-Za-z0-9_./:-]*")
    _text(data["effort"], "effort", r"[a-z][a-z0-9_-]*")
    _object(data["cli"], ("name", "required_version", "mode"), "cli")
    if data["cli"]["name"] != "codex":
        raise ConfigurationError("cli.name: only the codex adapter is defined")
    if data["cli"]["mode"] != "non_interactive":
        raise ConfigurationError("cli.mode must be 'non_interactive'")
    _text(data["cli"]["required_version"], "cli.required_version", r"\d+\.\d+\.\d+(?:[-+][A-Za-z0-9.-]+)?")
    _object(data["task"], ("id", "prompt_path", "prompt_version", "prompt_sha256"), "task")
    for key in ("id", "prompt_path", "prompt_version"):
        _text(data["task"][key], f"task.{key}")
    _text(data["task"]["prompt_sha256"], "task.prompt_sha256", r"[0-9a-f]{64}")
    _object(data["capabilities"], ("skills", "tools", "mcp_servers", "agents"), "capabilities")
    for key, entries in data["capabilities"].items():
        if not isinstance(entries, list):
            raise ConfigurationError(f"capabilities.{key} must be an explicit array; [] permits none")
        if key != "skills":
            for entry in entries:
                _text(entry, f"capabilities.{key} entry")
    _object(data["artifact"], ("delivery",), "artifact")
    if data["artifact"]["delivery"] != "final_response":
        raise ConfigurationError("artifact.delivery: only final_response is implemented; file delivery is unsupported")


def _unique_object(pairs):
    keys = tuple(key for key, _ in pairs)
    if len(keys) != len(set(keys)):
        raise ConfigurationError("JSON contains duplicate keys")
    return dict(pairs)


def _read(path, label, limit):
    try:
        if not path.is_file():
            raise ConfigurationError(f"{label} must be an existing regular file")
        with path.open("rb") as stream:
            content = stream.read(limit + 1)
        if len(content) > limit:
            raise ConfigurationError(f"{label} exceeds the {limit}-byte input limit")
        return content
    except OSError as error:
        raise ConfigurationError(f"Cannot read {label}: {error.strerror}") from error


def skill_digest(directory):
    """Hash every regular file by relative path and content; reject links."""
    root = Path(directory)
    if root.is_symlink() or not root.is_dir() or not (root / "SKILL.md").is_file():
        raise ConfigurationError("skill path must be a real directory containing SKILL.md")
    digest = hashlib.sha256()
    for path in sorted(root.rglob("*")):
        if path.is_symlink():
            raise ConfigurationError("skill bundle must not contain symbolic links")
        if path.is_dir():
            continue
        if not path.is_file():
            raise ConfigurationError("skill bundle must contain only regular files")
        digest.update(path.relative_to(root).as_posix().encode("utf-8") + b"\x00")
        digest.update(hashlib.sha256(_read(path, "skill file", 10 * 1024 * 1024)).digest())
    return digest.hexdigest()


def _validate_skills(entries, base):
    names = tuple(entry.get("name") if isinstance(entry, dict) else None for entry in entries)
    if any(not isinstance(name, str) for name in names) or len(set(names)) != len(names):
        raise ConfigurationError("skills must have unique explicit names")
    for entry in entries:
        _object(entry, ("name", "path", "sha256", "invocation"), "skill")
        _text(entry["name"], "skill.name", r"[A-Za-z0-9][A-Za-z0-9_-]*")
        _text(entry["path"], "skill.path")
        _text(entry["sha256"], "skill.sha256", r"[0-9a-f]{64}")
        if entry["invocation"] != "explicit":
            raise ConfigurationError("skill.invocation must be 'explicit'")
        if skill_digest(base / entry["path"]) != entry["sha256"]:
            raise ConfigurationError("skill hash does not match the selected bundle")


def load_experiment(path):
    path = Path(path).resolve()
    raw = _read(path, "experiment", 1024 * 1024)
    try:
        data = json.loads(raw.decode("utf-8"), object_pairs_hook=_unique_object)
    except (UnicodeError, json.JSONDecodeError, RecursionError) as error:
        raise ConfigurationError("experiment must contain valid UTF-8 JSON within the parser nesting limit") from error
    if not isinstance(data, dict):
        raise ConfigurationError("experiment must be a JSON object")
    if data.get("status") != "ready":
        raise ConfigurationError("experiment.status must be 'ready'; draft inputs cannot run")
    _validate_settings(data)
    _validate_skills(data["capabilities"]["skills"], path.parent)
    prompt_path = (path.parent / data["task"]["prompt_path"]).resolve()
    prompt = _read(prompt_path, "prompt", 10 * 1024 * 1024)
    if hashlib.sha256(prompt).hexdigest() != data["task"]["prompt_sha256"]:
        raise ConfigurationError("task.prompt_sha256: prompt hash does not match the declared input")
    try:
        text = prompt.decode("utf-8")
    except UnicodeError as error:
        raise ConfigurationError("prompt must contain valid UTF-8 text") from error
    if not text.strip():
        raise ConfigurationError("task.prompt_path contains an empty prompt")
    if re.search(r"(?im)^Status:\s*Draft\b", text):
        raise ConfigurationError("task.prompt_path contains a draft; review and freeze the prompt first")
    return Experiment(path, raw, prompt_path, prompt, data["model"])
