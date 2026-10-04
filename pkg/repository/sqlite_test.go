package repository_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aobaiwaki/penlight-v2/pkg/model"
	"github.com/aobaiwaki/penlight-v2/pkg/repository"
)

func setupTestDB(t *testing.T) (*repository.SQLiteRepository, func()) {
	t.Helper()
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	repo, err := repository.NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("failed to create test repo: %v", err)
	}

	// Apply migration schema
	schemaBytes, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000001_init.up.sql"))
	if err != nil {
		t.Fatalf("failed to read migration file: %v", err)
	}

	if _, err := repo.DB().Exec(string(schemaBytes)); err != nil {
		t.Fatalf("failed to execute migration: %v", err)
	}

	cleanup := func() {
		repo.Close()
	}
	return repo, cleanup
}

func TestSQLiteRepository_MasterDataAndAnswerLogs(t *testing.T) {
	ctx := context.Background()
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	// 1. Insert test group
	now := time.Now().UTC().Truncate(time.Second)
	_, err := repo.DB().ExecContext(ctx, `
		INSERT INTO groups (id, name, short_name, slug, theme_color_hex, display_order, is_active, created_at, updated_at)
		VALUES ('grp_01', '日向坂46', '日向坂', 'hinatazaka46', '#7CC7E8', 1, 1, ?, ?);
	`, now.Format(time.RFC3339), now.Format(time.RFC3339))
	if err != nil {
		t.Fatalf("failed to insert group: %v", err)
	}

	groups, err := repo.ListGroups(ctx)
	if err != nil {
		t.Fatalf("ListGroups failed: %v", err)
	}
	if len(groups) != 1 || groups[0].ID != "grp_01" || groups[0].Name != "日向坂46" {
		t.Fatalf("unexpected groups: %+v", groups)
	}

	// 2. Insert colors
	_, err = repo.DB().ExecContext(ctx, `
		INSERT INTO colors (id, group_id, name, hex_code, display_order, created_at, updated_at)
		VALUES 
			('col_01', 'grp_01', 'パステルブルー', '#7CC7E8', 1, ?, ?),
			('col_02', 'grp_01', 'ホワイト', '#FFFFFF', 2, ?, ?);
	`, now.Format(time.RFC3339), now.Format(time.RFC3339), now.Format(time.RFC3339), now.Format(time.RFC3339))
	if err != nil {
		t.Fatalf("failed to insert colors: %v", err)
	}

	colors, err := repo.ListColors(ctx)
	if err != nil {
		t.Fatalf("ListColors failed: %v", err)
	}
	if len(colors) != 2 {
		t.Fatalf("expected 2 colors, got %d", len(colors))
	}

	// 3. Insert members and image
	_, err = repo.DB().ExecContext(ctx, `
		INSERT INTO members (
			id, group_id, family_name, given_name, family_name_kana, given_name_kana,
			generation, status, left_color_id, right_color_id, ordered,
			created_at, updated_at
		) VALUES (
			'mem_01', 'grp_01', '正源司', '陽子', 'しょうげんじ', 'ようこ',
			4, 'active', 'col_01', 'col_02', 0,
			?, ?
		);
		INSERT INTO member_images (
			id, member_id, image_key, title, is_primary, display_order, created_at, updated_at
		) VALUES (
			'img_01', 'mem_01', 'img_01.webp', '13th Single 制服', 1, 0, ?, ?
		);
	`, now.Format(time.RFC3339), now.Format(time.RFC3339), now.Format(time.RFC3339), now.Format(time.RFC3339))
	if err != nil {
		t.Fatalf("failed to insert member and image: %v", err)
	}

	members, err := repo.ListMembers(ctx)
	if err != nil {
		t.Fatalf("ListMembers failed: %v", err)
	}
	if len(members) != 1 || members[0].FamilyName != "正源司" || members[0].Penlight.LeftColorID != "col_01" {
		t.Fatalf("unexpected members: %+v", members)
	}
	if members[0].PrimaryImage() == nil || members[0].PrimaryImage().ImageKey != "img_01.webp" {
		t.Fatalf("expected primary image img_01.webp, got %+v", members[0].PrimaryImage())
	}


	// 4. Insert AnswerLog (idempotent)
	log := model.AnswerLog{
		ID:             "ans_01",
		QuizQuestionID: "quiz_01",
		TargetMemberID: "mem_01",
		GroupID:        "grp_01",
		IsCorrect:      true,
		ResponseTimeMs: 1200,
		AnsweredAt:     now,
	}
	if err := repo.InsertAnswerLog(ctx, log); err != nil {
		t.Fatalf("InsertAnswerLog failed: %v", err)
	}
	// duplicate insertion should be ignored without error
	if err := repo.InsertAnswerLog(ctx, log); err != nil {
		t.Fatalf("duplicate InsertAnswerLog failed: %v", err)
	}

	// 5. Batch Insert
	logs := []model.AnswerLog{
		{
			ID:             "ans_02",
			QuizQuestionID: "quiz_02",
			TargetMemberID: "mem_01",
			GroupID:        "grp_01",
			IsCorrect:      false,
			ResponseTimeMs: 2500,
			AnsweredAt:     now,
		},
	}
	if err := repo.BatchInsertAnswerLogs(ctx, logs); err != nil {
		t.Fatalf("BatchInsertAnswerLogs failed: %v", err)
	}
}

