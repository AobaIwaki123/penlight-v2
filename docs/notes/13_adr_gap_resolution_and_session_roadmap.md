# 13. ADR ギャップ解消およびセッション分割ロードマップ

- **ステータス**: 完了 (Completed)
- **日付**: 2026-10-07
- **関連 ADR**:
  - [ADR-0007: Local-First / オフライン PWA アーキテクチャ](../../adr/0007-local-first-offline-pwa-architecture.md)
  - [ADR-0009: クイズ出題アルゴリズムにおける Strategy パターン (Superseded)](../../adr/0009-quiz-generation-strategy-pattern.md)
  - [ADR-0010: Google OIDC 認証とセッションセキュリティ (Deferred)](../../adr/0010-google-oidc-authentication-and-session-security.md)
  - [ADR-0019: 解答形式 Strategy と自由回答設計](../../adr/0019-quiz-format-strategy-and-color-palette-architecture.md)
  - [ADR-0028: ADR コンプライアンステスト駆動アーキテクチャ](../../adr/0028-compliance-test-driven-adr-traceability-architecture.md)
  - [ADR-0032: 回答ログにおけるマルチターゲット多態性永続化](../../adr/0032-answer-log-multi-target-polymorphism-architecture.md)

---

## 1. 意思決定とステータス整理方針

1. **ADR-0009（4択 Strategy パターン）の廃止**:
   - [ADR-0019](../../adr/0019-quiz-format-strategy-and-color-palette-architecture.md) において全色パレットからの自由回答入力（Donut / Grid）を本採用したため、ダミー4択選択肢の生成ロジック（CIELAB色差ひっかけ等）は不要となった。
   - ステータスを `Superseded by ADR-0019` に更新し、設計の不要なブレを解消する。
2. **ADR-0010（Google OIDC 認証）の保留**:
   - ゲスト完結での快適なクイズ体験を最優先とするため、Google 認証・クラウド同期機能の実装は保留（`Deferred`）とする。
3. **ADR-0028（コンプライアンステスト）の網羅**:
   - 制定済みの全 ADR に対し、仕様要件を機械的に検証するテスト（`ADR-00XX/Y`）を実装し、動的保証率を 100% に引き上げる。

---

## 2. セッション分割ロードマップ

```mermaid
flowchart TD
    S0["Session 0: ADRステータス是正 & 基礎コンプライアンステスト網羅<br/>(完了・PR #34)"]
    S1["Session 1: オフライン回答ログ同期 & 画像キャッシュ (手動フェッチ)<br/>(完了・PR #35)"]
    S2["Session 2: バックエンド回答集計 API<br/>(完了・PR #37)"]
    S3["Session 3: 履歴・統計 UI & 苦手克服モード<br/>(完了・PR #39)"]

    S0 --> S1 --> S2 --> S3
```



### Session 0: ADRステータス是正 & 基礎コンプライアンステスト網羅（完了・PR #34）
- **ゴール**: 設計ドキュメントのステータス不整合を解消し、未テストの基礎ADRを網羅して CI で機械保証する。
- **対象タスク**:
  - `ADR-0009` を `Superseded by ADR-0019` に更新。
  - `ADR-0010` を `Deferred`（保留）に更新。
  - `adr/README.md` の一覧ステータスを同期。
  - `pkg/quiz/adr_compliance_test.go` に基礎ADR（0001, 0004, 0005, 0006, 0014, 0015, 0017）のコンプライアンステストを追加。

### Session 1: オフライン回答ログ同期 & 画像キャッシュ (手動フェッチ)（完了・PR #35）
- **ゴール**: 圏外でのクイズ完結と復帰時同期を最小オーバーヘッドで実現する（ADR-0007, ADR-0008, ADR-0032）。
- **Part A: 回答ログ Outbox バッチ同期**:
  - クライアント: 解答確定時に IndexedDB（`answer_outbox` ストア）へ回答レコードを保存。
  - バックエンド: `POST /api/v1/quiz/answers/batch` エンドポイントを実装（SQLite の `answer_logs` へ多態性一括永続化、ADR-0032 準拠）。
  - クライアント: `navigator.onLine` や起動時のオンライン検知により、未送信キューを自動バッチ送信。
- **Part B: 画像オフライン保持 & 手動全件フェッチ**:
  - 表示済み画像保持: CacheStorage（`penlight-images-v1`）により、出題・表示された画像をローカルに永続保持（`public/sw.js`）。
  - オフラインモード手動フェッチ: ユーザーが明示的にクリックした時のみ全件（未キャッシュ分のみ差分）をダウンロード。
  - 機能提供トグル: コード内の設定定数（`frontend/src/config/features.ts`）で手軽にフェッチ機能提供の ON/OFF を切り替えられる設計。
  - 拡張性: 将来的なグループ絞り込みフェッチ（`groupId?: string`）を許容するインターフェース設計。

### Session 2: バックエンド回答集計 API（完了・PR #37・Backend）
- **ゴール**: SQLite に蓄積・同期された `answer_logs` から、正答率や苦手対象を抽出・集計する API を提供する。
- **成果物**:
  - モデル拡張: `TargetStat`, `GroupStat`, `QuizStatisticsResponse` を定義。将来の集計項目拡張に備え JSON `metadata` / `extra` フィールドを配備。
  - リポジトリ: `pkg/repository/sqlite.go` に `GetQuizStatistics` を実装（全体サマリー、グループ別集計、メンバー・楽曲多態性ワーストN件抽出）。
  - サーバー: `GET /api/v1/quiz/statistics` エンドポイントを `pkg/server/server.go` に提供（`group_id`, `user_id`, `target_type`, `limit` フィルタリング）。
  - スキーマ同期: `api/openapi.yaml`、TS 型定義、ER 図の同期。
  - テスト: 単体テストおよび ADR-0032/2 コンプライアンステストの網羅。

### Session 3: 履歴・統計 UI & 苦手克服モード（完了・PR #39・Frontend）
- **ゴール**: 蓄積された回答ログを活用し、正答率や苦手カラーの復習モードを直感的な UI で提供する。
- **成果物**:
  - ローカル履歴永続化: IndexedDB（`answer_history` ストア・DB v2）により、同期後もローカルに最新回答ログを保持・集計。
  - 統計モーダル UI: ヘッダー/ポータルに統計ボタン（`IconChartBar`）を配置し、総回答数、全体・グループ別正答率、苦手メンバー/楽曲ランキング（ワースト5件）を可視化。
  - 苦手克服クイズモード: 誤答率の高いメンバー/楽曲を優先出題するデッキ生成（復習モード）。
  - オンライン/オフライン協調: オンライン時は `GET /api/v1/quiz/statistics`、オフライン時は IndexedDB ローカル履歴からフォールバック集計。
  - 人間による UI/UX レビューの実施。



