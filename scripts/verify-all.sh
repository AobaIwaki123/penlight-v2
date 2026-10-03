#!/usr/bin/env bash
set -euo pipefail

# scripts/verify-all.sh
# 1コマンドでスキーマ再生成、静的解析、単体テスト、Biome、typos、actionlint を一括検証 (Ref: ADR-0013)

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

echo "=== [1/5] Running Schema Generation & Tests ==="
./scripts/generate-all.sh

echo "=== [2/5] Verifying Schema Sync (git diff) ==="
git diff --exit-code frontend/src/types/generated.ts assets/schema/

echo "=== [3/5] Running Biome Lint & Format Check ==="
biome check

echo "=== [4/5] Running typos Spell Check ==="
typos

echo "=== [5/5] Running actionlint (GitHub Actions Static Check) ==="
if which actionlint >/dev/null 2>&1; then
    actionlint
else
    echo "actionlint not found, skipping."
fi

echo "=== ✅ All checks passed successfully! ==="
