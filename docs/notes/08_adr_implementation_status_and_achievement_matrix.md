# 08. 全 ADR 実装達成率および現状ステータスマトリクス

- **ステータス**: 現状整理ノート (Temporary Status Matrix)
- **日付**: 2026-10-05
- **関連 ADR**:
  - [ADR-0028: ADR コンプライアンステスト駆動によるトレーサビリティおよび二重管理防止アーキテクチャ](../../adr/0028-compliance-test-driven-adr-traceability-architecture.md)

---

## 1. 達成率サマリー (2026-10-05 時点)

全 28 件の ADR に対するコード実装状況の総点検結果：

| 区分 | 件数 | 割合 | 備考 |
|---|---|---|---|
| **✅ 完全実装済み (Implemented)** | 21 件 | **77.8%** | 要件がコード化され、検証をパスしている状態 |
| **⚠️ 一部実装 (Partial)** | 3 件 | **11.1%** | バックエンドのみ実装、または特定機能のみ未着手 |
| **❌ 未実装 (Unimplemented)** | 3 件 | **11.1%** | 設計合意済みだがコード未着手（直近起票分含む） |
| **⏸️ 保留・スキップ (YAGNI)** | 1 件 | - | ADR-0009（自由回答パレット採用により4択ダミー保留） |
| **合計（有効 ADR）** | **27 件** | **100.0%** | ※保留を除く有効な意思決定 |

> **現状の実装カバー率**: **88.9%** (完全実装 21 件 + 一部実装 3 件 / 全 27 件)

---

## 2. 全 ADR (0001〜0028) 実装状況マトリクス

