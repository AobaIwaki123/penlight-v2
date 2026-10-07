---
id: ADR-0034
title: クイズ回答統計集計と苦手克服復習デッキアーキテクチャの採用
status: Accepted
scope: System
primary_category: APP
categories: [APP, DAT, DOM]
tags: [local-first, pwa, quiz-strategy, sqlite]
deciders: [user, ai]
date: 2026-10-07
---

# 0034. クイズ回答統計集計と苦手克服復習デッキアーキテクチャの採用 (0034-quiz-statistics-and-weak-target-review-architecture.md)

- **ステータス**: 承認 (Accepted)
- **日付**: 2026-10-07

---

## 1. 背景と解決すべき課題 (Context & Problem)

[ADR-0007](./0007-local-first-offline-pwa-architecture.md) および [ADR-0032](./0032-answer-log-multi-target-polymorphism-architecture.md) において、オフライン回答ログのキューイング・バッチ同期、およびメンバー（`mem_`）と楽曲（`sng_`）の多態性永続化が整備された。

しかし、蓄積された回答ログを活用してユーザー自身の正答率や苦手カラーを可視化・復習する仕組みについて、以下の課題が存在していた。

1. **集計モデルの結合度と拡張性のトレードオフ**:
   - 将来的に「時系列推移」「連続正解ストリーク」「迷い度スコア（回答速度加重）」などの派生メトリクスが増加する可能性がある。
   - すべてを固定の DB カラムや Go 構造体フィールドとして静的に定義すると、スキーマ変更コストが肥大化し YAGNI 原則（[ADR-0013](./0013-typescript-ai-agent-driven-development-toolchain.md)）に反する。
2. **Local-First とオフライン集計の対称性**:
   - ライブ会場等の圏外環境において、サーバー API（`/api/v1/quiz/statistics`）が通信不能であっても、端末内（IndexedDB）の回答履歴から即時に同一形式の統計を算出・可視化できる必要がある。
3. **誤答フィードバックの即時アクション化**:
   - 単に「誤答率」を提示するだけでなく、苦手な対象（メンバー／楽曲）のみを優先抽出して即座に再挑戦できる復習デッキ生成が求められていた。

---

## 2. 決定事項と具現化仕様 (Decision & Specification)

確固たるコア指標と将来の拡張余地を両立させるため、**「コア 3 階層 + JSON 拡張」モデル**、**Local-First 2 層ストレージ**、および **苦手克服デッキ生成** を正式採用する。

### 1. [Contract] コア 3 階層 + JSON 拡張モデル (`pkg/model/dto.go`)
- コア集計項目（全体サマリー、グループ別成績、苦手ワースト N 件）を型安全に定義し、派生・実験的項目は `metadata` / `extra`（JSON オブジェクト）に逃がす。
  - `QuizStatisticsResponse`: `total_answers`, `total_correct`, `accuracy_rate`, `average_response_time_ms`, `groups`, `weak_targets`, `extra`
  - `TargetStat`: `target_id`, `target_type` (`member` | `song`), `group_id`, `name`, `total_answers`, `correct_answers`, `accuracy_rate`, `average_response_time_ms`, `metadata`
  - `GroupStat`: `group_id`, `group_name`, `total_answers`, `correct_answers`, `accuracy_rate`, `average_response_time_ms`, `metadata`

### 2. [Backend] 多態性集計 API (`pkg/server/server.go`, `pkg/repository/sqlite.go`)
- `GET /api/v1/quiz/statistics`:
  - クエリパラメータ: `user_id` (任意), `group_id` (任意), `target_type` (`member` | `song`, 任意), `limit` (デフォルト 5 件)
  - SQLite `answer_logs` テーブルから全体正答率・平均時間を集計。
  - メンバー（`members`）および楽曲（`songs`）の名前を JOIN し、正答率昇順・回答数降順でワースト N 件を高速算出。

### 3. [Frontend] IndexedDB 2 層ストレージ & 対称集計 (`idb.ts`, `statistics.ts`)
- データベースバージョンを `2` にアップグレードし、一時送信キュー（`answer_outbox`）に加えて永続履歴ストア（`answer_history`）を新設。
- 解答確定時に両方のストアへ同時保存。
- 圏外・オフライン時は、端末内の `answer_history` からバックエンド API と完全に同一ロジック・同一構造でフォールバック集計を実施。

### 4. [Frontend] 統計モーダル UI & 苦手克服デッキ生成 (`StatisticsModal.tsx`, `QuizContainer.tsx`)
- ポータル画面、クイズ画面、および結果発表画面に成績・統計分析ボタン（`IconChartBar`）を配置。
- 全体正答率リングプログレス、累計回答数、平均回答時間、グループ別成績、ワーストランキングを表示。
- 「苦手対象を重点復習する」操作により、ワースト対象を優先選出した復習デッキを自動生成してクイズを開始。

---

## 3. 採否の根拠とトレードオフ (Rationale & Trade-offs)

| 評価軸 | コア3階層 + JSON拡張（採用） | 完全静的スキーマ（代替案A） | 毎回全ログ返却・フロント集計（代替案B） |
|---|---|---|---|
| **将来の拡張性** | **極めて高**（`extra`/`metadata`で後方互換維持） | **低**（カラム追加とDDL変更が必要） | **高**（フロント側で自由に計算） |
| **通信帯域・速度** | **最小**（集計サマリー数KBのみ転送） | **最小** | **大**（回答数増加でMB単位に肥大化） |
| **Local-First親和性** | **高**（共通TS型でクライアント単独計算可能） | **中** | **高** |
| **実装のシンプルさ** | **高**（基本のGROUP BYで完結） | **中** | **低** |

---

## 4. 得られる効果と留意点 (Consequences)

### ポジティブな影響
- **スキーマの安定性**: 時系列や新しいスコアリング指標を追加する際にも、API の破壊的変更や SQLite マイグレーションを回避できる。
- **完全なオフライン耐性 (ADR-0007 整合)**: 会場圏外でも直前のプレイ結果が即時に統計と苦手リストに反映され、復習クイズを楽しめる。
- **学習効率の最大化**: 誤答率の高いメンバー／楽曲をワンタップで再復習できるため、ペンの色当て精度が短時間で向上する。

### 留意点と対策
- **ローカル履歴の容量管理**:
  - IndexedDB の `answer_history` が肥大化しないよう、クライアント側で直近 2,000 件程度を保持する上限管理を将来的に設ける。
