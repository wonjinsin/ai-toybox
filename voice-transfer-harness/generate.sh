#!/bin/bash
set -euo pipefail

if [[ $# -ne 1 || -z "$1" ]]; then
    printf 'Usage: bash generate.sh "Japanese text"\n' >&2
    exit 2
fi

ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
mkdir -p "$ROOT/outputs"
OUTPUT_DIR="$(mktemp -d "$ROOT/outputs/run.XXXXXX")"

say -v 'Reed (일본어(일본))' -r 165 -o "$OUTPUT_DIR/source.aiff" "$1"
ffmpeg -nostdin -hide_banner -loglevel error -n \
    -i "$OUTPUT_DIR/source.aiff" -ar 16000 -ac 1 "$OUTPUT_DIR/source.wav"

# PyTorch and FAISS must share one OpenMP runtime on this Mac.
env DYLD_LIBRARY_PATH="$ROOT/.venv/lib/python3.13/site-packages/torch/lib" \
    PYTHONPATH="$ROOT/.runtime/rvc" \
    TORCH_FORCE_WEIGHTS_ONLY_LOAD=1 \
    OMP_NUM_THREADS=4 OPENBLAS_NUM_THREADS=1 MKL_NUM_THREADS=4 \
    RVC_AUDIO_FORCE_CPU=1 \
    "$ROOT/.venv/bin/python" -m infer.cli \
    --model "$ROOT/models/BradPitt.pth" \
    --input "$OUTPUT_DIR/source.wav" \
    --output "$OUTPUT_DIR/brad-pitt-ja.wav" \
    --index "$ROOT/models/added_IVF840_Flat_nprobe_1_BradPitt_v2.index" \
    --f0-method rmvpe --pitch 0 --index-rate 0.75

printf '\nSource: %s\nConverted: %s\n' \
    "$OUTPUT_DIR/source.wav" "$OUTPUT_DIR/brad-pitt-ja.wav"
