---
name: schema-sync
description: Go構造体（pkg/model/）を唯一のマスターとし、TypeScript型・DDL・OpenAPI・ER図への一括同期と二重管理防止を統制する。
---

# スキーマ一元管理・同期規約 (schema-sync)

> **管轄 ADR**: [ADR-0004](../../../adr/0004-go-schema-as-single-source-of-truth.md)

本スキルは、Go 構造体（`pkg/model/`）を Single Source of Truth（マスター）とし、派生コードの整合性を維持するための規約と手順を定める。

## 絶対遵守ルール (Invariants)

1. **派生成果物の手動編集は厳禁**:
   - `frontend/src/types/generated.ts`（TS型定義）
   - `assets/schema/er-diagram.md`（ER図）
   - `api/openapi.yaml`（API仕様書）
   上記ファイルを手書きで編集してはならない。必ず `pkg/model/*.go` を編集し、生成スクリプト経由で出力すること。
2. **モデル層の外部依存ゼロ**:
   - `pkg/model/` は純粋なドメイン型・DTO・エラー型のみを定義し、外部サードパーティライブラリや DB ドライバに依存しない。

---

## スキーマ変更の手順

`pkg/model/` の構造体を変更した際は、必ず以下の手順で派生成果物を再同期すること。

```bash
# 1. ER図の自動再生成
go run scripts/gen-er-diagram.go

# 2. 静的解析とコンパイルチェック
go vet ./...
go test ./...
```

---

## チェックリスト（変更後）

- [ ] `pkg/model/` 以外の派生ファイルを直接手書きで書き換えていないか
- [ ] `go run scripts/gen-er-diagram.go` で ER 図が最新化されたか
- [ ] `go vet ./...` がエラーゼロで通過したか
- [ ] 新規エンティティに適切な TypeID プレフィックス（`id PK`）が付与されているか
