---
name: schema-sync
description: Go構造体（pkg/model/）を唯一のマスターとし、TypeScript型・DDL・OpenAPI・ER図への一括同期と二重管理防止を統制する。
---

# スキーマ一元管理・同期規約 (schema-sync)

> **管轄 ADR**: [ADR-0004](../../../adr/0004-go-schema-as-single-source-of-truth.md), [ADR-0017](../../../adr/0017-repository-interface-and-pure-go-sqlite-architecture.md)

本スキルは、Go 構造体（`pkg/model/`）を Single Source of Truth（マスター）とし、派生コードの整合性を維持するための規約と手順を定める。

## 絶対遵守ルール (Invariants)

1. **派生成果物の手動編集は厳禁**:
   - `frontend/src/types/generated.ts`（TS型定義）
   - `assets/schema/er-diagram.md`（ER図）
   - `api/openapi.yaml`（API仕様書）
   上記ファイルを手書きで編集してはならない。必ず `pkg/model/*.go` を編集し、生成スクリプト経由で出力すること。
2. **モデル層の外部依存ゼロ**:
   - `pkg/model/` は純粋なドメイン型・DTO・エラー型、および `Repository` インターフェースのみを定義し、外部サードパーティライブラリや DB ドライバに依存しない（Go 標準 `context.Context` のみ許容）。
3. **リポジトリ層の独自モデル排除 (Zero-Duplication)**:
   - `pkg/repository/` 内に独自 struct（中間 DTO や DB 専用モデル）を作ってはならない。必ず `pkg/model` の正本構造体に直接 `rows.Scan` し、二重管理を根絶する。
4. **依存性逆転 (DIP) の厳守**:
   - 上位レイヤ（`pkg/quiz`, `cmd/server`）は具象 DB（`pkg/repository.SQLiteRepository`）に直接依存せず、`pkg/model.Repository` インターフェースにのみ依存すること。

---

## スキーマ変更・生成コマンドリスト

`pkg/model/` を変更した際や、派生成果物（TS型・ER図）を同期する際は、以下のコマンドを使用する。

```bash
# 【基本】スキーマ・型定義・ER図の一括再生成と検証（最優先）
./scripts/generate-all.sh

# 【差分検証】生成漏れや手動改変がないか確認（CI / コミット前）
git diff --exit-code frontend/src/types/generated.ts assets/schema/

# 【初回環境セットアップ】tygo が見つからない場合
go install github.com/gzuidhof/tygo@latest
```

> **Note (tygo の設定)**: `tygo.yaml` の `path` には、Go モジュールのフルパス（`github.com/aobaiwaki/penlight-v2/pkg/model`）を指定すること（`pkg/model` だと標準パッケージ探索になりエラーとなる）。

---

## チェックリスト（変更後）

- [ ] `pkg/model/` 以外の派生成果物を直接手書きで書き換えていないか
- [ ] `./scripts/generate-all.sh` が正常終了し、`frontend/src/types/generated.ts` と `assets/schema/` が同期されたか
- [ ] `git diff --exit-code` で意図しない破壊的変更がないか確認したか
- [ ] 新規エンティティに適切な TypeID プレフィックス（`grp_`, `col_`, `mem_` 等）が付与されているか
