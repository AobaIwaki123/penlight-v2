package quiz_test

import (
	"context"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aobaiwaki/penlight-v2/pkg/model"
	"github.com/aobaiwaki/penlight-v2/pkg/quiz"
	"github.com/aobaiwaki/penlight-v2/pkg/repository"
)

// setupRealSeedDB loads the real seed database generated from seeds/seed.sql.
func setupRealSeedDB(t *testing.T) (*repository.SQLiteRepository, func()) {
	t.Helper()
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	repo, err := repository.NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("failed to create sqlite repo: %v", err)
	}

	for _, migrationFile := range []string{"000001_init.up.sql", "000002_add_series_and_songs.up.sql"} {
		migrationBytes, err := os.ReadFile(filepath.Join("..", "..", "migrations", migrationFile))
		if err != nil {
			t.Fatalf("failed to read migration %s: %v", migrationFile, err)
		}
		if _, err := repo.DB().Exec(string(migrationBytes)); err != nil {
			t.Fatalf("failed to apply migration %s: %v", migrationFile, err)
		}
	}

	seedBytes, err := os.ReadFile(filepath.Join("..", "..", "seeds", "seed.sql"))
	if err != nil {
		t.Fatalf("failed to read seed.sql: %v", err)
	}
	if _, err := repo.DB().Exec(string(seedBytes)); err != nil {
		t.Fatalf("failed to apply seed.sql: %v", err)
	}

	return repo, func() { repo.Close() }
}

// =========================================================================
// ADR-0018: Candidate Pool Filtering Specification Compliance Tests
// =========================================================================

