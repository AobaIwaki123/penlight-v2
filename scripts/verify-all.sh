#!/usr/bin/env bash
set -euo pipefail

# scripts/verify-all.sh
# 1コマンドでスキーマ再生成、静的解析、単体テスト、Biome、typos、actionlint を一括検証 (Ref: ADR-0013)

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

STAGE_MODE=false
AI_MODE=false

for arg in "$@"; do
    if [ "$arg" == "--stage" ] || [ "$arg" == "-s" ]; then
        STAGE_MODE=true
    elif [ "$arg" == "--ai" ] || [ "$arg" == "--quiet" ] || [ "$arg" == "-q" ]; then
        AI_MODE=true
    fi
done

run_step() {
    local step_num="$1"
    local total_steps="8"
    local title="$2"
    shift 2

    if [ "$AI_MODE" = true ]; then
        local out
        local exit_code=0
        out=$("$@" 2>&1) || exit_code=$?
        if [ "$exit_code" -ne 0 ]; then
            echo "❌ FAILED: [$step_num/$total_steps] $title (exit code: $exit_code)" >&2
            echo "$out" >&2
            exit "$exit_code"
        fi
    else
        echo "=== [$step_num/$total_steps] $title ==="
        "$@"
    fi
}

run_step 1 "Running Schema Generation & Tests" ./scripts/generate-all.sh

if [ "$STAGE_MODE" = true ]; then
    if [ "$AI_MODE" = false ]; then
        echo "=== Auto-staging generated artifacts (--stage) ==="
    fi
    git add frontend/src/types/generated.ts assets/schema/ seeds/seed.sql seeds/data/image_sources.json data/image_sources.json 2>/dev/null || true
fi

run_step 2 "Verifying Schema Sync (git diff)" git diff --exit-code frontend/src/types/generated.ts assets/schema/
run_step 3 "Verifying Master Data Integrity" go run scripts/verify_master.go
run_step 4 "Verifying Dockerfile & Base Image Integrity" go run scripts/verify_dockerfile.go
run_step 5 "Running Biome Lint & Format Check" biome check

step6_frontend() {
    npm run --prefix frontend typecheck
    npm run --prefix frontend knip
}
run_step 6 "Running Frontend Typecheck & Knip Audit" step6_frontend

run_step 7 "Running typos Spell Check" typos

step8_actionlint() {
    if which actionlint >/dev/null 2>&1; then
        actionlint
    elif [ "$AI_MODE" = false ]; then
        echo "actionlint not found, skipping."
    fi
}
run_step 8 "Running actionlint (GitHub Actions Static Check)" step8_actionlint

echo "=== ✅ All checks passed successfully! ==="
