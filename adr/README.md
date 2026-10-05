# Architecture Decision Records (ADR)

技術選定および設計方針の意思決定記録。
新しい決定は連番（`0017-...`）で追記する。

## 1. 意思決定一覧

| 番号 | タイトル | ステータス |
|---|---|---|
| [0001](./0001-surrogate-key-typeid-uuidv7.md) | プレフィックス付き UUID v7 (TypeID) によるサロゲートキーの採用 | Accepted |
| [0002](./0002-backend-go-architecture.md) | バックエンド言語としての Go (Golang) の採用 | Accepted |
| [0003](./0003-deployment-target-and-container-registry.md) | デプロイ環境およびコンテナレジストリ選定 | Accepted |
| [0004](./0004-go-schema-as-single-source-of-truth.md) | Go 構造体を正本とするスキーマ一元管理および自動生成パイプラインの採用 | Accepted |
| [0005](./0005-database-selection-sqlite-wal.md) | データベースとしての SQLite (WAL モード) の採用 | Accepted |
| [0006](./0006-domain-schema-and-typeid-structure.md) | プレフィックス付きサロゲートキーと動的ドメインスキーマ構成の採用 | Accepted |
| [0007](./0007-local-first-offline-pwa-architecture.md) | ライブ会場での完全動作を保証する Local-First / オフライン PWA アーキテクチャの採用 | Accepted |
| [0008](./0008-immutable-image-caching-and-zero-purge.md) | 画像アセットの永久不変キャッシュ (RFC 8246 immutable) とゼロパージ運用の採用 | Accepted |
| [0009](./0009-quiz-generation-strategy-pattern.md) | クイズ出題アルゴリズムにおける Strategy パターンの採用 | Proposed |
| [0010](./0010-google-oidc-authentication-and-session-security.md) | Google OIDC 認証と HttpOnly セッション Cookie の採用 | Accepted |
| [0011](./0011-directory-structure-and-responsibility-boundaries.md) | プロジェクトディレクトリ構成および責務境界の策定 | Accepted |
| [0012](./0012-kubernetes-deployment-and-gitops-architecture.md) | Kubernetes デプロイおよび ArgoCD GitOps アーキテクチャの採用 | Accepted |
| [0013](./0013-typescript-ai-agent-driven-development-toolchain.md) | AI 駆動開発を支える機械的支援ツールチェーンおよび超高速自律修復サイクルの採用 | Accepted |
| [0014](./0014-error-handling-and-minimal-problem-details.md) | エラー設計および最小 Problem Details 規約の採用 | Accepted |
| [0015](./0015-minimal-configuration-and-secrets-management.md) | 最小環境変数および機密性分離アーキテクチャの採用 | Accepted |
| [0016](./0016-adr-governance-and-immutable-sequential-architecture.md) | ADR ガバナンスおよびフラット不変連番管理の採用 | Accepted |
| [0017](./0017-repository-interface-and-pure-go-sqlite-architecture.md) | データアクセス層におけるリポジトリインターフェース集約および Pure Go SQLite 直接実装の採用 | Accepted |
| [0018](./0018-quiz-candidate-pool-filtering-architecture.md) | クイズ出題における母集団フィルタリング設計の分離 | Accepted |
| [0019](./0019-quiz-format-strategy-and-color-palette-architecture.md) | 解答形式 Strategy と自由回答（カラーパレット選択）設計の採用 | Accepted |
| [0020](./0020-quiz-target-member-selection-strategy.md) | 出題対象メンバー選出におけるブレンドデッキ戦略の採用 | Accepted |
| [0021](./0021-gitops-master-data-synchronization-and-versioning-architecture.md) | GitOps マスタデータ同期アーキテクチャおよび複数画像・バージョン管理テーブルの採用 | Accepted |
| [0022](./0022-pluggable-quiz-ui-and-presentation-layout-architecture.md) | プラガブル解答インターフェースおよび表示レイアウト共存アーキテクチャの採用 | Accepted |
| [0023](./0023-master-data-sync-guarantee-and-verification.md) | マスタデータ差分同期保証および整合性機械検証スクリプトの採用 | Accepted |
| [0024](./0024-photo-metadata-human-in-the-loop-architecture.md) | 写真メタデータ（衣装・シングル種別）の人間参加型（HITL）策定プロセスの採用 | Accepted |
| [0025](./0025-embed-sql-migrations-and-seeds-into-binary.md) | マイグレーションSQLおよびシードデータの embed.FS 完全内包アーキテクチャ | Accepted |
| [0026](./0026-multi-series-hierarchy-and-isolation-architecture.md) | 同一アプリ内におけるシリーズ（Series）階層分離アーキテクチャの採用 | Accepted |
| [0027](./0027-song-penlight-color-data-structure.md) | 楽曲ペンライトカラー（1色/2色・左右なし）データ構造の採用 | Accepted |
| [0028](./0028-compliance-test-driven-adr-traceability-architecture.md) | ADR コンプライアンステスト駆動によるトレーサビリティおよび二重管理防止アーキテクチャ | Accepted |
| [0029](./0029-portal-and-filter-integrated-mode-architecture.md) | アプリエントリーポータルおよびフィルター統合型モード選択アーキテクチャの採用 | Accepted |
| [0030](./0030-portal-layout-and-visual-identity-ui.md) | ポータル画面レイアウトおよびビジュアル・アイデンティティ重視UIの採用 | Accepted |
| [0031](./0031-generic-quiz-engine-and-target-abstraction.md) | クイズ出題・判定エンジンのジェネリック抽象化アーキテクチャの採用 | Accepted |
| [0032](./0032-answer-log-multi-target-polymorphism-architecture.md) | 回答ログにおけるマルチターゲット多態性永続化アーキテクチャの採用 | Accepted |


---

## 2. メタデータ規約 (Frontmatter)

各 ADR の先頭には以下の Frontmatter を必須とする。

```yaml
---
id: ADR-0005
title: データベースとしての SQLite (WAL モード) の採用
status: Accepted
scope: System
primary_category: DAT
categories: [DAT, ARC]
tags: [sqlite, wal, litestream]
deciders: [user, ai]
date: 2026-10-03
---
```

---

## 3. 公式タグ統制一覧 (Controlled Tag Vocabulary)

タグの表記揺れや無秩序な増殖を防ぐため、**以下の公式許可タグからのみ選択（1つのADRにつき最大3〜4個）** すること。アドホックなタグの追加は禁止。

| 分類コード | カテゴリ | 公式許可タグ (Allowlist) | 対象スコープ |
|---|---|---|---|
| **ARC** | 基盤・デプロイ | `go`, `k8s`, `cloudflare`, `gitops`, `monorepo`, `directory-structure`, `configuration` | システム構造、実行言語、インフラ、環境設定 |
| **DOM** | ドメイン・スキーマ | `typeid`, `uuidv7`, `single-source-of-truth`, `surrogate-key`, `extensible-group` | 識別子設計、正本構造体、拡張グループ設計 |
| **DAT** | データ・ストレージ | `sqlite`, `wal`, `litestream`, `immutable-cache`, `s3-backup` | 永続化、ファイルI/O、キャッシュ制御 |
| **APP** | アプリ・機能 | `local-first`, `pwa`, `quiz-strategy`, `oidc`, `problem-details`, `cielab` | クイズアルゴリズム、認証、オフライン、エラー |
| **DEV** | 開発体験・規約 | `toolchain`, `biome`, `knip`, `adr-governance`, `yagni` | コード衛生、静的解析、ADR規約、設計原則 |
