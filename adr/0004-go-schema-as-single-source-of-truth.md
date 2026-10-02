# 0004. Go 構造体を正本とするスキーマ一元管理および自動生成パイプラインの採用 (0004-go-schema-as-single-source-of-truth.md)

- **ステータス**: 承認 (Accepted)
- **日付**: 2026-10-03

---

## 1. 背景と解決すべき課題 (Context & Problem)

一般的な Web 開発では、以下の複数レイヤでデータスキーマが個別・手動で管理され、不整合やドキュメントの形骸化が頻発します。
1. バックエンドのデータモデル定義
2. フロントエンドの TypeScript 型定義
3. データベースの DDL / マイグレーション
4. API 仕様書（OpenAPI / Swagger / JSON Schema）
5. 入力値バリデーションルール

これらを手書きで同期させようとすると、青葉坂46 等の新グループや新期生追加時に修正漏れが発生し、AI エージェントが古い型定義に基づいて誤ったコードを生成する原因となります。

---

## 2. 決定事項と具現化仕様 (Decision & Specification)

`backend/pkg/model/*.go` の **Go 構造体を唯一の正本（Single Source of Truth: SSoT）** と定め、ツールチェーンによってすべての派生成果物を自動生成する。

```
[Go Struct (正本: pkg/model/*.go)]
       │
       ├── (tygo) ──────────────► frontend/src/types/generated.ts (TS型)
       ├── (invopop/jsonschema) ─► api/schema/*.json (JSON Schema / 動的Admin UI)
       ├── (swag / ogen) ───────► api/openapi/openapi.yaml (OpenAPI 3.1)
       └── (scripts/gen-er) ────► assets/schema/er-diagram.md (Mermaid ER図)
```

1. **手動編集の禁止**: フロントエンドの `generated.ts` や OpenAPI 仕様書を手動編集することを厳禁とし、変更は必ず Go 構造体から行う。
2. **ワンコマンド同期**: `scripts/generate-all.sh` で全成果物を 1 秒で再生成。
3. **CI 検証**: GitHub Actions で `generate-all.sh` を実行し、`git diff --exit-code` で未生成コミットを自動 reject。

---

## 3. 採否の根拠とトレードオフ (Rationale & Trade-offs)

### スキーマ管理アプローチの比較検討

| アプローチ | 正本 (SSoT) | メリット | デメリット・課題 | 判定 |
|---|---|---|---|---|
| **OpenAPI ファースト** | YAML / JSON | 言語中立、仕様先行 | YAML 記述が極めて冗長、DBタグや内部ロジックの表現が困難 | **見送り**: 開発体験悪化 |
| **Prisma / ORM ファースト** | `schema.prisma` | DB と型が強固に一致 | Node.js 依存、Go バックエンドとの親和性が低い | **却下**: アーキテクチャ不整合 |
| **Go Struct コードファースト [採択]** | **Go 構造体 (`pkg/model`)** | **自己文書化、型安全、軽量、AI 協業に最適** | codegen スクリプトの整備が初期に必要 | **採用**: 開発速度と厳密性の両立 |

- **決定打**: Go の構造体タグ（`json`, `db`, `validate`）と GoDoc コメントは、コード自体が仕様書として機能し、追加の外部言語や冗長な YAML を書く必要がない。`tygo`（Go AST 直接解析）により、外部依存ゼロで忠実な TypeScript 型が得られる。

---

## 4. 得られる効果と留意点 (Consequences)

### ポジティブな影響 (Positive)
- **多重管理の完全撲滅**: Go 構造体を 1 箇所修正するだけで、フロントエンド型・API ドキュメント・バリデーションが完全同期。
- **AI 協業精度の最大化**: AI エージェントに「Go 構造体が正本である」と指示するだけで、フロントエンド・バックエンドの双方でブレのないコードが生成される。

### 留意点と対策 (Negative & Mitigation)
- **生成漏れリスク**:
  - 対策: pre-commit hook および GitHub Actions CI で未同期のコミットを自動検出してブロックする。
