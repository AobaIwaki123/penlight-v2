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

	// Apply migration schemas
	for _, migrationFile := range []string{"000001_init.up.sql", "000002_add_series_and_songs.up.sql", "000003_add_member_verified_at.up.sql", "000004_add_metadata_edit_proposals.up.sql"} {
		schemaBytes, err := os.ReadFile(filepath.Join("..", "..", "migrations", migrationFile))
		if err != nil {
			t.Fatalf("failed to read migration file %s: %v", migrationFile, err)
		}
		if _, err := repo.DB().Exec(string(schemaBytes)); err != nil {
			t.Fatalf("failed to execute migration %s: %v", migrationFile, err)
		}
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

	// 3. Insert photo types, members and image
	_, err = repo.DB().ExecContext(ctx, `
		INSERT INTO photo_types (id, group_id, name, slug, display_order, created_at, updated_at)
		VALUES ('pht_01', 'grp_01', '13th制服', '13th-uniform', 1, ?, ?);
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
			id, member_id, photo_type_id, image_key, is_primary, display_order, created_at, updated_at
		) VALUES (
			'img_01', 'mem_01', 'pht_01', 'img_01.webp', 1, 0, ?, ?
		);
	`, now.Format(time.RFC3339), now.Format(time.RFC3339), now.Format(time.RFC3339), now.Format(time.RFC3339), now.Format(time.RFC3339), now.Format(time.RFC3339))
	if err != nil {
		t.Fatalf("failed to insert member and image: %v", err)
	}

	photoTypes, err := repo.ListPhotoTypes(ctx, "grp_01")
	if err != nil {
		t.Fatalf("ListPhotoTypes failed: %v", err)
	}
	if len(photoTypes) != 1 || photoTypes[0].ID != "pht_01" || photoTypes[0].Slug != "13th-uniform" {
		t.Fatalf("unexpected photoTypes: %+v", photoTypes)
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
	if members[0].PrimaryImage().PhotoType == nil || members[0].PrimaryImage().PhotoType.Name != "13th制服" {
		t.Fatalf("expected photo type for primary image, got %+v", members[0].PrimaryImage().PhotoType)
	}

	// 4. Insert AnswerLog (idempotent, polymorphic Ref: ADR-0032)
	targetMemID := model.ID("mem_01")
	log := model.AnswerLog{
		ID:             "ans_01",
		QuizQuestionID: "quiz_01",
		TargetMemberID: &targetMemID,
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

	// 5. Batch Insert with Member and Song targets
	// Insert test song first for FK
	_, err = repo.DB().ExecContext(ctx, `
		INSERT INTO songs (id, group_id, title, color1_id, created_at, updated_at)
		VALUES ('sng_01', 'grp_01', 'テスト楽曲', 'col_01', '2026-10-04T00:00:00Z', '2026-10-04T00:00:00Z');
	`)
	if err != nil {
		t.Fatalf("failed to insert test song: %v", err)
	}

	targetSongID := model.ID("sng_01")
	logs := []model.AnswerLog{
		{
			ID:             "ans_02",
			QuizQuestionID: "quiz_02",
			TargetMemberID: &targetMemID,
			GroupID:        "grp_01",
			IsCorrect:      false,
			ResponseTimeMs: 2500,
			AnsweredAt:     now,
		},
		{
			ID:             "ans_03",
			QuizQuestionID: "quiz_03",
			TargetSongID:   &targetSongID,
			GroupID:        "grp_01",
			IsCorrect:      true,
			ResponseTimeMs: 1800,
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

	// Verify series
	seriesList, err := repo.ListSeries(ctx)
	if err != nil {
		t.Fatalf("ListSeries failed: %v", err)
	}
	if len(seriesList) != 2 {
		t.Fatalf("expected 2 series (Sakamichi, Ikolove), got %d", len(seriesList))
	}

	// Verify groups
	groups, err := repo.ListGroups(ctx)
	if err != nil {
		t.Fatalf("ListGroups failed: %v", err)
	}
	if len(groups) != 6 {
		t.Fatalf("expected 6 groups (3 Sakamichi, 3 Ikolove), got %d", len(groups))
	}

	// Verify colors
	colors, err := repo.ListColors(ctx)
	if err != nil {
		t.Fatalf("ListColors failed: %v", err)
	}
	if len(colors) != 82 {
		t.Fatalf("expected 82 colors (41 Sakamichi, 41 Ikolove), got %d", len(colors))
	}

	// Verify active members (ListMembers filters out graduated members)
	members, err := repo.ListMembers(ctx)
	if err != nil {
		t.Fatalf("ListMembers failed: %v", err)
	}
	if len(members) != 118 {
		t.Fatalf("expected 118 active members (85 Sakamichi + 33 Ikolove), got %d", len(members))
	}

	// Verify songs
	songs, err := repo.ListSongs(ctx)
	if err != nil {
		t.Fatalf("ListSongs failed: %v", err)
	}
	if len(songs) != 9 {
		t.Fatalf("expected 9 songs, got %d", len(songs))
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

	// Verify photo types
	photoTypes, err := repo.ListPhotoTypes(ctx, "")
	if err != nil {
		t.Fatalf("ListPhotoTypes failed: %v", err)
	}
	if len(photoTypes) != 11 {
		t.Fatalf("expected 11 photo types (8 Sakamichi, 3 Ikolove), got %d", len(photoTypes))
	}

	// Verify member images list
	images, err := repo.ListMemberImages(ctx, shogenji.ID)
	if err != nil {
		t.Fatalf("ListMemberImages failed: %v", err)
	}
	if len(images) == 0 {
		t.Fatal("expected at least 1 image for 正源司陽子")
	}
	if images[0].PhotoType == nil || images[0].PhotoType.Name == "" {
		t.Fatalf("expected loaded PhotoType on image, got %+v", images[0].PhotoType)
	}

	// Spot check an equal_love member's image
	var maika *model.Member
	for i := range members {
		if members[i].FamilyName == "佐々木" && members[i].GivenName == "舞香" {
			maika = &members[i]
			break
		}
	}
	if maika == nil {
		t.Fatal("member 佐々木舞香 not found in seed")
	}
	if maika.PrimaryImage() == nil {
		t.Fatal("expected primary image for 佐々木舞香")
	}
	maikaImages, err := repo.ListMemberImages(ctx, maika.ID)
	if err != nil || len(maikaImages) == 0 {
		t.Fatalf("expected images for 佐々木舞香, got err: %v, count: %d", err, len(maikaImages))
	}
	if maikaImages[0].PhotoType == nil || maikaImages[0].PhotoType.Slug != "21th_single" {
		t.Fatalf("expected 21th_single photo type for 佐々木舞香, got %+v", maikaImages[0].PhotoType)
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

func TestSQLiteRepository_SeriesAndSongs(t *testing.T) {
	ctx := context.Background()
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	now := time.Now().UTC().Truncate(time.Second).Format(time.RFC3339)

	// 1. Insert series
	_, err := repo.DB().ExecContext(ctx, `
		INSERT INTO series (id, name, slug, display_order, created_at, updated_at)
		VALUES
			('ser_01', '坂道シリーズ', 'sakamichi', 1, ?, ?),
			('ser_02', 'イコノイジョイ', 'ikolove', 2, ?, ?);
	`, now, now, now, now)
	if err != nil {
		t.Fatalf("failed to insert series: %v", err)
	}

	// 2. Test ListSeries and GetSeries
	seriesList, err := repo.ListSeries(ctx)
	if err != nil {
		t.Fatalf("ListSeries failed: %v", err)
	}
	if len(seriesList) != 2 || seriesList[0].Slug != "sakamichi" || seriesList[1].Slug != "ikolove" {
		t.Fatalf("unexpected series list: %+v", seriesList)
	}

	ser, err := repo.GetSeries(ctx, "ser_01")
	if err != nil {
		t.Fatalf("GetSeries failed: %v", err)
	}
	if ser == nil || ser.Name != "坂道シリーズ" {
		t.Fatalf("unexpected series: %+v", ser)
	}

	// 3. Insert groups with series_id
	_, err = repo.DB().ExecContext(ctx, `
		INSERT INTO groups (id, series_id, name, short_name, slug, theme_color_hex, display_order, is_active, created_at, updated_at)
		VALUES
			('grp_hinata', 'ser_01', '日向坂46', '日向坂', 'hinatazaka46', '#7CC7E8', 1, 1, ?, ?),
			('grp_equal', 'ser_02', '=LOVE', 'イコラブ', 'equal-love', '#FFC0CB', 2, 1, ?, ?);
	`, now, now, now, now)
	if err != nil {
		t.Fatalf("failed to insert groups: %v", err)
	}

	// 4. Test ListGroups and ListGroupsBySeries
	groups, err := repo.ListGroups(ctx)
	if err != nil {
		t.Fatalf("ListGroups failed: %v", err)
	}
	if len(groups) != 2 || groups[0].SeriesID != "ser_01" || groups[1].SeriesID != "ser_02" {
		t.Fatalf("unexpected groups with series_id: %+v", groups)
	}

	sakamichiGroups, err := repo.ListGroupsBySeries(ctx, "ser_01")
	if err != nil {
		t.Fatalf("ListGroupsBySeries failed: %v", err)
	}
	if len(sakamichiGroups) != 1 || sakamichiGroups[0].ID != "grp_hinata" {
		t.Fatalf("unexpected sakamichi groups: %+v", sakamichiGroups)
	}

	// 5. Insert colors
	_, err = repo.DB().ExecContext(ctx, `
		INSERT INTO colors (id, group_id, name, hex_code, display_order, created_at, updated_at)
		VALUES
			('col_sky', 'grp_hinata', 'スカイブルー', '#7CC7E8', 1, ?, ?),
			('col_white', 'grp_hinata', 'ホワイト', '#FFFFFF', 2, ?, ?),
			('col_pink', 'grp_equal', 'ピンク', '#FFC0CB', 1, ?, ?);
	`, now, now, now, now, now, now)
	if err != nil {
		t.Fatalf("failed to insert colors: %v", err)
	}

	// 6. Insert members and test ListMembersBySeries
	_, err = repo.DB().ExecContext(ctx, `
		INSERT INTO members (id, group_id, family_name, given_name, family_name_kana, given_name_kana, generation, status, left_color_id, right_color_id, ordered, created_at, updated_at)
		VALUES
			('mem_01', 'grp_hinata', '正源司', '陽子', 'しょうげんじ', 'ようこ', 4, 'active', 'col_sky', 'col_white', 0, ?, ?),
			('mem_02', 'grp_equal', '佐々木', '舞香', 'ささき', 'まいか', 1, 'active', 'col_pink', 'col_pink', 0, ?, ?);
	`, now, now, now, now)
	if err != nil {
		t.Fatalf("failed to insert members: %v", err)
	}

	sakamichiMembers, err := repo.ListMembersBySeries(ctx, "ser_01")
	if err != nil {
		t.Fatalf("ListMembersBySeries failed: %v", err)
	}
	if len(sakamichiMembers) != 1 || sakamichiMembers[0].ID != "mem_01" {
		t.Fatalf("unexpected sakamichi members: %+v", sakamichiMembers)
	}

	// 7. Insert songs and test ListSongs, ListSongsByGroup, ListSongsBySeries
	_, err = repo.DB().ExecContext(ctx, `
		INSERT INTO songs (id, group_id, title, kana, color1_id, color2_id, created_at, updated_at)
		VALUES
			('sng_01', 'grp_hinata', 'キュン', 'きゅん', 'col_sky', 'col_white', ?, ?),
			('sng_02', 'grp_equal', '絶対アイドル辞めないで', 'ぜったいあいどるやめないで', 'col_pink', NULL, ?, ?);
	`, now, now, now, now)
	if err != nil {
		t.Fatalf("failed to insert songs: %v", err)
	}

	allSongs, err := repo.ListSongs(ctx)
	if err != nil {
		t.Fatalf("ListSongs failed: %v", err)
	}
	if len(allSongs) != 2 {
		t.Fatalf("expected 2 songs, got %d", len(allSongs))
	}

	hinataSongs, err := repo.ListSongsByGroup(ctx, "grp_hinata")
	if err != nil {
		t.Fatalf("ListSongsByGroup failed: %v", err)
	}
	if len(hinataSongs) != 1 || hinataSongs[0].Title != "キュン" || hinataSongs[0].Color2ID == nil {
		t.Fatalf("unexpected hinata songs: %+v", hinataSongs)
	}

	equalSongs, err := repo.ListSongsBySeries(ctx, "ser_02")
	if err != nil {
		t.Fatalf("ListSongsBySeries failed: %v", err)
	}
	if len(equalSongs) != 1 || equalSongs[0].Title != "絶対アイドル辞めないで" || equalSongs[0].Color2ID != nil {
		t.Fatalf("unexpected equal songs: %+v", equalSongs)
	}

	// 8. Test GetQuizStatistics
	mem01 := model.ID("mem_01")
	song01 := model.ID("sng_01")
	batchLogs := []model.AnswerLog{
		{
			ID:             "ans_101",
			QuizQuestionID: "quiz_101",
			TargetMemberID: &mem01,
			GroupID:        "grp_hinata",
			IsCorrect:      false,
			ResponseTimeMs: 3000,
			AnsweredAt:     time.Now().UTC(),
		},
		{
			ID:             "ans_102",
			QuizQuestionID: "quiz_102",
			TargetMemberID: &mem01,
			GroupID:        "grp_hinata",
			IsCorrect:      true,
			ResponseTimeMs: 1500,
			AnsweredAt:     time.Now().UTC(),
		},
		{
			ID:             "ans_103",
			QuizQuestionID: "quiz_103",
			TargetSongID:   &song01,
			GroupID:        "grp_hinata",
			IsCorrect:      true,
			ResponseTimeMs: 1200,
			AnsweredAt:     time.Now().UTC(),
		},
	}
	if err := repo.BatchInsertAnswerLogs(ctx, batchLogs); err != nil {
		t.Fatalf("BatchInsertAnswerLogs failed: %v", err)
	}

	stats, err := repo.GetQuizStatistics(ctx, model.QuizStatisticsFilter{Limit: 5})
	if err != nil {
		t.Fatalf("GetQuizStatistics failed: %v", err)
	}
	if stats.TotalAnswers != 3 {
		t.Fatalf("expected TotalAnswers=3, got %d", stats.TotalAnswers)
	}
	if stats.TotalCorrect != 2 {
		t.Fatalf("expected TotalCorrect=2, got %d", stats.TotalCorrect)
	}
	if len(stats.Groups) != 1 || stats.Groups[0].GroupID != "grp_hinata" {
		t.Fatalf("unexpected group stats: %+v", stats.Groups)
	}
	if len(stats.WeakTargets) < 2 {
		t.Fatalf("expected at least 2 weak targets, got %d", len(stats.WeakTargets))
	}
	// Weakest should be member (50% accuracy) followed by song (100% accuracy)
	if stats.WeakTargets[0].TargetID != "mem_01" || stats.WeakTargets[0].AccuracyRate != 0.5 {
		t.Fatalf("unexpected weakest target: %+v", stats.WeakTargets[0])
	}
}
