# 開発規約 (AGENTS.md)

プロダクトの概要や開発背景は [README.md](./README.md)、設計方針の詳細・決定経緯は [ADR インデックス](./adr/README.md) を参照すること。

## 1. 主要な設計方針 (ADR)

- **型定義の一元管理**: `pkg/model/` をマスターとし、TS型やDDLは自動生成する ([ADR-0004](./adr/0004-go-schema-as-single-source-of-truth.md))
- **サロゲートキー**: 全エンティティで TypeID を使用し、自然キーは使わない ([ADR-0001](./adr/0001-surrogate-key-typeid-uuidv7.md), [ADR-0006](./adr/0006-domain-schema-and-typeid-structure.md))
- **単一バイナリ**: フロントエンド静的アセットを Go バイナリに内包して配信する ([ADR-0002](./adr/0002-backend-go-architecture.md), [ADR-0011](./adr/0011-directory-structure-and-responsibility-boundaries.md))
- **Local-First**: ライブ会場の圏外でもクライアント単独でクイズが完結する ([ADR-0007](./adr/0007-local-first-offline-pwa-architecture.md))
- **不変画像**: `mem_<uuid>.webp` と永続キャッシュにより CDN パージを不要にする ([ADR-0008](./adr/0008-immutable-image-caching-and-zero-purge.md))
- **動的マスタ**: グループやメンバーをコード内に固定せず、DBで動的管理する ([ADR-0006](./adr/0006-domain-schema-and-typeid-structure.md))
- **最小構成**: 環境変数は 4 つ、エラーコードは 6 つに固定する ([ADR-0014](./adr/0014-error-handling-and-minimal-problem-details.md), [ADR-0015](./adr/0015-minimal-configuration-and-secrets-management.md))
- **リポジトリ集約**: 上位ロジックは具象DBではなく `pkg/model/repository.go` のインターフェースに依存し、DB層は `pkg/model` の構造体に直接マッピングする ([ADR-0017](./adr/0017-repository-interface-and-pure-go-sqlite-architecture.md))

## 2. 協業プロセス（AI Slop の抑止）

### 要求を捉える

- ユーザーが求める結果、指定した振る舞い、制約を区別して把握する。指定されていない実装手段は、要求を満たす範囲で選ぶ。
- 会話中の補足や訂正は、以後の設計・実装・検証に反映する。すでに明示された希望を繰り返し確認しない。
- 判断に必要な情報は、まず関連するコード・資料・実際の挙動から確認する。それでも要求の解釈が分かれ、成果物に実質的な違いが生じる場合は、論点を絞って確認する。

### 実装を選ぶ

- 既存の実装や体験が参照された場合は、何を踏襲する要求なのかを確認する。部品や名称が共通しているだけで、要求を満たしたと判断しない。
- 再利用・拡張・新規実装は、要求への適合性、既存機能への影響、保守負担から選ぶ。既存実装の維持も、新規実装への置き換えも、それ自体を目的にしない。
- 実装上の困難を理由に、指定された操作や機能を黙って変更・省略しない。要求の変更が必要な場合は、制約と影響を説明して合意する。

### 結果を確かめる

- 検証では、ユーザーが指定した操作・結果・制約と、完成したものを照合する。テストの通過や部品の存在だけを達成の根拠にしない。
- 見た目や操作感が要求に含まれる場合は、実際に表示・操作して確認する。比較対象が指定されている場合は、その対象との違いも確認する。
- 不一致が見つかった場合は、実装が動いていても未完了として扱い、修正する。確認できない事項は、確認済みの結果と分けて報告する。
- 完了報告では、何を実現し、何を根拠に確認したかを簡潔に示す。品質に関する断定は、確認できた範囲に限る。

### 提案・承認・変更の手順

