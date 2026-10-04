---
name: container-guard
description: Dockerfileのベースイメージ最新LTS特定、マルチステージ分離、非root実行、CGOフリー静的バイナリ化、およびコンテナセキュリティを統制する。
---

# コンテナイメージ・Dockerfile 検証規約 (container-guard)

> **管轄 ADR**: [ADR-0003](../../../adr/0003-deployment-target-and-container-registry.md), [ADR-0012](../../../adr/0012-kubernetes-deployment-and-gitops-architecture.md)

本スキルは、`Dockerfile` を作成・更新する際に、AI の古い学習データによる「過去の Node/Go/Alpine バージョンの採用」を防ぎ、マルチステージビルドによる最小サイズ化とセキュリティ（非root実行・CGOフリー）を徹底するための運用手順を定める。

---

## 1. 絶対遵守ルール (Invariants)

1. **AI の学習データのみでベースイメージを決めることの禁止**:
   - AI は過去の LTS バージョン（例: Node 20, Go 1.22, Alpine 3.19 等）を最新と誤認しやすいため、必ずレジストリ API や GitHub リリースタグをリアルタイムに確認して最新 Active LTS / 安定版を採用すること。
2. **不確定タグ (`latest`) の使用禁止**:
   - ビルド環境の再現性を担保するため、`FROM node:latest` などの浮動タグは禁止。必ずメジャー・LTS バリアント（例: `node:22-alpine`, `golang:1.24-alpine`, `alpine:3.21`）を明示すること。
3. **最小権限の非root実行 (Non-root Execution)**:
   - 本番イメージは root 権限で起動してはならない。必ず専用の非rootユーザー（UID 10001: `penlight` 等）を作成し、`USER` ディレクティブで実行権限を降格させること。
4. **CGOフリー・完全静的バイナリ化 (`CGO_ENABLED=0`)**:
   - 純Go SQLite（`modernc.org/sqlite`）を採用しているため、`CGO_ENABLED=0` を指定して Cgo 依存・libc 依存を完全に排除し、`-trimpath -ldflags="-s -w"` でビルドすること。
5. **マルチステージビルドによる機密・不要ファイルの排除**:
   - Node.js SDK や npm キャッシュ、Go コンパイラ、Git、ソースコード全体を本番イメージに持ち込んではならない。実行ステージにはコンパイル済み単一バイナリ、CA 証明書、および `/data` ボリュームのみを含めること。
6. **`.dockerignore` の網羅性**:
   - `.git`, `node_modules`, `data/`, `*.sqlite*`, ドキュメント類など、ビルドコンテキストを不要に肥大化させるファイルは必ず `.dockerignore` で除外すること。

---

## 2. ベースイメージ最新 LTS / 安定版の特定コマンドリスト

新しいコンテナイメージを採用または更新する際は、以下のコマンドでリアルタイムに最新タグを確認する。

```bash
# 【Node.js】最新 Active LTS タグの確認 (Docker Hub API)
python3 -c "
import urllib.request, json
url = 'https://hub.docker.com/v2/repositories/library/node/tags?page_size=20&name=alpine'
req = urllib.request.Request(url, headers={'User-Agent': 'Mozilla/5.0'})
with urllib.request.urlopen(req) as resp:
    data = json.loads(resp.read().decode())
    tags = [r['name'] for r in data.get('results', []) if 'alpine' in r['name']]
    print('Node Alpine Tags:', tags[:10])
"

# 【Go】最新リリースバージョンの確認 (Go 公式 API)
curl -s "https://go.dev/dl/?mode=json" | python3 -c "
import sys, json
data = json.load(sys.stdin)
print('Latest Go Version:', data[0]['version'])
"

# 【Alpine】最新安定版タグの確認 (Docker Hub API)
python3 -c "
import urllib.request, json
url = 'https://hub.docker.com/v2/repositories/library/alpine/tags?page_size=10'
req = urllib.request.Request(url, headers={'User-Agent': 'Mozilla/5.0'})
with urllib.request.urlopen(req) as resp:
    data = json.loads(resp.read().decode())
    print('Alpine Tags:', [r['name'] for r in data.get('results', [])][:5])
"
```

---

## 3. 自動検証スクリプト (`scripts/verify_dockerfile.go`)

`./scripts/verify-all.sh` [4/8] および CI に組み込まれており、以下の不整合を機械的に事前検知して `exit 1` でブロックする：
1. `go.mod` の Go バージョン要件（`>= 1.26`）と `Dockerfile` の `golang:X.Y-alpine` の不整合
2. Node.js ベースイメージが Active LTS（`>= 22`）を満たしているか
3. `CGO_ENABLED=0`（純Go静的コンパイル）の指定漏れ
4. 非root実行用の `USER` ディレクティブの指定漏れ

---

## 4. ローカル検証チェックリスト (Dockerfile 作成・変更時)

- [ ] `go run scripts/verify_dockerfile.go` がグリーンで通過したか
- [ ] Node.js は最新 Active LTS（`node:22-alpine` 等）であることを確認したか
- [ ] Go は `go.mod` と一致する安定版（`golang:1.26-alpine` 等）であることを確認したか
- [ ] Alpine は最新安定版（`alpine:3.24` 等）であることを確認したか
- [ ] Go コンパイル時に `CGO_ENABLED=0 GOOS=linux` および `-trimpath -ldflags="-s -w"` が付与されているか
- [ ] 本番ランタイムコンテナで `USER`（非root）が指定され、ボリューム領域（`/data`）の所有権が適切に設定されているか
- [ ] `.dockerignore` に不要な開発アセットや DB ファイルが含まれているか
- [ ] `./scripts/verify-all.sh` が全てグリーンで通過したか
