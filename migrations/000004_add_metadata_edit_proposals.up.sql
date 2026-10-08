-- Generated from pkg/model by scripts/gen-proposal-migration.go. DO NOT EDIT.
-- Ref: ADR-0004, ADR-0036, ADR-0037

ALTER TABLE members ADD COLUMN metadata_revision INTEGER NOT NULL DEFAULT 1 CHECK (metadata_revision >= 1);
ALTER TABLE master_versions ADD COLUMN data_revision INTEGER NOT NULL DEFAULT 1 CHECK (data_revision >= 1);

CREATE TABLE metadata_edit_proposals (
    id TEXT PRIMARY KEY,
    member_id TEXT NOT NULL REFERENCES members(id) ON DELETE RESTRICT,
    base_revision INTEGER NOT NULL CHECK (base_revision >= 1),
    changes_json TEXT NOT NULL CHECK (json_valid(changes_json)),
    status TEXT NOT NULL,
    submitted_at TEXT NOT NULL,
    approved_at TEXT,
    rejected_at TEXT,
    proposer_user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
    approver_user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
    rejection_reason TEXT,
    CHECK (status IN ('pending', 'approved', 'rejected')),
    CHECK (
        (status = 'pending' AND approved_at IS NULL AND rejected_at IS NULL AND approver_user_id IS NULL AND rejection_reason IS NULL)
        OR (status = 'approved' AND approved_at IS NOT NULL AND rejected_at IS NULL AND rejection_reason IS NULL)
        OR (status = 'rejected' AND approved_at IS NULL AND rejected_at IS NOT NULL)
    )
);

CREATE INDEX idx_metadata_edit_proposals_status_submitted
    ON metadata_edit_proposals(status, submitted_at, id);
CREATE INDEX idx_metadata_edit_proposals_member_status
    ON metadata_edit_proposals(member_id, status, submitted_at, id);
