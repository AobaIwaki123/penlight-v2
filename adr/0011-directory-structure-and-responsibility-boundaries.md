# 0011. プロジェクトディレクトリ構成および責務境界の策定 (0011-directory-structure-and-responsibility-boundaries.md)

- **ステータス**: 承認 (Accepted)
- **日付**: 2026-10-03

---

## 1. 背景と解決すべき課題 (Context & Problem)

Go バックエンド、TypeScript フロントエンド、Kubernetes マニフェスト、およびコード生成スクリプトが同一リポジトリに同居するモノレポ構成において、以下の問題が生じます。
1. **パッケージの循環参照・境界汚染**: どのディレクトリが最上流で、どのファイルが自動生成物（手動編集禁止）なのかが不明確。
2. **AI エージェントの迷走**: AI がフロントエンドの自動生成ファイルを直接書き換えてしまい、次回のビルドで上書きされて変更が消失する事故。
3. **ビルド成果物の肥大化**: バックエンドとフロントエンドを別々にデプロイすると、インフラ運用コストが倍増する。

---

## 2. 決定事項と具現化仕様 (Decision & Specification)

Standard Go Layout およびモダンフロントエンドの関心分離（SoC）に準拠した以下の階層構造および単方向依存ルールを正式採用する。

### 構成ツリーと責務定義

```
space/penlight-v2/
├── backend/                   # [Go サーバー]
│   ├── cmd/server/main.go     # エントリーポイント
│   ├── pkg/model/             # 【最上流正本】ドメインモデル・ID規約 (外部依存ゼロ)
│   ├── pkg/quiz/              # 出題エンジン & Strategy パターン
│   ├── pkg/repository/        # SQLite WAL リポジトリ
│   ├── pkg/auth/              # Google OIDC 認証
│   ├── pkg/image/             # WebP 最適化パイプライン
│   ├── migrations/            # SQLite DDL
│   └── embedded/              # フロントエンド dist を内包する embed.FS
├── frontend/                  # [React 19 / Next.js 静的エクスポート]
│   ├── src/types/generated.ts # 【自動生成】tygo が出力した TS 型 (手動編集厳禁)
│   ├── src/features/quiz/     # クイズ画面 & クライアント出題純粋関数
│   ├── src/features/offline/  # IndexedDB キャッシュ & Outbox 同期
│   └── src/features/admin/    # 動的 Admin フォーム
├── deploy/                    # 自宅 k8s マニフェスト (Kustomize: base / overlays)
├── scripts/                   # generate-all.sh, gen-er-diagram.go
└── assets/                    # schema/er-diagram.md (自動生成ドキュメント資産)
```

1. **単方向コード生成フロー**: `backend/pkg/model` (正本) ──► `frontend/src/types/generated.ts`（逆流や手書き編集は禁止）。
2. **単一バイナリ Pod 運用**: フロントエンド静的アセットを `backend/embedded/` で内包し、1 Pod で完結。
3. **実行ランタイムと責務境界の厳格な分離**:
   - **サーバーサイド (Go)**: 自宅 k8s 上で稼働。DB管理、静的アセット・画像プロキシ配信、OIDC認証、オフライン回答ログのバッチ受付・永続化を担う（端末・クライアント内では動作しない）。
   - **クライアントサイド (TypeScript / Next.js PWA)**: ユーザーのブラウザ（端末）上で稼働。IndexedDB マスタデータを用いた Local-First 完全オフライン出題・採点、UIレンダリング、Service Worker キャッシュを担う。

---

## 3. 採否の根拠とトレードオフ (Rationale & Trade-offs)

### プロジェクト配置構成の比較検討

| 構成案 | デプロイ形態 | 型同期の容易性 | AI 開発の安定度 | 判定 |
|---|---|---|---|---|
| **完全分離 (2 Pod / 2 リポジトリ)** | 独立デプロイ | 煩雑 (npm パッケージ配布等) | 低い (コンテキスト分断) | **却下**: 運用コスト倍増 |
| **Node.js 側で Go バイナリ呼出** | 1 Pod | 不安定 | 低い | **却下**: アーキテクチャ歪曲 |
| **Go 内包モノレポ (単一バイナリ) [採択]** | **1 Pod (Go embed.FS)** | **直接ファイル生成 (tygo)** | **極めて高い (明確な境界)** | **採用**: 最小運用・最大整合 |

- **決定打**: フロントエンドを静的ビルドして Go バイナリに埋め込むことで、k8s マニフェストを 1 つにまとめられ、ネットワーク遅延もゼロになる。最上流（Go モデル）と最下流（フロントエンド型）が 1 つのワークスペース内で完結するため、AI のコンテキスト把握が最もスムーズになる。

---

## 4. 得られる効果と留意点 (Consequences)

### ポジティブな影響 (Positive)
- **AI の破壊的編集の防止**: 「`generated.ts` は手動編集禁止、`pkg/model` を編集して `generate-all.sh` を叩く」という単一ルールにより手戻りを根絶。
- **デプロイの単純化**: 1 つのコンテナイメージ（15〜30MB）をプッシュ・デプロイするだけでフロントとバックが同時に最新化。

### 留意点と対策 (Negative & Mitigation)
- **ビルド順序の依存**:
  - フロントエンドのビルド（`npm run build`）を先に行い、その成果物を Go コンパイル時に `embed.FS` に取り込むパイプライン順序を `Dockerfile` および CI で厳格に定義する。