func TestADR0018_Compliance_FilterRules(t *testing.T) {
	ctx := context.Background()
	repo, cleanup := setupRealSeedDB(t)
	defer cleanup()

	members, err := repo.ListMembers(ctx)
	if err != nil {
		t.Fatalf("failed to list real members: %v", err)
	}
	groups, err := repo.ListGroups(ctx)
	if err != nil {
		t.Fatalf("failed to list groups: %v", err)
	}

	var hinataID, sakuraID, nogiID model.ID
	for _, g := range groups {
		if g.Slug == "hinatazaka46" {
			hinataID = g.ID
		} else if g.Slug == "sakurazaka46" {
			sakuraID = g.ID
		} else if g.Slug == "nogizaka46" {
			nogiID = g.ID
		}
	}

	t.Run("ADR-0018: Group Filtering on Real Seed Data", func(t *testing.T) {
		filteredHinata, err := quiz.FilterMembers(members, model.QuizFilter{GroupID: &hinataID})
		if err != nil {
			t.Fatalf("FilterMembers hinata failed: %v", err)
		}
		filteredSakura, err := quiz.FilterMembers(members, model.QuizFilter{GroupID: &sakuraID})
		if err != nil {
			t.Fatalf("FilterMembers sakura failed: %v", err)
		}
		filteredNogi, err := quiz.FilterMembers(members, model.QuizFilter{GroupID: &nogiID})
		if err != nil {
			t.Fatalf("FilterMembers nogi failed: %v", err)
		}

		if len(filteredHinata) == 0 || len(filteredSakura) == 0 || len(filteredNogi) == 0 {
			t.Fatal("expected members for all three groups")
		}
		if len(filteredHinata)+len(filteredSakura)+len(filteredNogi) != len(members) {
			t.Fatalf("sum of groups (%d+%d+%d) does not match total active members (%d)",
				len(filteredHinata), len(filteredSakura), len(filteredNogi), len(members))
		}

		for _, m := range filteredHinata {
			if m.GroupID != hinataID {
				t.Fatalf("expected group %s, got %s for member %s", hinataID, m.GroupID, m.ID)
			}
		}
		for _, m := range filteredNogi {
			if m.GroupID != nogiID {
				t.Fatalf("expected group %s, got %s for member %s", nogiID, m.GroupID, m.ID)
			}
		}
	})

	t.Run("ADR-0018: Generation Filtering on Real Seed Data (Hinata 4th Gen: exactly 11)", func(t *testing.T) {
		gen4Filter := model.QuizFilter{
			GroupID:     &hinataID,
			Generations: []int{4},
		}
		gen4Members, err := quiz.FilterMembers(members, gen4Filter)
		if err != nil {
			t.Fatalf("FilterMembers gen4 failed: %v", err)
		}

		// Hinatazaka 4th generation has 11 active members in seed data
		if len(gen4Members) != 11 {
			t.Fatalf("expected 11 Hinatazaka 4th gen members, got %d", len(gen4Members))
		}
		for _, m := range gen4Members {
			if m.Generation != 4 || m.GroupID != hinataID {
				t.Fatalf("invalid member in gen 4 result: %+v", m)
			}
		}
	})

	t.Run("ADR-0018: Multi-Generation Filtering ([3, 4])", func(t *testing.T) {
		multiGenFilter := model.QuizFilter{
			GroupID:     &hinataID,
			Generations: []int{3, 4},
		}
		res, err := quiz.FilterMembers(members, multiGenFilter)
		if err != nil {
			t.Fatalf("FilterMembers multi-gen failed: %v", err)
		}

		for _, m := range res {
			if m.Generation != 3 && m.Generation != 4 {
				t.Fatalf("unexpected generation %d in result", m.Generation)
			}
		}
	})

	t.Run("ADR-0018: Costume (PhotoType) Filtering on Real Seed Data", func(t *testing.T) {
		photoTypes, err := repo.ListPhotoTypes(ctx, hinataID)
		if err != nil {
			t.Fatalf("ListPhotoTypes failed: %v", err)
		}
		if len(photoTypes) == 0 {
			t.Fatal("expected photo types for Hinatazaka")
		}

		targetPhotoType := photoTypes[0]
		costumeFilter := model.QuizFilter{
			GroupID:      &hinataID,
			PhotoTypeIDs: []model.ID{targetPhotoType.ID},
		}
		res, err := quiz.FilterMembers(members, costumeFilter)
		if err != nil {
			t.Fatalf("costume filter failed: %v", err)
		}

		if len(res) == 0 {
			t.Fatalf("expected members for costume %s", targetPhotoType.Name)
		}

		for _, m := range res {
			hasCostume := false
			for _, img := range m.Images {
				if img.PhotoTypeID == targetPhotoType.ID {
					hasCostume = true
					break
				}
			}
			if !hasCostume {
				t.Fatalf("member %s does not have costume %s", m.ID, targetPhotoType.ID)
			}
		}
	})

	t.Run("ADR-0018 & ADR-0014: Minimum Candidate Validation (HTTP 400 & CodeInsufficientMembers)", func(t *testing.T) {
		// Non-existent generation to trigger insufficient candidate error
		emptyFilter := model.QuizFilter{
			GroupID:     &hinataID,
			Generations: []int{99},
		}
		res, err := quiz.FilterMembers(members, emptyFilter)
		if err == nil {
			t.Fatalf("expected insufficient candidate error, got %d members", len(res))
		}

		appErr, ok := err.(model.AppError)
		if !ok {
			t.Fatalf("expected model.AppError, got %T: %v", err, err)
		}
		if appErr.Status != http.StatusBadRequest {
			t.Fatalf("expected HTTP 400, got %d", appErr.Status)
		}
		if appErr.Code != model.CodeInsufficientMembers {
			t.Fatalf("expected Code %s, got %s", model.CodeInsufficientMembers, appErr.Code)
		}
	})
}

// =========================================================================
// ADR-0019: Free-Answer Palette & Inverted Orientation Judgment Tests
// =========================================================================

