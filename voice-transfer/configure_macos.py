import platform
from pathlib import Path
import subprocess
import sysconfig

if platform.system() != "Darwin":
    raise SystemExit("This environment configuration is for macOS only.")

packages = Path(sysconfig.get_path("purelib"))
shared_runtime = packages / "torch" / "lib" / "libomp.dylib"
if not shared_runtime.is_file():
    raise FileNotFoundError(shared_runtime)

for name in ("_swigfaiss.abi3.so", "libfaiss.dylib"):
    library = packages / "faiss" / name
    subprocess.run(
        ["install_name_tool", "-change", "@loader_path/.dylibs/libomp.dylib", "@loader_path/../torch/lib/libomp.dylib", str(library)],
        check=True,
    )
    subprocess.run(["codesign", "--force", "--sign", "-", str(library)], check=True)
    print(f"Linked {library.name} to PyTorch's OpenMP runtime.")
