#!/usr/bin/env bash
set -euo pipefail

# scripts/verify-all.sh
# 1コマンドでスキーマ再生成、静的解析、単体テスト、Biome、typos、actionlint を一括検証 (Ref: ADR-0013)

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

STAGE_MODE=false
for arg in "$@"; do
    if [ "$arg" == "--stage" ] || [ "$arg" == "-s" ]; then
        STAGE_MODE=true
    fi
done

echo "=== [1/8] Running Schema Generation & Tests ==="
./scripts/generate-all.sh

if [ "$STAGE_MODE" = true ]; then
    echo "=== Auto-staging generated artifacts (--stage) ==="
    git add frontend/src/types/generated.ts assets/schema/ seeds/seed.sql seeds/data/image_sources.json data/image_sources.json 2>/dev/null || true
fi

echo "=== [2/8] Verifying Schema Sync (git diff) ==="
git diff --exit-code frontend/src/types/generated.ts assets/schema/

echo "=== [3/8] Verifying Master Data Integrity ==="
go run scripts/verify_master.go

echo "=== [4/8] Verifying Dockerfile & Base Image Integrity ==="
go run scripts/verify_dockerfile.go

echo "=== [5/8] Running Biome Lint & Format Check ==="
biome check

echo "=== [6/8] Running Frontend Typecheck & Knip Audit ==="
npm run --prefix frontend typecheck
npm run --prefix frontend knip

echo "=== [7/8] Running typos Spell Check ==="
typos

echo "=== [8/8] Running actionlint (GitHub Actions Static Check) ==="
if which actionlint >/dev/null 2>&1; then
    actionlint
else
    echo "actionlint not found, skipping."
fi

echo "=== ✅ All checks passed successfully! ==="