func TestADR0019_Compliance_JudgmentRules(t *testing.T) {
	ctx := context.Background()
	repo, cleanup := setupRealSeedDB(t)
	defer cleanup()

	members, err := repo.ListMembers(ctx)
	if err != nil {
		t.Fatalf("failed to list real members: %v", err)
	}

	var shogenji *model.Member
	for i := range members {
		if members[i].FamilyName == "正源司" && members[i].GivenName == "陽子" {
			shogenji = &members[i]
			break
		}
	}
	if shogenji == nil {
		t.Fatal("正源司陽子 not found in real seed data")
	}

	correctL := shogenji.Penlight.LeftColorID
	correctR := shogenji.Penlight.RightColorID
	dummyColor := model.ID("col_dummy_wrong_color")

	t.Run("ADR-0019: Direct order match (Left, Right) is TRUE", func(t *testing.T) {
		if !quiz.JudgeAnswer(correctL, correctR, *shogenji) {
			t.Fatal("expected true for direct match")
		}
	})

	t.Run("ADR-0019: Inverted order match (Right, Left) is TRUE (Hand switch tolerated)", func(t *testing.T) {
		if !quiz.JudgeAnswer(correctR, correctL, *shogenji) {
			t.Fatal("expected true for inverted match")
		}
	})

	t.Run("ADR-0019: One color wrong is FALSE", func(t *testing.T) {
		if quiz.JudgeAnswer(correctL, dummyColor, *shogenji) {
			t.Fatal("expected false when one color is wrong")
		}
		if quiz.JudgeAnswer(dummyColor, correctR, *shogenji) {
			t.Fatal("expected false when one color is wrong")
		}
	})

	t.Run("ADR-0019: Both colors wrong is FALSE", func(t *testing.T) {
		if quiz.JudgeAnswer(dummyColor, dummyColor, *shogenji) {
			t.Fatal("expected false when both colors are wrong")
		}
	})

	t.Run("ADR-0019: Double tap on identical colors (Same-Color Member)", func(t *testing.T) {
		sameColorMember := model.Member{
			ID: "mem_same",
			Penlight: model.PenlightPair{
				LeftColorID:  model.ID("col_white"),
				RightColorID: model.ID("col_white"),
			},
		}

		// Double tap with correct color -> TRUE
		if !quiz.JudgeAnswer("col_white", "col_white", sameColorMember) {
			t.Fatal("expected true for double tap match")
		}

		// Double tap with wrong color -> FALSE
		if quiz.JudgeAnswer("col_red", "col_red", sameColorMember) {
			t.Fatal("expected false for double tap wrong color")
		}

		// One white, one blue -> FALSE
		if quiz.JudgeAnswer("col_white", "col_blue", sameColorMember) {
			t.Fatal("expected false for partial match on same-color member")
		}
	})
}

// =========================================================================
// ADR-0020: Blended Deck Strategy Specification Compliance Tests
// =========================================================================

