-- Migration 000003: Add verified_at to members table
-- Ref: docs/notes/10_metadata_review_and_annotation_design.md

PRAGMA foreign_keys = ON;

ALTER TABLE members ADD COLUMN verified_at TEXT;
CREATE INDEX IF NOT EXISTS idx_members_verified ON members(verified_at);
