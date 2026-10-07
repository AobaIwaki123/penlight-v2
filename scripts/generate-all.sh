#!/usr/bin/env bash
set -euo pipefail

# scripts/generate-all.sh
# Go 構造体 (pkg/model/) を正本としたスキーマ・型定義・ER図の一括自動生成 (Ref: ADR-0004)

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

go run scripts/gen-proposal-migration.go

echo "=== [1/4] Rebuilding Master Seed SQL from seeds/data/ ==="
go run scripts/build_seed.go

echo "=== [2/4] Generating ER Diagram from Go AST ==="
go run scripts/gen-er-diagram.go

echo "=== [3/4] Generating TypeScript Types via tygo ==="
TYGO_BIN="$(which tygo 2>/dev/null || true)"
if [ -z "$TYGO_BIN" ]; then
    GOPATH_TYGO="$(go env GOPATH)/bin/tygo"
    if [ -x "$GOPATH_TYGO" ]; then
        TYGO_BIN="$GOPATH_TYGO"
    fi
fi

if [ -n "$TYGO_BIN" ]; then
    "$TYGO_BIN" generate
else
    echo "tygo binary not found in PATH or GOPATH, running via 'go run github.com/gzuidhof/tygo@latest'..."
    go run github.com/gzuidhof/tygo@latest generate
fi

echo "=== [4/4] Running Go Static Analysis and Tests ==="
go vet ./...
go test ./...

echo "=== All schema artifacts successfully generated and verified! ==="
