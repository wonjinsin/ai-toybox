"""Read-only integrity and offline environment boundary for local inference."""

from __future__ import annotations

import hashlib
import json
import os
from pathlib import Path, PurePosixPath
import platform
import re
import shutil
import subprocess
import sys
import tempfile

PROJECT_ROOT = Path(__file__).resolve().parent
MODEL_NAME = "BradPitt.pth"
INDEX_NAME = "added_IVF840_Flat_nprobe_1_BradPitt_v2.index"
REED_VOICE = "Reed (일본어(일본))"


def sha256(path: Path) -> str:
    with path.open("rb") as stream:
        return hashlib.file_digest(stream, "sha256").hexdigest()


def safe_path(root: Path, relative: str) -> Path:
    part = PurePosixPath(relative)
    if part.is_absolute() or ".." in part.parts or not part.parts or "\\" in relative:
        raise RuntimeError(f"Unsafe artifact path: {relative!r}")
    path = root.joinpath(*part.parts)
    for component in (path, *path.parents):
        if component == root:
            break
        if component.is_symlink():
            raise RuntimeError(f"Unsafe symbolic link in artifact path: {component}")
    return path


def verify_file(path: Path, expected: str) -> None:
    if not path.is_file() or path.is_symlink():
        raise RuntimeError(f"missing regular runtime artifact: {path}; expected SHA-256 {expected}")
    observed = sha256(path)
    if observed != expected:
        raise RuntimeError(f"Artifact mismatch: {path}; expected SHA-256 {expected}; observed {observed}")


def load_manifest(backend: str) -> dict:
    path = PROJECT_ROOT / f"artifacts-{backend}.json"
    if not path.is_file():
        raise RuntimeError(f"missing artifact manifest: {path}")
    result = json.loads(path.read_text())
    if result.get("schema_version") != 1 or not isinstance(result.get("artifacts"), list):
        raise RuntimeError(f"Invalid artifact manifest: {path}")
    return result


def locked_requirements(backend: str) -> dict[str, str]:
    path = PROJECT_ROOT / f"requirements-{backend}.lock"
    if not path.is_file():
        raise RuntimeError(f"missing dependency lock: {path}")
    return {
        re.sub(r"[-_.]+", "-", name).lower(): version
        for name, version in re.findall(r"^([A-Za-z0-9_.-]+)(?:\[[^\]\r\n]+\])?==([^\s\\]+)", path.read_text(), re.MULTILINE)
    }


def _run(command: list[str], *, env: dict[str, str] | None = None, cwd: Path | None = None) -> str:
    result = subprocess.run(command, env=env, cwd=cwd, text=True, capture_output=True)
    if result.returncode:
        raise RuntimeError(f"Runtime command failed ({result.returncode}): {' '.join(command)}\n{result.stderr.strip()}\n{result.stdout.strip()}")
    return result.stdout.strip()


def check_prerequisites() -> dict:
    if platform.system() != "Darwin" or platform.machine() != "arm64":
        raise RuntimeError(f"Unsupported host: expected macOS arm64; observed {platform.system()} {platform.machine()}")
    if sys.version_info[:2] != (3, 13):
        raise RuntimeError(f"Python prerequisite: expected 3.13; observed {platform.python_version()}")
    for binary in ("ffmpeg", "say", "afplay"):
        if not shutil.which(binary):
            raise RuntimeError(f"missing system prerequisite: {binary}")
    voices = _run(["say", "-v", "?"])
    if not re.search(r"^Reed(?:\s|\().*\bja_JP\b", voices, re.MULTILINE):
        raise RuntimeError("missing installed Japanese Reed speech voice; inspect say -v '?'")
    return {
        "python": platform.python_version(),
        "macos": platform.mac_ver()[0],
        "architecture": platform.machine(),
        "ffmpeg": _run(["ffmpeg", "-version"]).splitlines()[0],
        "voices": voices.splitlines(),
    }