func TestADR0020_Compliance_BlendedDeckStrategy(t *testing.T) {
	rng := rand.New(rand.NewSource(12345))
	pool := generateMembers(25) // 25 members

	t.Run("ADR-0020 Property 1: Zero duplicates within a deck", func(t *testing.T) {
		deck := quiz.BuildBlendedDeck(pool, nil, 10, rng)
		if len(deck) != 10 {
			t.Fatalf("expected deck of 10, got %d", len(deck))
		}

		seen := make(map[model.ID]bool)
		for _, m := range deck {
			if seen[m.ID] {
				t.Fatalf("duplicate member %s in deck", m.ID)
			}
			seen[m.ID] = true
		}
	})

	t.Run("ADR-0020 Property 2: 7:3 Target Allocation when both unseen and seen exist", func(t *testing.T) {
		// First 10 members are seen in history
		history := make([]model.AnswerLog, 10)
		for i := 0; i < 10; i++ {
			id := pool[i].ID
			history[i] = model.AnswerLog{TargetMemberID: &id}
		}

		deck := quiz.BuildBlendedDeck(pool, history, 10, rng)
		if len(deck) != 10 {
			t.Fatalf("expected 10 members, got %d", len(deck))
		}

		seenCount := 0
		unseenCount := 0
		seenMap := make(map[model.ID]bool)
		for _, log := range history {
			seenMap[log.GetTargetID()] = true
		}

		for _, m := range deck {
			if seenMap[m.ID] {
				seenCount++
			} else {
				unseenCount++
			}
		}

		if unseenCount != 7 || seenCount != 3 {
			t.Fatalf("expected 7 unseen and 3 seen, got %d unseen and %d seen", unseenCount, seenCount)
		}
	})

	t.Run("ADR-0020 Property 3: Unseen Zeroing Simulation (Solving Coupon Collector Problem)", func(t *testing.T) {
		// Simulate playing multiple sessions until all 25 members have been seen
		simRng := rand.New(rand.NewSource(42))
		var accumulatedHistory []model.AnswerLog

		totalSeenSet := make(map[model.ID]bool)
		maxSessions := 6 // In 25 members with 7 unseen per session, should take ceil(25/7) = 4 sessions

		for session := 1; session <= maxSessions; session++ {
			deck := quiz.BuildBlendedDeck(pool, accumulatedHistory, 10, simRng)
			if len(deck) != 10 {
				t.Fatalf("session %d: expected deck size 10, got %d", session, len(deck))
			}

			// Simulate answering all questions in the deck
			for _, m := range deck {
				memID := m.ID
				accumulatedHistory = append(accumulatedHistory, model.AnswerLog{
					TargetMemberID: &memID,
					AnsweredAt:     time.Now().UTC(),
				})
				totalSeenSet[m.ID] = true
			}

			if len(totalSeenSet) == len(pool) {
				t.Logf("ADR-0020 Verified: All %d members successfully presented within %d sessions (no stranded members)",
					len(pool), session)
				return
			}
		}

		t.Fatalf("failed to present all members within %d sessions: %d/%d presented",
			maxSessions, len(totalSeenSet), len(pool))
	})

	t.Run("ADR-0020 Property 4: Graceful Degradation on Boundaries", func(t *testing.T) {
		// Boundary A: All Unseen (initial game)
		d1 := quiz.BuildBlendedDeck(pool, nil, 10, rng)
		if len(d1) != 10 {
			t.Fatalf("expected 10 items for all unseen, got %d", len(d1))
		}

		// Boundary B: All Seen (subsequent game after full mastery)
		allSeenHistory := make([]model.AnswerLog, len(pool))
		for i, m := range pool {
			id := m.ID
			allSeenHistory[i] = model.AnswerLog{TargetMemberID: &id}
		}
		d2 := quiz.BuildBlendedDeck(pool, allSeenHistory, 10, rng)
		if len(d2) != 10 {
			t.Fatalf("expected 10 items for all seen, got %d", len(d2))
		}

		// Boundary C: Pool smaller than deckSize (pool=5, deckSize=10)
		smallPool := generateMembers(5)
		d3 := quiz.BuildBlendedDeck(smallPool, nil, 10, rng)
		if len(d3) != 5 {
			t.Fatalf("expected 5 items clamped to pool size, got %d", len(d3))
		}

		// Boundary D: Zero or negative inputs
		if d := quiz.BuildBlendedDeck(nil, nil, 10, rng); d != nil {
			t.Fatalf("expected nil on nil pool, got %+v", d)
		}
		if d := quiz.BuildBlendedDeck(pool, nil, 0, rng); d != nil {
			t.Fatalf("expected nil on 0 deckSize, got %+v", d)
		}
		if d := quiz.BuildBlendedDeck(pool, nil, -5, rng); d != nil {
			t.Fatalf("expected nil on negative deckSize, got %+v", d)
		}
	})

	t.Run("ADR-0020 Property 5: Presentation Image Selection Rules", func(t *testing.T) {
		pht1 := model.ID("pht_uniform")
		pht2 := model.ID("pht_live")

		memberWithMultipleImages := model.Member{
			ID: "mem_multi",
			Images: []model.MemberImage{
				{ID: "img_01", PhotoTypeID: pht1, IsPrimary: true},
				{ID: "img_02", PhotoTypeID: pht2, IsPrimary: false},
			},
		}

		// Case A: Specific photo type filtered -> must select that photo type
		filterCostume := model.QuizFilter{PhotoTypeIDs: []model.ID{pht2}}
		imgSelected := quiz.SelectQuestionImage(memberWithMultipleImages, filterCostume, rng)
		if imgSelected == nil || imgSelected.PhotoTypeID != pht2 {
			t.Fatalf("expected image with photo type %s, got %+v", pht2, imgSelected)
		}

		// Case B: No costume filter -> selects randomly among available images
		imgRandom := quiz.SelectQuestionImage(memberWithMultipleImages, model.QuizFilter{}, rng)
		if imgRandom == nil || (imgRandom.ID != "img_01" && imgRandom.ID != "img_02") {
			t.Fatalf("expected one of member's images, got %+v", imgRandom)
		}

		// Case C: Member has no images -> gracefully falls back to nil or PrimaryImage
		memberWithoutImages := model.Member{ID: "mem_no_images"}
		imgFallback := quiz.SelectQuestionImage(memberWithoutImages, model.QuizFilter{}, rng)
		if imgFallback != nil {
			t.Fatalf("expected nil when no images exist, got %+v", imgFallback)
		}
	})
}