| ADR | タイトル | 分類 | 実装状況 | 主な実装コード / 未実装の差分 (Gap) |
|---|---|---|---|---|
| **0001** | TypeID (UUIDv7) サロゲートキー | DOM | ✅ 完了 | `pkg/model/id.go`, 全テーブルの主キー |
| **0002** | バックエンド Go アーキテクチャ | ARC | ✅ 完了 | `cmd/server/main.go`, `pkg/` |
| **0003** | デプロイ環境 (k8s / ghcr.io) | ARC | ✅ 完了 | `Dockerfile`, `deploy/` マニフェスト |
| **0004** | Go 正本スキーマ一元管理 (tygo) | DOM | ✅ 完了 | `scripts/generate-all.sh`, `frontend/src/types/generated.ts` |
| **0005** | SQLite (WAL モード) 採用 | DAT | ✅ 完了 | `pkg/repository/sqlite.go` |
| **0006** | 動的ドメインスキーマ構成 | DOM | ✅ 完了 | `migrations/000001_init.up.sql`, `pkg/model/` |
| **0007** | Local-First / オフライン PWA | APP | ⚠️ **一部実装** | **実装済**: `/api/v1/sync/bootstrap`, ローカル出題・採点<br/>**未実装**: 未表示画像の事前プリフェッチ、IndexedDB 回答蓄積 & Outbox バッチ同期、Service Worker |
| **0008** | 画像アセット永久不変キャッシュ | DAT | ✅ 完了 | `pkg/server/server.go`（RFC 8246 immutable ヘッダー） |
| **0009** | 4択 Strategy パターン | APP | ⏸️ **保留** | 自由回答パレット（ADR-0019）採用に伴い保留（YAGNI） |
| **0010** | Google OIDC 認証とセッション | APP | ❌ **未実装** | User モデルと環境変数の枠組みのみ。ログインハンドラー・Cookie 発行は未着手（現在ゲスト運用） |
| **0011** | ディレクトリ構成と責務境界 | ARC | ✅ 完了 | `cmd/`, `pkg/`, `frontend/`, `deploy/` 構造の順守 |
| **0012** | Kubernetes & ArgoCD GitOps | ARC | ✅ 完了 | `deploy/` 単一 Pod Recreate 構成マニフェスト |
| **0013** | TypeScript AI 駆動ツールチェーン | DEV | ✅ 完了 | Biome, Knip, typos, `scripts/verify-all.sh` |
| **0014** | 最小 Problem Details 規約 | APP | ✅ 完了 | `pkg/model/error.go`（6つの固定エラーコード） |
| **0015** | 最小環境変数 (4つ) と機密性分離 | ARC | ✅ 完了 | `pkg/config/config.go`（厳格な 4 環境変数制限） |
| **0016** | ADR ガバナンスとフラット連番 | DEV | ✅ 完了 | `doc-lifecycle` スキル、連番採番規約 |
| **0017** | リポジトリ IF & Pure Go SQLite | DAT | ✅ 完了 | `pkg/model/repository.go`, `pkg/repository/sqlite.go` |
| **0018** | クイズ母集団フィルタリング分離 | APP | ⚠️ **一部実装** | **実装済**: Go 側ロジック、期生・卒業生フィルタ<br/>**未実装**: フロント側の衣装（PhotoType）選択 UI |
| **0019** | 解答形式 Strategy と自由回答パレット | APP | ✅ 完了 | `PaletteGridInput.tsx`, `DonutRingModal.tsx`, 左右反転許容 |
| **0020** | ブレンドデッキ戦略 (7:3) | APP | ⚠️ **一部実装** | **実装済**: Go 側 `BuildBlendedDeck`<br/>**未実装**: フロント側でのブレンド適用（現在単純シャッフル）、出題ごとの衣装ランダム選出 |
| **0021** | GitOps マスタデータ同期 | DAT | ✅ 完了 | `seeds/seed.sql`, `master_versions` テーブル, ETag 検証 |
| **0022** | プラガブル出題 UI レイアウト | APP | ✅ 完了 | Classic, Compact, Overlay の 3 レイアウト切り替え |
| **0023** | マスタデータ同期整合性機械検証 | DAT | ✅ 完了 | `scripts/build_seed.go`, seed-verify テスト |
| **0024** | 写真メタデータ HITL 策定プロセス | DEV | ✅ 完了 | 人間参加型マスターデータ更新ワークフロー |
| **0025** | embed.FS SQL 完全内包 | ARC | ✅ 完了 | マイグレーション・シードの単一バイナリ内包 |
| **0026** | シリーズ階層分離アーキテクチャ | DOM | ❌ **未実装** | 直前に設計承認。`Series` モデル・DB マイグレーション未着手 |
| **0027** | 楽曲ペンライトカラーデータ構造 | DOM | ❌ **未実装** | 直前に設計承認。`Song` モデル・DB マイグレーション未着手 |
| **0028** | Compliance Test 駆動トレーサビリティ | DEV | ✅ 完了 | 本規約の制定および既存コンプライアンステスト |

---

## 3. 主な未実装（Gap）領域の整理

現在コードベース上に存在する主な未実装領域は、以下の 3 つのクラスタに大別される：

1. **オフライン PWA & 画像キャッシュ完全化（ADR-0007, ADR-0018, ADR-0020）**:
   - ロードマップ作成済み（`docs/notes/06_...`, `docs/notes/07_...`）。
   - 衣装別ランダム出題、衣装フィルタリング UI、起動時画像バックグラウンドプリフェッチ、Service Worker。
2. **シリーズ分離 & 楽曲カラー（ADR-0026, ADR-0027）**:
   - 設計ノート（`docs/notes/06_series_isolation_...`）作成・承認済み。
   - モデル定義（`pkg/model`）、DB マイグレーション、シードデータ拡張が次の着手対象。
3. **Google OIDC 認証 & 回答ログ同期（ADR-0010, ADR-0007 Outbox）**:
   - 現在はゲストプレイに特化しているため実害はないが、ユーザー識別とクラウド同期を行う際に必要。

---

## 4. ADR-0028 への移行方針

本ノートは現時点の棚卸しスナップショット（静的記録）である。
今後の実装作業においては、[ADR-0028](../../adr/0028-compliance-test-driven-adr-traceability-architecture.md) に基づき、**各機能に対応するコンプライアンステスト（`TestADR00XX_...`）を追加・グリーンにすること** をもって、機械的に実装完了を保証・追跡していく。
