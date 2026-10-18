#!/bin/bash
set -euo pipefail

ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
exec python3.13 -B "$ROOT/harness.py" "$@"