- **一括生成の禁止**: ドキュメント作成や設計時、一度に全体を一括出力してはならない。
- **セクション単位の段階的承認**: まず見出し構成を合意し、セクション単位でドラフトを提示してユーザーの認可（フィードバック・承認）を得ながら 1 つずつ埋めていく。
- **案のターミナル先行提示**: ADR・設計・実装方針の案は、ファイル編集やコード変更の前にターミナルへ提示し、ユーザーの承認後に反映する。
- **事実とポインタの優先**: 概念の教科書的解説や陳腐化するバージョン表記を記述しない。事実、数値、決定事項、および ADR へのポインタのみを簡潔に記述する。
- **再利用ファーストと事前提案**: 新しい API・型・抽象化を追加する前に、既存の API・DTO・Repository・フロント状態を棚卸しし、オプション追加や後方互換な拡張を先に比較する。ADR にない新規設計は理由・代替案・影響を提示し、認可後に実装する。
- **CI実行完了の待機禁止**: ローカル検証（`./scripts/verify-all.sh` / `actionlint`）をパスして push / PR 作成した後は、リモート CI の完了を待つポーリング（`gh run watch` や `sleep` 等）を行わず、即座に次の作業やユーザー報告へ移ること。
- **PRの自動マージ禁止**: ローカル検証をパスして push / PR 作成を行った段階でユーザーに報告すること。マージは必ずユーザーからの明示的な指示・認可を受けてから実行し、エージェントが独断でマージしてはならない。
- **PR 経由の変更徹底（main 直接 push の禁止）**: すべてのコード修正、機能追加、リファクタリング、およびマニフェスト更新は、必ずトピックブランチ（`feat/...`, `fix/...`, `chore/...` 等）を作成して PR 経由で行うこと。`main` ブランチへの直接コミット・直接 push は厳格に禁止する。

## 3. ADR と Skill の対応規約 (Governance & Enforcement)

- **決定と執行の分離**: アーキテクチャの意思決定・背景は `adr/` に記録し、その遵守・機械的検証は対応する Skill（`.agents/skills/`）が担う (`Skill ||--< ADR`)。
- **管轄 ADR の明記**: 新しい Skill を作成する際は、どの ADR を保護・執行するためのスキルかを冒頭に明記すること。あらゆるルールを 1 つの Skill に集約せず、ADR の目的に応じて疎結合に分割する。

## 4. 技術スタック

- **バックエンド**: Go（標準ライブラリ中心）
- **データベース**: SQLite（WAL モード、ドライバ: `modernc.org/sqlite`）
- **フロントエンド**: Next.js / React（完全静的エクスポート `output: 'export'`）
- **UI / スタイル**: Mantine, Tabler Icons
- **オフラインストレージ**: IndexedDB（ブラウザ内データ保持）
- **インフラ**: 自宅 Kubernetes（単一 Pod 構成）、Cloudflare Tunnel、Cloudflare R2（Litestream による DB バックアップ）

## 5. ディレクトリ構成

```
penlight-v2/
├── .agents/skills/ # プロジェクト固有の執行スキル (Skill ||--< ADR)
├── adr/            # アーキテクチャ意思決定記録 (ADR)
├── docs/notes/     # 調査・設計検討用ノート (ADR昇格前の壁打ち領域)
├── cmd/server/     # main.go（エントリーポイント、依存関係の注入とサーバー起動）
├── pkg/
│   ├── model/      # 型定義のマスター（外部依存ゼロのドメインモデル・DTO・エラー型）
│   ├── config/     # 環境変数ローダー
│   ├── quiz/       # クイズ出題・採点ロジック
│   └── repository/ # SQLite (WAL) データアクセス層
├── migrations/     # SQLite テーブル定義
├── frontend/       # Next.js 静的フロントエンド
├── deploy/         # Kubernetes デプロイ用マニフェスト
├── scripts/        # 型定義・スキーマの自動生成スクリプト
└── api/            # API 仕様書（OpenAPI）
```

## 6. クイックスタート

外部ミドルウェア（PostgreSQL や Redis 等）の立ち上げは不要。ローカル環境では環境変数の設定なしでそのまま起動できる。

### 開発環境の一括管理 (Make コマンド)

バックエンド（Go / `:8080`）とフロントエンド開発サーバー（Next.js / `:3000`）をバックグラウンドプロセスとして一括管理できる。ポートのリッスン待機や多重起動防止・ゾンビプロセスの確実な解放（PID & ポートキル）が自動で行われる。

