---
id: ADR-0025
title: マイグレーションSQLおよびシードデータの embed.FS 完全内包アーキテクチャ
status: Accepted
scope: Backend
primary_category: ARC
categories: [ARC, DAT]
tags: [go, sqlite, k8s, directory-structure]
deciders: [user, ai]
date: 2026-10-04
---

# 0025. マイグレーションSQLおよびシードデータの embed.FS 完全内包アーキテクチャ (0025-embed-sql-migrations-and-seeds-into-binary.md)

- **ステータス**: 承認 (Accepted)
- **日付**: 2026-10-04

---

## 1. 背景と解決すべき課題 (Context & Problem)

[ADR-0002](./0002-backend-go-architecture.md)（Go バックエンドアーキテクチャ）および [ADR-0011](./0011-directory-structure-and-responsibility-boundaries.md)（単一バイナリ配信）では、フロントエンド静的アセットを Go バイナリ内に内包し、外部依存ゼロの単一バイナリとして配布・実行する方針を定めている。

しかし、バックエンド初期化時のデータベースマイグレーション（`migrations/000001_init.up.sql`）およびマスタ同期用シードデータ（`seeds/seed.sql`、`seeds/data/image_sources.json`）の読み込み処理において、以下の設計課題が存在していた。

1. **ディスク相対パス探索への依存**:
   - `pkg/server/server.go` 内で `os.ReadFile` を用い、カレントディレクトリや親ディレクトリ（`..`, `../..`）を走査してファイルを取得していた。
2. **コンテナランタイム環境での起動失敗**:
   - マルチステージ Dockerfile の最小実行ステージ（`runner`）では、静的バイナリ `/app/server` のみをコピーしていたため、実行時ファイルシステム上に SQL ファイルが存在せず、Kubernetes 上で Pod が `CrashLoopBackOff`（`failed to read migration SQL`）となって起動不能に陥った。
3. **作業ディレクトリ（CWD）依存の脆さ**:
   - 実行時の作業ディレクトリがリポジトリルート以外の場合、ローカル開発環境や CLI からの実行時にもパス探索が失敗するリスクを抱えていた。

---

## 2. 決定事項と具現化仕様 (Decision & Specification)

### 1. `embed.FS` によるマイグレーションおよびシードデータの完全内包
Go 標準ライブラリの `embed` パッケージを採用し、SQL および JSON データをコンパイル時にバイナリ内部に組み込む。

- **`migrations/embed.go`**:
  ```go
  package migrations

  import "embed"

  // FS embeds all migration SQL files.
  //
  //go:embed *.sql
  var FS embed.FS
  ```
- **`seeds/embed.go`**:
  ```go
  package seeds

  import "embed"

  // FS embeds seed SQL and master data json files.
  //
  //go:embed seed.sql data/*.json
  var FS embed.FS
  ```

※ Go の `//go:embed` ディレクティブは親ディレクトリ（`..`）の走査を禁止しているため、各パッケージ内に `embed.go` を配置し、公開変数 `FS` を外部パッケージから参照する疎結合な設計とする。

### 2. サーバー起動時ローダーの改修 (`pkg/server/server.go`)
- マイグレーション適用時:
  - まず `migrations.FS.ReadFile("000001_init.up.sql")` から直接読み込む。
  - 存在しない場合のみ、ローカル開発互換のためディスクフォールバック（`resolveFile`）を試みる。
- シードデータ・マスタ更新時:
  - `seeds.FS.ReadFile("seed.sql")` および `seeds.FS.ReadFile("data/image_sources.json")` からバイナリ内リソースを展開する。

### 3. Dockerfile におけるビルド保証
- バックエンドビルドステージ（Stage 2）で `COPY migrations/ ./migrations/` および `COPY seeds/ ./seeds/` を行い、バイナリ内に確実にパックする。
- 実行ステージ（Stage 3）では `/app/server` バイナリのみをコピーする最小構成を維持し、攻撃対象領域（アタックサーフェス）を最小化する。

---

## 3. 結果とトレードオフ (Consequences)

### メリット
1. **完全なゼロ外部依存（Single Binary Invariant）の達成**:
   - 設定された `DATA_DIR` 以外に一切の外部ディスクリソース（SQL、JSON、HTML、JS）を要求せず、バイナリ単独でマイグレーション・シード投入・Web配信・API提供が完結する。
2. **実行時エラーの根絶**:
   - ファイルの未配置や作業ディレクトリ差異による起動クラッシュが構造上発生し得ない。
3. **コンテナセキュリティの強化**:
   - ランタイムイメージ内に不要なファイルやスクリプトを一切配置せず、 Alpine 最小構成かつ非 root ユーザーで安全に稼働できる。

### デメリット・トレードオフ
- **バイナリサイズのわずかな増加**:
  - SQL および JSON ファイル群（約 300KB）がバイナリに含まれるが、フロントエンド静的アセット（数MB）と比較して極めて小さく、実用上のオーバーヘッドは無視できる。
