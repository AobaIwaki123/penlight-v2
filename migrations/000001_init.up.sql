-- Penlight Quiz v2 Initial Migration (SQLite WAL)
-- All primary keys use prefixed UUID v7 (TypeID: grp_, col_, mem_, usr_, ans_)

PRAGMA foreign_keys = ON;

-- 1. Idol Groups
CREATE TABLE IF NOT EXISTS groups (
    id TEXT PRIMARY KEY,                       -- grp_<uuidv7>
    name TEXT NOT NULL,                        -- e.g. "日向坂46"
    short_name TEXT NOT NULL,                  -- e.g. "日向坂"
    slug TEXT NOT NULL UNIQUE,                 -- e.g. "hinatazaka46"
    theme_color_hex TEXT NOT NULL,             -- e.g. "#7CC7E8"
    display_order INTEGER NOT NULL DEFAULT 0,
    is_active INTEGER NOT NULL DEFAULT 1,      -- 1: active, 0: inactive
    created_at TEXT NOT NULL,                  -- ISO 8601 UTC
    updated_at TEXT NOT NULL
);

-- 2. Official Penlight Colors
CREATE TABLE IF NOT EXISTS colors (
    id TEXT PRIMARY KEY,                       -- col_<uuidv7>
    group_id TEXT,                             -- NULL for common/standard colors
    name TEXT NOT NULL,                        -- e.g. "スカイブルー"
    hex_code TEXT NOT NULL,                    -- e.g. "#00BFFF"
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_colors_group ON colors(group_id, display_order);

-- 3. Idol Members
CREATE TABLE IF NOT EXISTS members (
    id TEXT PRIMARY KEY,                       -- mem_<uuidv7>
    group_id TEXT NOT NULL,                    -- grp_<uuidv7>
    family_name TEXT NOT NULL,                 -- e.g. "加藤"
    given_name TEXT NOT NULL,                  -- e.g. "史帆"
    family_name_kana TEXT NOT NULL,            -- e.g. "かとう"
    given_name_kana TEXT NOT NULL,             -- e.g. "しほ"
    generation INTEGER NOT NULL,               -- e.g. 1
    status TEXT NOT NULL CHECK(status IN ('active', 'graduated', 'hiatus')),
    left_color_id TEXT NOT NULL,               -- col_<uuidv7>
    right_color_id TEXT NOT NULL,              -- col_<uuidv7>
    ordered INTEGER NOT NULL DEFAULT 0,        -- 0: symmetric, 1: ordered left/right
    joined_at TEXT,
    graduated_at TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY (left_color_id) REFERENCES colors(id),
    FOREIGN KEY (right_color_id) REFERENCES colors(id)
);

CREATE INDEX IF NOT EXISTS idx_members_group_status ON members(group_id, status, generation);

-- 4. Member Images (1:N multiple images per member, Ref: ADR-0021)
CREATE TABLE IF NOT EXISTS member_images (
    id TEXT PRIMARY KEY,                       -- img_<uuidv7>
    member_id TEXT NOT NULL,                   -- mem_<uuidv7>
    image_key TEXT NOT NULL UNIQUE,            -- img_<uuidv7>.webp (ADR-0008 immutable cache)
    title TEXT NOT NULL,                       -- e.g. "13th Single 制服", "5thひな誕祭個別タオル"
    is_primary INTEGER NOT NULL DEFAULT 0,     -- 1: primary/default image, 0: variation
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (member_id) REFERENCES members(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_member_images_member ON member_images(member_id, is_primary);

-- 5. Master Data Versioning (Ref: ADR-0021)
CREATE TABLE IF NOT EXISTS master_versions (
    id TEXT PRIMARY KEY,                       -- 'current'
    version TEXT NOT NULL,                    -- Git commit hash or semantic version
    updated_at TEXT NOT NULL
);

-- 6. Users (Google OIDC / Guest)
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,                       -- usr_<uuidv7>
    google_sub TEXT UNIQUE,                    -- Google Subject ID
    email TEXT,
    display_name TEXT NOT NULL,
    avatar_url TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    last_login_at TEXT NOT NULL
);

-- 7. Quiz Answer Logs
CREATE TABLE IF NOT EXISTS answer_logs (
    id TEXT PRIMARY KEY,                       -- ans_<uuidv7> (client-generated for idempotent batch sync)
    user_id TEXT,                              -- usr_<uuidv7>, NULL for anonymous guest
    quiz_question_id TEXT NOT NULL,            -- quiz_<uuidv7>
    target_member_id TEXT NOT NULL,            -- mem_<uuidv7>
    group_id TEXT NOT NULL,                    -- grp_<uuidv7> for aggregation speed
    is_correct INTEGER NOT NULL,               -- 1: correct, 0: wrong
    response_time_ms INTEGER NOT NULL,
    answered_at TEXT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (target_member_id) REFERENCES members(id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_answer_logs_user ON answer_logs(user_id, group_id, is_correct);
CREATE INDEX IF NOT EXISTS idx_answer_logs_target ON answer_logs(target_member_id, is_correct);