```bash
# 基本操作（一括）
make up             # バックエンド・フロントエンドを両方起動（ポート待機後にステータス表示）
make status         # 各プロセスの稼働状態（RUNNING / STOPPED、PID、ポート、URL）を確認
make restart        # 両プロセスを安全に一括再起動
make down           # 全プロセスを一括停止（ポート強制解放を含む）

# ログ確認
make logs           # バックエンド・フロントエンド両方のログをリアルタイム表示 (Ctrl+C で抜ける)
make logs-back      # バックエンドのログのみ表示
make logs-front     # フロントエンドのログのみ表示

# コンポーネント別の個別管理
make up-back        # バックエンドのみ起動 (:8080)
make down-back      # バックエンドのみ停止
make restart-back   # バックエンドのみ再起動
make up-front       # フロントエンドのみ起動 (:3000)
make down-front     # フロントエンドのみ停止
make restart-front  # フロントエンドのみ再起動
make up-argocd      # ArgoCD Application を適用・起動
make down-argocd    # ArgoCD Application を削除・停止
make restart-argocd # ArgoCD Application を再作成・再起動

# 開発・検証ショートカット
make verify         # 全自動一括検証 (詳細ログ表示)
make verify-ai      # 全自動一括検証 (AI用・成功時は1行のみ出力しトークン節約)
make test           # Go テスト実行 (go test ./...)
make gen            # スキーマ・ER図一括再生成 (./scripts/generate-all.sh)
```

### 単一バイナリでの個別起動 (本番互換)

```bash
go run ./cmd/server
```

起動後、`http://localhost:8080` でアクセス可能（カレントディレクトリの `./data` 配下に SQLite ファイルと画像フォルダが自動生成される）。

### 環境変数仕様

設定可能な項目は以下の 4 つに限定する。ポート（8080）やタイムアウト値などはコード内定数に固定されており、環境変数化しない。

| 変数名 | 必須 | デフォルト値 | 説明 |
|---|---|---|---|
| `DATA_DIR` | 任意 | `./data` | SQLite データベースおよび画像アセットの格納ディレクトリ |
| `SESSION_SECRET` | 本番のみ | (ローカル用固定文字列) | Cookie 署名用シークレットキー (32バイト以上) |
| `GOOGLE_CLIENT_ID` | 任意 | 空（ゲストモード） | Google OIDC 認証用クライアント ID |
| `GOOGLE_CLIENT_SECRET` | 任意 | 空（ゲストモード） | Google OIDC 認証用シークレット |

## 7. 変更後の検証手順

コードを変更した際は、以下のコマンドで静的解析とテストの通過を確認すること。

```bash
# 【推奨】全自動一括検証（スキーマ再生成・同期検証・テスト・Biome・typos・actionlint）
make verify-ai      # AIエージェント実行時（成功時は1行のみ、エラー時のみ詳細出力でトークンを大幅節約）
make verify         # 人間実行時（全ステップの詳細ログを表示）
# または直接実行: ./scripts/verify-all.sh [--ai]

# バックエンドの静的解析とテスト（個別実行）
go vet ./...
go test ./...

# pkg/model/ を変更した場合（型定義・スキーマ・ER図の一括再生成）
./scripts/generate-all.sh

# コード整形・構文リント・誤字脱字の自動修復 (Ref: ADR-0013)
biome check --write
typos -w
```

## 8. 用語集

- **TypeID**: エンティティの種類を表す接頭辞が付いた一意な識別子（例: メンバーは `mem_<uuid>`、グループは `grp_<uuid>`）。名前重複や改名によるデータの破損を防ぐために使用する。
- **WAL (Write-Ahead Logging)**: SQLite の動作モードの一つ。変更を専用ログに先行追記することで、読み取り処理と書き込み処理を互いに邪魔させずに同時に実行できる。
- **Litestream**: SQLite の更新差分をリアルタイムに検知し、外部ストレージ（Cloudflare R2 等）へ秒単位で自動バックアップするツール。
