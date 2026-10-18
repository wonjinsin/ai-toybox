#!/usr/bin/env python3.13
"""Install pinned local inference inputs without changing existing identities."""

from __future__ import annotations

import argparse
import json
import os
from pathlib import Path, PurePosixPath
import re
import shutil
import stat
import subprocess
import sys
import tarfile
import tempfile
import urllib.request
import venv
import zipfile

sys.dont_write_bytecode = True
os.environ["PYTHONDONTWRITEBYTECODE"] = "1"

from runtime import (
    PROJECT_ROOT, check_dependencies, check_prerequisites, check_runtime,
    load_manifest, offline_environment, safe_path, sha256, verify_file,
    verify_source,
)


def _archive_name(name: str) -> None:
    path = PurePosixPath(name)
    if path.is_absolute() or ".." in path.parts or "\\" in name or not path.parts:
        raise RuntimeError(f"Unsafe archive member: {name!r}")


def _promote(staged: Path, destination: Path, expected: str) -> None:
    verify_file(staged, expected)
    if destination.exists() or destination.is_symlink():
        verify_file(destination, expected)
        return
    destination.parent.mkdir(parents=True, exist_ok=True)
    try:
        os.link(staged, destination)
    except FileExistsError:
        verify_file(destination, expected)


def download_artifact(root: Path, artifact: dict, staging: Path) -> Path:
    destination = safe_path(root, artifact["path"])
    if destination.exists() or destination.is_symlink():
        verify_file(destination, artifact["sha256"])
        print(f"Reuse {artifact['path']}", flush=True)
        return destination
    staged = staging / "download"
    print(f"Download {artifact['url']}", flush=True)
    with urllib.request.urlopen(artifact["url"], timeout=120) as response, staged.open("wb") as output:
        shutil.copyfileobj(response, output)
    _promote(staged, destination, artifact["sha256"])
    staged.unlink()
    return destination


def extract_artifact(root: Path, artifact: dict, archive: Path, staging: Path) -> None:
    expected = {member["member"]: member for member in artifact.get("members", [])}
    if not expected:
        return
    archive_format = artifact.get("format")
    if archive_format == "zip":
        with zipfile.ZipFile(archive) as opened:
            names = set()
            for info in opened.infolist():
                _archive_name(info.filename)
                if info.filename in names:
                    raise RuntimeError(f"Duplicate archive member: {info.filename}")
                names.add(info.filename)
                mode = info.external_attr >> 16
                if stat.S_ISLNK(mode):
                    raise RuntimeError(f"Unsafe symbolic link archive member: {info.filename}")
                if not info.is_dir() and info.filename not in expected:
                    raise RuntimeError(f"Unexpected archive member: {info.filename}")
            for name, member in expected.items():
                if name not in names or opened.getinfo(name).is_dir():
                    raise RuntimeError(f"missing archive member: {name}")
                staged = staging / "member"
                with opened.open(name) as source, staged.open("wb") as output:
                    shutil.copyfileobj(source, output)
                _promote(staged, safe_path(root, member["path"]), member["sha256"])
                staged.unlink()
    elif archive_format == "tar.gz":
        with tarfile.open(archive, "r:gz") as opened:
            names = set()
            for info in opened.getmembers():
                _archive_name(info.name)
                if info.name in names:
                    raise RuntimeError(f"Duplicate archive member: {info.name}")
                names.add(info.name)
                if not info.isfile() and not info.isdir():
                    raise RuntimeError(f"Unsafe non-regular archive member: {info.name}")
                if info.isfile() and info.name not in expected:
                    raise RuntimeError(f"Unexpected archive member: {info.name}")
            for name, member in expected.items():
                if name not in names or not opened.getmember(name).isfile():
                    raise RuntimeError(f"missing archive member: {name}")
                staged = staging / "member"
                with opened.extractfile(name) as source, staged.open("wb") as output:
                    shutil.copyfileobj(source, output)
                _promote(staged, safe_path(root, member["path"]), member["sha256"])
                staged.unlink()
    else:
        raise RuntimeError(f"Unsupported archive format: {archive_format}")


def _installation_metadata(root: Path) -> dict | None:
    record = safe_path(root, ".runtime/installation.json")
    if not record.exists():
        return None
    metadata = json.loads(record.read_text())
    required = {"schema_version", "host", "dependency_locks", "artifact_manifests"}
    if not isinstance(metadata, dict) or set(metadata) != required or metadata.get("schema_version") != 1:
        raise RuntimeError(f"Unknown installation metadata: {record}; expected installer schema_version 1")
    if any(not isinstance(metadata[key], dict) for key in required - {"schema_version"}):
        raise RuntimeError(f"Invalid installation metadata fields: {record}")
    for key, filename in (("dependency_locks", "requirements-{}.lock"), ("artifact_manifests", "artifacts-{}.json")):
        for backend, observed in metadata[key].items():
            if backend not in {"rvc", "kokoro"}:
                raise RuntimeError(f"Unknown installation backend in {record}: {backend}")
            expected = sha256(PROJECT_ROOT / filename.format(backend))
            if observed != expected:
                raise RuntimeError(f"Installation provenance mismatch: {record}: {backend} {key}; expected SHA-256 {expected}; observed {observed}")
    return metadata


