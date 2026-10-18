import argparse
import json
import os
from pathlib import Path
import random
import sys

import numpy as np
import soundfile as sf
import torch

work = Path(__file__).resolve().parent
root = work.parents[1]
parser = argparse.ArgumentParser(description="Run a fixed manifest of RVC voice-similarity experiments.")
parser.add_argument("manifest", type=Path)
args = parser.parse_args()
manifest_path = args.manifest.resolve()
trials = json.loads(manifest_path.read_text())
if not trials or len({trial["id"] for trial in trials}) != len(trials):
    parser.error("Provide a nonempty trial list with unique IDs.")

os.environ.update(TORCH_FORCE_WEIGHTS_ONLY_LOAD="1", OMP_NUM_THREADS="4", OPENBLAS_NUM_THREADS="1", HF_HUB_OFFLINE="1")
os.environ["weight_root"] = str(root / "models/brad-pitt")
os.environ["rmvpe_root"] = str(root / "vendor/rvc/assets/rmvpe")
os.environ["index_root"] = str(root / "vendor/rvc/logs")
os.environ["outside_index_root"] = str(root / "models/brad-pitt")
os.chdir(root / "vendor/rvc")
sys.path.insert(0, str(root / "vendor/rvc"))
torch.set_num_threads(4)
from infer.cli import create_config
from infer.vc.modules import VC

vc = VC(create_config())
vc.get_vc("BradPitt.pth")
output_dir = work / "candidates"
output_dir.mkdir(exist_ok=True)
for trial in trials:
    source = work / "sources" / f"{trial['prompt']}-{trial['voice']}.wav"
    if not source.is_file():
        raise FileNotFoundError(source)
    seed = trial.get("seed", 11)
    random.seed(seed)
    np.random.seed(seed)
    torch.manual_seed(seed)
    status, result = vc.vc_single(
        0, str(source), trial["pitch"], "rmvpe",
        str(root / "models/brad-pitt/added_IVF840_Flat_nprobe_1_BradPitt_v2.index"),
        trial["index_rate"], 0, 1.0, trial.get("protect", 0.33),
    )
    if result is None or result[0] is None or result[1] is None:
        raise RuntimeError(f"Trial {trial['id']} failed: {status}")
    target = output_dir / f"{trial['id']}.wav"
    sf.write(target, result[1], result[0])
    metadata = {**trial, "source": str(source), "output": str(target), "sample_rate": result[0], "samples": len(result[1]), "status": status}
    target.with_suffix(".json").write_text(json.dumps(metadata, ensure_ascii=False, indent=2) + "\n")
    print(json.dumps({"id": trial["id"], "output": str(target), "samples": len(result[1]), "seed": seed}), flush=True)
