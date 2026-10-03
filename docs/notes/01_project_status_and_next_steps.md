# プロジェクト現在地と次の実装タスク (docs/notes/01)

本ドキュメントは、セッション間の開発引き継ぎおよび現在の実装フェーズ・直近タスクを明確にするための引き継ぎノートである。

---

## 1. 現在の開発フェーズ

- **Phase 0（アーキテクチャ設計・ドキュメント・ガバナンス基盤）**: **完了**
- **Phase 1（スキーマ自動生成パイプライン・フロントエンド足場）**: **直近の着手タスク**
- **Phase 2（バックエンド コア実装・サーバー起動）**: 後続タスク

---

## 2. 完了していること (Phase 0: Done)

1. **ドキュメント・規約整備**:
   - `README.md`: アプリ概要、特徴、v1からの刷新点（技術用語なし）、ドキュメントリンク
   - `AGENTS.md`: 7大設計方針、協業プロセス（AI Slop抑止）、ADRとSkillの対応規約、技術スタック、ディレクトリ構成、起動手順、検証手順、用語集
   - `adr/`: 全16件のADR（0001〜0016）および連番一覧表（`adr/README.md`）
2. **正本モデル層・スキーマ**:
   - `pkg/model/*.go`: 外部依存ゼロのドメインモデル・DTO・エラー型（SSoT）
   - `pkg/config/config.go`: 4つの最小環境変数ローダー
   - `migrations/000001_init.up.sql`: SQLite テーブル定義 (DDL)
   - `api/openapi.yaml`: OpenAPI 3.0 仕様書
   - `scripts/gen-er-diagram.go`: Go AST 解析による ER 図自動生成スクリプト
   - `assets/schema/er-diagram.md`: 自動生成された ER 図アセット
3. **執行スキル (`.agents/skills/`)**:
   - `doc-lifecycle`: note壁打ちから認可に基づくADR昇格・不変性管理
   - `schema-sync`: Goモデル一元管理と派生ファイル手書き編集禁止
   - `typeid-guard`: 不変サロゲートキー (TypeID) 強制
   - `api-boundary`: 4環境変数・6大エラーコード強制
   - `local-first-pwa`: Local-First出題・バッチ同期・静的エクスポート・画像不変キャッシュ
   - `~/.agents/skills/skill-sanitizer`: グローバル品質監査・AI Slop抑止スクリプト
4. **Git 設定**:
   - `.gitignore` 配備完了

---

## 3. 次のセッションで着手すべきタスク (Phase 1: Immediate Tasks)

### タスク 1: スキーマ一括生成スクリプトの実装 (`scripts/generate-all.sh`)
- **目的**: `AGENTS.md` の検証コマンドに記載されている `./scripts/generate-all.sh` を実体化する (ADR-0004)。
- **内容**:
  1. `go run scripts/gen-er-diagram.go` の実行
  2. `tygo` を用いた Go 構造体からの TypeScript 型定義自動出力 (`frontend/src/types/generated.ts`)
  3. `go vet ./...` および `go test ./...` の一括実行

### タスク 2: フロントエンド初期足場の構築 (`frontend/`)
- **目的**: Next.js 静的エクスポート (`output: 'export'`) の基盤作成 (ADR-0002, ADR-0011)。
- **内容**:
  1. Next.js プロジェクトの初期化
  2. Mantine / Tabler Icons のセットアップ
  3. `frontend/src/types/generated.ts`（自動生成された型）の読み込み確認

---

## 4. 後続の開発タスク (Phase 2: Core Backend)

1. **SQLite WAL リポジトリ層の実装 (`pkg/repository/`)**:
   - `modernc.org/sqlite` による接続プール・WAL 有効化
   - グループ・メンバー・カラーの読み込みクエリ
   - 回答ログのバッチ保存 (`INSERT OR IGNORE`)
2. **クイズ出題ロジックの実装 (`pkg/quiz/`)**:
   - Strategy パターンによる出題エンジン（完全ランダム、CIELAB 色差類似色、期生別等）
3. **サーバー起動エントリーポイントの実装 (`cmd/server/main.go`)**:
   - HTTP ルーティング（Go 標準 `net/http`）
   - フロントエンド静的ファイルの `embed.FS` 配信
   - `go run ./cmd/server` によるローカル起動検証
