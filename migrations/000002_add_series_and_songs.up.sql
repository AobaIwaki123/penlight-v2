-- Migration 000002: Add series and songs hierarchy, polymorphic answer logs
-- Ref: ADR-0026, ADR-0027, ADR-0032

PRAGMA foreign_keys = ON;

-- 1. Idol Series / Franchises (Ref: ADR-0026)
CREATE TABLE IF NOT EXISTS series (
    id TEXT PRIMARY KEY,                       -- ser_<uuidv7>
    name TEXT NOT NULL,                        -- e.g. "坂道シリーズ", "イコノイジョイ"
    slug TEXT NOT NULL UNIQUE,                 -- e.g. "sakamichi", "ikolove"
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,                  -- ISO 8601 UTC
    updated_at TEXT NOT NULL
);

-- 2. Add series_id to groups (Ref: ADR-0026)
ALTER TABLE groups ADD COLUMN series_id TEXT REFERENCES series(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_groups_series ON groups(series_id, display_order);

-- 3. Songs and Penlight Colors (Ref: ADR-0027)
CREATE TABLE IF NOT EXISTS songs (
    id TEXT PRIMARY KEY,                       -- sng_<uuidv7>
    group_id TEXT NOT NULL,                    -- grp_<uuidv7>
    title TEXT NOT NULL,                       -- e.g. "絶対アイドル辞めないで"
    kana TEXT,                                 -- reading for sorting (optional)
    color1_id TEXT NOT NULL,                   -- col_<uuidv7> (primary color)
    color2_id TEXT,                            -- col_<uuidv7> (secondary color, optional)
    created_at TEXT NOT NULL,                  -- ISO 8601 UTC
    updated_at TEXT NOT NULL,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY (color1_id) REFERENCES colors(id) ON DELETE RESTRICT,
    FOREIGN KEY (color2_id) REFERENCES colors(id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_songs_group ON songs(group_id);

-- 4. Polymorphic Answer Logs (Ref: ADR-0032)
CREATE TABLE IF NOT EXISTS answer_logs_new (
    id TEXT PRIMARY KEY,                       -- ans_<uuidv7>
    user_id TEXT,                              -- usr_<uuidv7>, NULL for anonymous guest
    quiz_question_id TEXT NOT NULL,            -- quiz_<uuidv7>
    target_member_id TEXT,                     -- mem_<uuidv7>, populated for member quizzes
    target_song_id TEXT,                       -- sng_<uuidv7>, populated for song quizzes
    group_id TEXT NOT NULL,                    -- grp_<uuidv7> for aggregation speed
    is_correct INTEGER NOT NULL,               -- 1: correct, 0: wrong
    response_time_ms INTEGER NOT NULL,
    answered_at TEXT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (target_member_id) REFERENCES members(id) ON DELETE CASCADE,
    FOREIGN KEY (target_song_id) REFERENCES songs(id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    CHECK (
        (target_member_id IS NOT NULL AND target_song_id IS NULL) OR
        (target_member_id IS NULL AND target_song_id IS NOT NULL)
    )
);

INSERT INTO answer_logs_new (
    id, user_id, quiz_question_id, target_member_id, target_song_id, group_id, is_correct, response_time_ms, answered_at
)
SELECT
    id, user_id, quiz_question_id, target_member_id, NULL, group_id, is_correct, response_time_ms, answered_at
FROM answer_logs;

DROP TABLE answer_logs;
ALTER TABLE answer_logs_new RENAME TO answer_logs;

CREATE INDEX IF NOT EXISTS idx_answer_logs_user ON answer_logs(user_id, group_id, is_correct);
CREATE INDEX IF NOT EXISTS idx_answer_logs_target ON answer_logs(target_member_id, is_correct);
CREATE INDEX IF NOT EXISTS idx_answer_logs_target_song ON answer_logs(target_song_id, is_correct);