def _preflight_existing(root: Path, backends: list[str]) -> None:
    _installation_metadata(root)
    for backend in backends:
        envdir = root / (".venv" if backend == "rvc" else ".venv-kokoro")
        if envdir.exists() or envdir.is_symlink():
            if envdir.is_symlink():
                raise RuntimeError(f"Unknown symbolic link environment: {envdir}")
            check_dependencies(root, backend)
        for artifact in load_manifest(backend)["artifacts"]:
            for item in [artifact, *artifact.get("members", [])]:
                path = safe_path(root, item["path"])
                if path.exists() or path.is_symlink():
                    verify_file(path, item["sha256"])
            if artifact.get("role") == "rvc-source" and (root / ".runtime/rvc").exists():
                verify_source(root, artifact)


def _install_environment(root: Path, backend: str, staging: Path) -> None:
    envdir = root / (".venv" if backend == "rvc" else ".venv-kokoro")
    if envdir.exists():
        print(f"Reuse {envdir}", flush=True)
        return
    print(f"Install {backend} environment: {envdir}", flush=True)
    try:
        venv.EnvBuilder(with_pip=True).create(envdir)
        python = envdir / "bin/python"
        lock = PROJECT_ROOT / f"requirements-{backend}.lock"
        text = lock.read_text()
        records = re.findall(r"^[A-Za-z0-9_.-]+==[^\n]*(?:\n[ \t]+--hash=[^\n]*)*", text, re.MULTILINE)
        build_names = {"setuptools", "wheel", "cython", "packaging"}
        build_records = [record for record in records if record.split("==", 1)[0].lower() in build_names]
        env = offline_environment()
        env.pop("PYTHONPATH", None)
        base = [str(python), "-B", "-m", "pip", "--disable-pip-version-check", "install", "--no-cache-dir", "--require-hashes"]
        if build_records:
            build_lock = staging / f"build-{backend}.lock"
            build_lock.write_text("\n".join(build_records) + "\n")
            subprocess.run([*base, "--no-deps", "-r", str(build_lock)], env=env, check=True)
        # Release large dictionary build trees before collecting/installing the full graph.
        for record in records:
            name = record.split("==", 1)[0].lower()
            if name not in {"mojimoji", "unidic-lite"}:
                continue
            source_lock = staging / f"source-{backend}-{name}.lock"
            source_lock.write_text(record + "\n")
            subprocess.run([*base, "--no-deps", "--no-build-isolation", "-r", str(source_lock)], env=env, check=True)
        subprocess.run([*base, "--no-build-isolation", "-r", str(lock)], env=env, check=True)
        check_dependencies(root, backend)
    except BaseException:
        if envdir.exists():
            shutil.rmtree(envdir)
        raise


def install(root: Path, *, rvc_only: bool = False) -> dict:
    host = check_prerequisites()
    backends = ["rvc"] if rvc_only else ["rvc", "kokoro"]
    if not rvc_only:
        if not shutil.which("xcrun"):
            raise RuntimeError("missing Apple Command Line Tools: xcrun is required to build locked Japanese processing dependencies")
        compiler = subprocess.run(["xcrun", "--find", "clang++"], capture_output=True, text=True)
        if compiler.returncode:
            raise RuntimeError(f"missing Apple C++ compiler: {compiler.stderr.strip()}")
    _preflight_existing(root, backends)
    root.mkdir(parents=True, exist_ok=True)
    runtime_dir = safe_path(root, ".runtime")
    runtime_dir.mkdir(parents=True, exist_ok=True)
    with tempfile.TemporaryDirectory(prefix=".setup-", dir=runtime_dir) as temporary:
        staging = Path(temporary)
        for backend in backends:
            _install_environment(root, backend, staging)
        for backend in backends:
            for artifact in load_manifest(backend)["artifacts"]:
                archive = download_artifact(root, artifact, staging)
                extract_artifact(root, artifact, archive, staging)
    result = check_runtime(root, "say" if rvc_only else "kokoro")
    record = safe_path(root, ".runtime/installation.json")
    metadata = {
        "schema_version": 1,
        "host": host,
        "dependency_locks": {backend: sha256(PROJECT_ROOT / f"requirements-{backend}.lock") for backend in backends},
        "artifact_manifests": {backend: sha256(PROJECT_ROOT / f"artifacts-{backend}.json") for backend in backends},
    }
    previous = _installation_metadata(root)
    if previous is not None:
        for key in ("dependency_locks", "artifact_manifests"):
            previous.setdefault(key, {}).update(metadata[key])
        metadata = previous
    encoded = json.dumps(metadata, indent=2) + "\n"
    if not record.exists() or record.read_text() != encoded:
        record.write_text(encoded)
    return result


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--runtime-root", type=Path, default=PROJECT_ROOT)
    parser.add_argument("--check", action="store_true", help="Read-only offline verification; no installation or downloads")
    parser.add_argument("--rvc-only", action="store_true", help="Install/check the converter and system-speech baseline only")
    args = parser.parse_args()
    root = args.runtime_root.expanduser().resolve()
    try:
        result = check_runtime(root, "say" if args.rvc_only else "kokoro") if args.check else install(root, rvc_only=args.rvc_only)
    except (RuntimeError, OSError, ValueError, subprocess.CalledProcessError, tarfile.TarError, zipfile.BadZipFile) as error:
        print(f"Runtime setup failed: {error}", file=sys.stderr)
        return 1
    print(json.dumps(result, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