def offline_environment() -> dict[str, str]:
    env = os.environ.copy()
    env.pop("KMP_DUPLICATE_LIB_OK", None)
    env.pop("PYTHONHOME", None)
    env.update({
        "PYTHONDONTWRITEBYTECODE": "1",
        "PYTHONPYCACHEPREFIX": "/dev/null",
        "PYTHONNOUSERSITE": "1",
        "HF_HUB_OFFLINE": "1",
        "TRANSFORMERS_OFFLINE": "1",
        "HF_HUB_DISABLE_TELEMETRY": "1",
        "DO_NOT_TRACK": "1",
        "TOKENIZERS_PARALLELISM": "false",
    })
    return env


def rvc_environment(root: Path) -> dict[str, str]:
    root = root.expanduser().resolve()
    python = root / ".venv/bin/python"
    if not python.is_file():
        raise RuntimeError(f"missing RVC Python environment: {python}; run bootstrap.py --runtime-root {root}")
    env = offline_environment()
    env.pop("PYTHONPATH", None)
    purelib = Path(_run([str(python), "-B", "-c", "import sysconfig; print(sysconfig.get_path('purelib'))"], env=env))
    torch_lib = purelib / "torch/lib"
    if not torch_lib.is_dir():
        raise RuntimeError(f"missing PyTorch shared library directory: {torch_lib}")
    env.update({
        "DYLD_LIBRARY_PATH": str(torch_lib),
        "PYTHONPATH": str(root / ".runtime/rvc"),
        "TORCH_FORCE_WEIGHTS_ONLY_LOAD": "1",
        "OMP_NUM_THREADS": "4",
        "OPENBLAS_NUM_THREADS": "1",
        "MKL_NUM_THREADS": "4",
        "RVC_AUDIO_FORCE_CPU": "1",
        "weight_root": str(root / "models"),
        "index_root": str(root / "models"),
        "outside_index_root": str(root / "models"),
        "rmvpe_root": str(root / ".runtime/rvc/assets/rmvpe"),
    })
    return env


def check_dependencies(root: Path, backend: str) -> dict[str, str]:
    python = root / (".venv" if backend == "rvc" else ".venv-kokoro") / "bin/python"
    if not python.is_file():
        raise RuntimeError(f"missing {backend} Python environment: {python}; run bootstrap.py --runtime-root {root}")
    code = "import importlib.metadata as m,json,platform,re; print(json.dumps({'python':platform.python_version(),'packages':{re.sub(r'[-_.]+','-',d.metadata['Name']).lower():d.version for d in m.distributions()}}))"
    env = offline_environment()
    env.pop("PYTHONPATH", None)
    observed = json.loads(_run([str(python), "-B", "-c", code], env=env))
    if not observed["python"].startswith("3.13."):
        raise RuntimeError(f"Environment mismatch: {python}; expected Python 3.13; observed {observed['python']}")
    expected = locked_requirements(backend)
    for name, version in expected.items():
        installed = observed["packages"].get(name, "missing")
        if installed != version:
            raise RuntimeError(f"Dependency mismatch: {python}: {name}; expected {version}; observed {installed}; existing environments are never upgraded")
    extras = set(observed["packages"]) - set(expected) - {"pip"}
    if extras:
        raise RuntimeError(f"Unknown installed dependencies in {python}: {', '.join(sorted(extras))}")
    _run([str(python), "-B", "-m", "pip", "check"], env=env)
    return expected


def verify_source(root: Path, artifact: dict) -> None:
    members = artifact["members"]
    for member in members:
        verify_file(safe_path(root, member["path"]), member["sha256"])
    expected = {member["path"] for member in members}
    expected.update(
        item["path"] for item in load_manifest("rvc")["artifacts"]
        if not item.get("members")
    )
    source_root = safe_path(root, ".runtime/rvc")
    for candidate in source_root.rglob("*"):
        relative = candidate.relative_to(source_root)
        if ".git" in relative.parts or "__pycache__" in relative.parts:
            continue
        if not candidate.is_file() and not candidate.is_symlink():
            continue
        if candidate.relative_to(root).as_posix() in expected:
            continue
        generated = relative.parts[0] in {"logs", "TEMP"}
        if generated and candidate.suffix.lower() not in {".py", ".pyi", ".so", ".dylib"} and not candidate.is_symlink():
            continue
        raise RuntimeError(f"Unknown RVC source file: {candidate}; expected absent from pinned source")


