# 開発規約 (AGENTS.md)

## 1. 主要な設計方針 (ADR)

プロダクトの概要や開発背景は [README.md](./README.md)、設計方針の詳細・決定経緯は [ADR インデックス](./adr/README.md) を参照すること。

- **型定義の一元管理**: `pkg/model/` をマスターとし、TS型やDDLは自動生成する ([ADR-0004](./adr/0004-go-schema-as-single-source-of-truth.md))
- **サロゲートキー**: 全エンティティで TypeID を使用し、自然キーは使わない ([ADR-0001](./adr/0001-surrogate-key-typeid-uuidv7.md), [ADR-0006](./adr/0006-domain-schema-and-typeid-structure.md))
- **単一バイナリ**: フロントエンド静的アセットを Go バイナリに内包して配信する ([ADR-0002](./adr/0002-backend-go-architecture.md), [ADR-0011](./adr/0011-directory-structure-and-responsibility-boundaries.md))
- **Local-First**: ライブ会場の圏外でもクライアント単独でクイズが完結する ([ADR-0007](./adr/0007-local-first-offline-pwa-architecture.md))
- **不変画像**: `mem_<uuid>.webp` と永続キャッシュにより CDN パージを不要にする ([ADR-0008](./adr/0008-immutable-image-caching-and-zero-purge.md))
- **動的マスタ**: グループやメンバーをコード内に固定せず、DBで動的管理する ([ADR-0006](./adr/0006-domain-schema-and-typeid-structure.md))
- **最小構成**: 環境変数は 4 つ、エラーコードは 6 つに固定する ([ADR-0014](./adr/0014-error-handling-and-minimal-problem-details.md), [ADR-0015](./adr/0015-minimal-configuration-and-secrets-management.md))

## 2. 技術スタック

- **バックエンド**: Go（標準ライブラリ中心）
- **データベース**: SQLite（WAL モード、ドライバ: `modernc.org/sqlite`）
- **フロントエンド**: Next.js / React（完全静的エクスポート `output: 'export'`）
- **UI / スタイル**: Mantine, Tabler Icons
- **オフラインストレージ**: IndexedDB（ブラウザ内データ保持）
- **インフラ**: 自宅 Kubernetes（単一 Pod 構成）、Cloudflare Tunnel、Cloudflare R2（Litestream による DB バックアップ）

## 3. ディレクトリ構成

```
penlight-v2/
├── adr/            # アーキテクチャ意思決定記録 (ADR)
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

## 4. クイックスタート

外部ミドルウェア（PostgreSQL や Redis 等）の立ち上げは不要。ローカル環境では環境変数の設定なしでそのまま起動できる。

```bash
cd backend
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

## 5. 変更後の検証手順

コードを変更した際は、以下のコマンドで静的解析とテストの通過を確認すること。

```bash
# バックエンドの静的解析とテスト
cd backend
go vet ./...
go test ./...

# pkg/model/ を変更した場合（型定義・スキーマ・ER図の一括再生成）
./scripts/generate-all.sh
```

## 6. 用語集

- **TypeID**: エンティティの種類を表す接頭辞が付いた一意な識別子（例: メンバーは `mem_<uuid>`、グループは `grp_<uuid>`）。名前重複や改名によるデータの破損を防ぐために使用する。
- **WAL (Write-Ahead Logging)**: SQLite の動作モードの一つ。変更を専用ログに先行追記することで、読み取り処理と書き込み処理を互いに邪魔させずに同時に実行できる。
- **Litestream**: SQLite の更新差分をリアルタイムに検知し、外部ストレージ（Cloudflare R2 等）へ秒単位で自動バックアップするツール。
