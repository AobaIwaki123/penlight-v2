---
id: ADR-0017
title: データアクセス層におけるリポジトリインターフェース集約および Pure Go SQLite 直接実装の採用
status: Accepted
scope: System
primary_category: DAT
categories: [DAT, ARC, DOM]
tags: [go, sqlite, single-source-of-truth, yagni]
deciders: [user, ai]
date: 2026-10-03
---

# 0017. データアクセス層におけるリポジトリインターフェース集約および Pure Go SQLite 直接実装の採用 (0017-repository-interface-and-pure-go-sqlite-architecture.md)

- **ステータス**: 承認 (Accepted)
- **日付**: 2026-10-03

---

## 1. 背景と解決すべき課題 (Context & Problem)

[ADR-0004](./0004-go-schema-as-single-source-of-truth.md) では `pkg/model/*.go` をドメインモデルの Single Source of Truth (SSoT) と定めている。
一方、データアクセス層において [ADR-0013](./0013-typescript-ai-agent-driven-development-toolchain.md) で選定された `sqlc` を導入したところ、以下の課題が生じた。

1. **生 SQL ファイル (`queries.sql`) の外出し管理コスト**:
   - サーバー側で実行する SQL はマスタ全件取得（3本）と回答ログ保存（1本）の計 4 本程度に過ぎないにもかかわらず、独立した SQL ファイルを別個に管理・同期する認知負荷が高い。
2. **モデル型の二重化問題**:
   - `sqlc` が `pkg/repository/models.go` に独自の構造体群を自動生成するため、正本である `pkg/model` の構造体と乖離し、型の二重管理が生じた。
3. **ORM (Prisma / GORM) 検討時の技術的制約**:
   - Prisma は Node.js 専用であり、単一バイナリ・軽量 Kubernetes 運用 ([ADR-0002](./0002-backend-go-architecture.md), [ADR-0011](./0011-directory-structure-and-responsibility-boundaries.md)) の前提と衝突する。
   - GORM は CGO 必須（`mattn/go-sqlite3`）であり、`CGO_ENABLED=0` によるクロスコンパイルや極小コンテナビルド ([ADR-0005](./0005-database-selection-sqlite-wal.md)) を阻害し、リフレクション多用により型安全性が低下する。

---

## 2. 決定事項と具現化仕様 (Decision & Specification)

`sqlc` および ORM を不採用とし、**Go 標準 `database/sql` + Pure Go `modernc.org/sqlite` による直接実装と、`pkg/model/repository.go` への公開インターフェース集約** を正式採用する。

### 構成アーキテクチャ

```
[呼び出し元: pkg/quiz, cmd/server]
       │ 依存
       ▼
[pkg/model/ (正本ドメイン層)]
  ├── group.go, member.go ... (正本構造体: SSoT)
  └── repository.go (公開インターフェース: Repository)
       ▲
       │ 実装 (implements)
[pkg/repository/ (データアクセス層)]
  └── sqlite.go (Pure Go modernc.org/sqlite WAL 接続 & Repository 実装)
```

1. **ドメイン集約インターフェース ([`pkg/model/repository.go`](../pkg/model/repository.go))**:
   - アプリ全体で参照するデータアクセス契約として `Repository` インターフェースを定義。
   - Go 標準の `context.Context` と `pkg/model` 内の型のみを使用し、**外部依存ゼロ** を厳格に維持。
2. **Pure Go SQLite 具象リポジトリ ([`pkg/repository/sqlite.go`](../pkg/repository/sqlite.go))**:
   - `modernc.org/sqlite`（Pure Go, CGO 不要）を使用。
   - SQLite WAL モード・BUSY タイムアウト（5秒）・外部キー制約を接続時に有効化。
   - クエリ結果を [`pkg/model`](../pkg/model) の正本構造体に直接マッピング（`rows.Scan`）し、独自モデルを一切持たない。
3. **不要ツールの廃止・クリーンアップ**:
   - `sqlc.yaml`, `pkg/repository/queries.sql`, および `sqlc` が出力した生成コード群を全廃。

---

## 3. 採否の根拠とトレードオフ (Rationale & Trade-offs)

### データアクセスアプローチの比較検討

| 評価項目 | Prisma | GORM | sqlc | Go 標準直接実装 [採択] |
|---|---|---|---|---|
| **言語・ランタイム** | Node.js (Go不可) | Go | Go | **Go 標準** |
| **Pure Go (CGO不要)** | - | ✕ (CGO必須) | ◯ | **◎ (完全対応)** |
| **SQL ファイル外出し** | 不要 | 不要 | **必須 (`queries.sql`)** | **不要 (Go内完結)** |
| **型の一元化 (ADR-0004)** | ✕ | △ (リフレクション) | △ (overrides必要) | **◎ (pkg/model 直結)** |
| **ファイル数・依存** | 巨大 | 巨大 | 外部ツール+生成6ファイル | **最少 (sqlite.go 1ファイル)** |
| **認知負荷** | - | 中 | 高 (SQL/yaml/Go往復) | **最低 (Goコードのみ)** |

- **決定打**: 本プロダクトは「Local-First」であり、サーバー側で必要な SQL はマスタ取得と回答ログ保存のわずか 4〜5 本のみである。この規模に対して外部コード生成ツール（`sqlc`）や巨大 ORM（GORM）を持ち込むのは YAGNI（過剰設計）であり、Go 標準の直接実装を採用することで最も依存が少なく、`pkg/model` の正本型とダイレクトに結合した堅牢な設計となる。

---

## 4. 得られる効果と留意点 (Consequences)

### ポジティブな影響 (Positive)
- **型の二重管理の完全撲滅**: `pkg/model` の構造体をそのまま `rows.Scan` で受け取るため、型定義は 1 箇所のみ。
- **ファイル構成の極限のスリム化**: `sqlc` 関連の 6 ファイルを廃止し、`pkg/repository/sqlite.go` 1 ファイルに完結。
- **依存性逆転 (DIP) の確立**: 上位ロジック（`pkg/quiz` 等）は `model.Repository` のみに依存するため、単体テスト時のモックが容易。
- **クロスコンパイル容易性**: Pure Go 実装（`CGO_ENABLED=0`）を完全維持。

### 留意点と対策 (Negative & Mitigation)
- **`rows.Scan` の手書き**:
  - 対策: クエリが 4 本（対象エンティティ 4 つ）のみであり、手動マッピングは全体で 30〜40 行程度に収まるため保守リスクは極小。カラム追加時は `go test` で即座に検知する。