def verify_artifacts(root: Path, backend: str, *, require_archives: bool = False) -> dict[str, str]:
    manifest = load_manifest(backend)
    identities = {}
    for artifact in manifest["artifacts"]:
        path = safe_path(root, artifact["path"])
        members = artifact.get("members", [])
        if not members or path.exists() or require_archives:
            verify_file(path, artifact["sha256"])
        if artifact.get("role") == "rvc-source":
            verify_source(root, artifact)
        else:
            for member in members:
                verify_file(safe_path(root, member["path"]), member["sha256"])
        if not members:
            identities[artifact["path"]] = artifact["sha256"]
        elif artifact.get("role") != "rvc-source":
            identities.update({member["path"]: member["sha256"] for member in members})
    return identities


_RVC_PREFLIGHT = r'''
import json, sys
from pathlib import Path
import torch, faiss, librosa, parselmouth, av, soundfile
from infer.cli import create_config
from infer.vc.modules import VC
from infer.hubert import load_hubert_model, hubert_audio_requires_normalization
from infer.rmvpe import RMVPE
root = Path(sys.argv[1])
config = create_config()
assert config.device == 'cpu' and not config.is_half, 'Expected CPU float32'
checkpoint = torch.load(root / 'models/BradPitt.pth', map_location='cpu', weights_only=True)
assert checkpoint['weight']['emb_g.weight'].shape[0] > 0
vc = VC(config)
vc.get_vc('BradPitt.pth')
hubert = load_hubert_model('cpu', False)
hubert_audio_requires_normalization()
rmvpe = RMVPE(str(root / '.runtime/rvc/assets/rmvpe/rmvpe.pt'), False, device='cpu')
index = faiss.read_index(str(root / 'models/added_IVF840_Flat_nprobe_1_BradPitt_v2.index'))
assert index.ntotal > 0 and index.d == 768
print(json.dumps({'device':'cpu','precision':'float32','index_vectors':index.ntotal,'index_dimensions':index.d,'sample_rate':vc.tgt_sr,'weights_only':True}))
'''


def check_runtime(root: Path, backend: str = "say") -> dict:
    if backend not in {"say", "kokoro"}:
        raise RuntimeError(f"Unknown source backend: {backend}")
    root = root.expanduser().resolve()
    if not (root / ".venv/bin/python").is_file():
        raise RuntimeError(f"missing RVC runtime: {root / '.venv'}; run bootstrap.py --runtime-root {root}")
    host = check_prerequisites()
    packages = {"rvc": check_dependencies(root, "rvc")}
    identities = verify_artifacts(root, "rvc")
    env = rvc_environment(root)
    # Librosa's eager Numba JIT ignores -B; keep its transient cache outside the target.
    with tempfile.TemporaryDirectory(prefix="voice-runtime-check-") as temporary:
        env["NUMBA_CACHE_DIR"] = temporary
        loaded = _run([str(root / ".venv/bin/python"), "-B", "-c", _RVC_PREFLIGHT, str(root)], env=env, cwd=root / ".runtime/rvc")
    model_load = json.loads(loaded.splitlines()[-1])
    sources = {
        name: value for name, value in load_manifest("rvc").get("sources", {}).items()
        if name != "python_dependencies"
    }
    if backend == "kokoro":
        packages["kokoro"] = check_dependencies(root, "kokoro")
        identities.update(verify_artifacts(root, "kokoro"))
        sources["kokoro"] = load_manifest("kokoro").get("sources", {}).get("model", {})
        kokoro_env = offline_environment()
        kokoro_env.pop("PYTHONPATH", None)
        _run([str(root / ".venv-kokoro/bin/python"), "-B", str(PROJECT_ROOT / "kokoro_source.py"), "--runtime-root", str(root), "--check"], env=kokoro_env)
    return {"schema_version": 1, "runtime_root": str(root), "backend": backend, "host": host, "packages": packages, "sources": sources, "artifacts": identities, "model_load": model_load}
