#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT="${1:-"$ROOT/test_output/render_core/gtk3_gpu_snapshot.png"}"
TIMEOUT_VALUE="${GPUI_SNAPSHOT_TIMEOUT:-12s}"
GOCACHE_VALUE="${GOCACHE:-/tmp/gpui-go-cache}"

mkdir -p "$(dirname "$OUT")"
cd "$ROOT"

echo "Generating GTK3 GPU snapshot: $OUT"
env GOCACHE="$GOCACHE_VALUE" GPUI_WS=gtk3 GPUI_GPU_SNAPSHOT="$OUT" timeout "$TIMEOUT_VALUE" go run ./demo

echo "Validating GTK3 GPU snapshot"
env GOCACHE="$GOCACHE_VALUE" go run ./cmd/validate_snapshot \
  -file "$OUT" \
  -width 800 \
  -height 600 \
  -min-non-bg 1000 \
  -min-colors 8