func TestSQLiteRepository_SeedDataImport(t *testing.T) {
	ctx := context.Background()
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	// Apply seed SQL generated from BigQuery
	seedBytes, err := os.ReadFile(filepath.Join("..", "..", "seeds", "seed.sql"))
	if err != nil {
		t.Fatalf("failed to read seed.sql: %v", err)
	}

	if _, err := repo.DB().ExecContext(ctx, string(seedBytes)); err != nil {
		t.Fatalf("failed to execute seed.sql: %v", err)
	}

	// Verify groups
	groups, err := repo.ListGroups(ctx)
	if err != nil {
		t.Fatalf("ListGroups failed: %v", err)
	}
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}

	// Verify colors
	colors, err := repo.ListColors(ctx)
	if err != nil {
		t.Fatalf("ListColors failed: %v", err)
	}
	if len(colors) != 30 {
		t.Fatalf("expected 30 colors (15 for Hinatazaka, 15 for Sakurazaka), got %d", len(colors))
	}

	// Verify active members (ListMembers filters out graduated members)
	members, err := repo.ListMembers(ctx)
	if err != nil {
		t.Fatalf("ListMembers failed: %v", err)
	}
	if len(members) != 63 {
		t.Fatalf("expected 63 active members, got %d", len(members))
	}

	// Spot check a member
	var shogenji *model.Member
	for i := range members {
		if members[i].FamilyName == "正源司" && members[i].GivenName == "陽子" {
			shogenji = &members[i]
			break
		}
	}
	if shogenji == nil {
		t.Fatal("member 正源司陽子 not found in seed")
	}
	if shogenji.Generation != 4 || shogenji.Status != "active" {
		t.Fatalf("unexpected data for 正源司陽子: %+v", shogenji)
	}
	if shogenji.Penlight.LeftColorID == "" || shogenji.Penlight.RightColorID == "" {
		t.Fatalf("missing colors for 正源司陽子: %+v", shogenji.Penlight)
	}
	if shogenji.PrimaryImage() == nil {
		t.Fatal("expected primary image for 正源司陽子")
	}

	// Verify member images list
	images, err := repo.ListMemberImages(ctx, shogenji.ID)
	if err != nil {
		t.Fatalf("ListMemberImages failed: %v", err)
	}
	if len(images) == 0 {
		t.Fatal("expected at least 1 image for 正源司陽子")
	}

	// Verify master version
	mv, err := repo.GetMasterVersion(ctx)
	if err != nil {
		t.Fatalf("GetMasterVersion failed: %v", err)
	}
	if mv == nil || mv.Version == "" {
		t.Fatalf("expected valid master version, got %+v", mv)
	}
}


