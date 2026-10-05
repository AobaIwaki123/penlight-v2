---
id: ADR-0032
title: 回答ログにおけるマルチターゲット多態性永続化アーキテクチャの採用
status: Accepted
scope: Database
primary_category: DAT
categories: [DAT, DOM]
tags: [sqlite, surrogate-key, single-source-of-truth]
deciders: [user, ai]
date: 2026-10-05
---

# 0032. 回答ログにおけるマルチターゲット多態性永続化アーキテクチャの採用 (0032-answer-log-multi-target-polymorphism-architecture.md)

- **ステータス**: 承認 (Accepted)
- **日付**: 2026-10-05

---

## 1. 背景と解決すべき課題 (Context & Problem)

初期スキーマ（[ADR-0005](./0005-database-selection-sqlite-wal.md), `migrations/000001_init.up.sql`）において、クイズ回答ログテーブル `answer_logs` は以下の定義となっていた。

```sql
CREATE TABLE IF NOT EXISTS answer_logs (
    id TEXT PRIMARY KEY,
    user_id TEXT,
    quiz_question_id TEXT NOT NULL,
    target_member_id TEXT NOT NULL,
    group_id TEXT NOT NULL,
    is_correct INTEGER NOT NULL,
    response_time_ms INTEGER NOT NULL,
    answered_at TEXT NOT NULL,
    FOREIGN KEY (target_member_id) REFERENCES members(id) ON DELETE CASCADE
);
```

[ADR-0027](./0027-song-penlight-color-data-structure.md)（楽曲クイズ）の導入に伴い、出題対象がメンバー（`mem_`）だけでなく楽曲（`sng_`）にも拡張された。
しかし、現行スキーマでは `target_member_id` が `NOT NULL` かつ `members(id)` への外部キー制約を持っているため、楽曲クイズの回答を保存しようとすると SQLite の外部キー制約違反（`FOREIGN KEY constraint failed`）が発生する。

出題対象が多態的（Polymorphic）になった環境において、リレーショナルの参照整合性と将来の拡張性を両立する永続化設計が求められる。

---

## 2. 決定事項と具現化仕様 (Decision & Specification)

SQLite の外部キー制約と連鎖削除（CASCADE）の安全性を最優先し、**NULL 許容 2 カラム + CHECK 排他制約方式** を正式採用する。

### 1. `answer_logs` テーブルのスキーマ改修 (`migrations/000002_add_series_and_songs.up.sql`)

```sql
-- answer_logs の target_member_id を NULL 許容とし、target_song_id を追加
ALTER TABLE answer_logs ADD COLUMN target_song_id TEXT REFERENCES songs(id) ON DELETE CASCADE;

-- ※ SQLite のテーブル再構築または CHECK 制約により排他性を保証:
-- CHECK (
--     (target_member_id IS NOT NULL AND target_song_id IS NULL) OR
--     (target_member_id IS NULL AND target_song_id IS NOT NULL)
-- )
```

### 2. 正本 Go モデルの定義 (`pkg/model/answer.go`)

```go
// AnswerLog records an individual quiz question response by a user for statistics and review.
type AnswerLog struct {
	ID             ID        `json:"id" db:"id,pk"`
	UserID         ID        `json:"user_id" db:"user_id,fk"`
	QuizQuestionID ID        `json:"quiz_question_id" db:"quiz_question_id"`
	TargetMemberID *ID       `json:"target_member_id,omitempty" db:"target_member_id,fk"` // mem_... (メンバー問題時)
	TargetSongID   *ID       `json:"target_song_id,omitempty" db:"target_song_id,fk"`     // sng_... (楽曲問題時)
	GroupID        ID        `json:"group_id" db:"group_id,fk"`
	IsCorrect      bool      `json:"is_correct" db:"is_correct"`
	ResponseTimeMs int       `json:"response_time_ms" db:"response_time_ms"`
	AnsweredAt     time.Time `json:"answered_at" db:"answered_at"`
}

// GetTargetID returns whichever target ID is populated.
func (a *AnswerLog) GetTargetID() ID {
	if a.TargetMemberID != nil {
		return *a.TargetMemberID
	}
	if a.TargetSongID != nil {
		return *a.TargetSongID
	}
	return ""
}
```

---

## 3. 採否の根拠とトレードオフ (Rationale & Trade-offs)

### 永続化アプローチの比較検討

| 評価軸 | 採用案（NULL許容2カラム＋CHECK制約） | 代替案（単一 target_id＋FK制約解除） |
|---|---|---|
| **参照整合性** | **完全**（SQLite FK が各テーブルに有効） | **不完全**（外部キー制約が貼れない） |
| **連鎖削除 (CASCADE)** | **安全**（メンバー/楽曲削除時に自動削除） | **手動**（アプリ側で手動 DELETE が必要） |
| **集計クエリの容易さ** | **高**（JOIN が直接結合可能） | **中**（条件分岐 JOIN が必要） |
| **スキーマの簡潔さ** | カラムが 1 つ増える | カラム 1 つで済む |

---

## 4. 得られる効果と留意点 (Consequences)

### ポジティブな影響
- **強固なデータ完全性**: 誤って存在しない楽曲 ID やメンバー ID がログに記録されることを DB レベルで物理遮断できる。
- **未出題/復習アルゴリズムとの整合 ([ADR-0031](./0031-generic-quiz-engine-and-target-abstraction.md))**:
  - `AnswerLog.GetTargetID()` により、ジェネリック出題エンジン `BuildBlendedDeck[T QuizTarget]` が対象種別を意識せず回答履歴と照合できる。

### 留意点と対策
- **インデックスの最適化**:
  - `idx_answer_logs_target_member` (`target_member_id, is_correct`) に加え、`idx_answer_logs_target_song` (`target_song_id, is_correct`) を作成し、楽曲クイズの履歴集計速度を維持する。
