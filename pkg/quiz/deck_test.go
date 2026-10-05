package quiz_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/aobaiwaki/penlight-v2/pkg/model"
	"github.com/aobaiwaki/penlight-v2/pkg/quiz"
)

func generateMembers(count int) []model.Member {
	members := make([]model.Member, count)
	for i := 0; i < count; i++ {
		members[i] = model.Member{
			ID:         model.ID(fmt.Sprintf("mem_%02d", i+1)),
			FamilyName: fmt.Sprintf("姓%d", i+1),
			GivenName:  fmt.Sprintf("名%d", i+1),
		}
	}
	return members
}

func TestBuildBlendedDeck(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	pool := generateMembers(20)

	t.Run("Zero duplicate guarantee in generated deck", func(t *testing.T) {
		deck := quiz.BuildBlendedDeck(pool, nil, 10, rng)
		if len(deck) != 10 {
			t.Fatalf("expected deck of 10, got %d", len(deck))
		}

		seen := make(map[model.ID]bool)
		for _, m := range deck {
			if seen[m.ID] {
				t.Fatalf("duplicate member %s found in deck", m.ID)
			}
			seen[m.ID] = true
		}
	})

	t.Run("All unseen: fully populated from unseen members", func(t *testing.T) {
		deck := quiz.BuildBlendedDeck(pool, nil, 10, rng)
		if len(deck) != 10 {
			t.Fatalf("expected 10 members, got %d", len(deck))
		}
	})

	t.Run("All seen: fallback to seen members", func(t *testing.T) {
		history := make([]model.AnswerLog, len(pool))
		for i := range pool {
			id := pool[i].ID
			history[i] = model.AnswerLog{TargetMemberID: &id}
		}

		deck := quiz.BuildBlendedDeck(pool, history, 10, rng)
		if len(deck) != 10 {
			t.Fatalf("expected 10 members, got %d", len(deck))
		}

		seen := make(map[model.ID]bool)
		for _, m := range deck {
			if seen[m.ID] {
				t.Fatalf("duplicate member %s found in deck", m.ID)
			}
			seen[m.ID] = true
		}
	})

	t.Run("Blended ratio: prioritizes unseen and includes review (seen)", func(t *testing.T) {
		// 10 members seen, 10 members unseen
		history := make([]model.AnswerLog, 10)
		for i := 0; i < 10; i++ {
			id := pool[i].ID
			history[i] = model.AnswerLog{TargetMemberID: &id}
		}

		deck := quiz.BuildBlendedDeck(pool, history, 10, rng)
		if len(deck) != 10 {
			t.Fatalf("expected 10 members, got %d", len(deck))
		}

		unseenCount := 0
		seenCount := 0
		seenSet := make(map[model.ID]bool)
		for _, log := range history {
			seenSet[log.GetTargetID()] = true
		}

		for _, m := range deck {
			if seenSet[m.ID] {
				seenCount++
			} else {
				unseenCount++
			}
		}

		// Target ratio is 7:3
		if unseenCount != 7 || seenCount != 3 {
			t.Fatalf("expected 7 unseen and 3 seen, got %d unseen and %d seen", unseenCount, seenCount)
		}
	})

	t.Run("Pool smaller than deckSize", func(t *testing.T) {
		smallPool := generateMembers(5)
		deck := quiz.BuildBlendedDeck(smallPool, nil, 10, rng)
		if len(deck) != 5 {
			t.Fatalf("expected deck size clamped to 5, got %d", len(deck))
		}
	})

	t.Run("Empty pool or non-positive deckSize", func(t *testing.T) {
		if deck := quiz.BuildBlendedDeck(nil, nil, 10, rng); deck != nil {
			t.Fatalf("expected nil for empty pool, got %+v", deck)
		}
		if deck := quiz.BuildBlendedDeck(pool, nil, 0, rng); deck != nil {
			t.Fatalf("expected nil for zero deckSize, got %+v", deck)
		}
	})

	t.Run("Randomness over multiple executions", func(t *testing.T) {
		r1 := rand.New(rand.NewSource(time.Now().UnixNano()))
		r2 := rand.New(rand.NewSource(time.Now().UnixNano() + 100))

		deck1 := quiz.BuildBlendedDeck(pool, nil, 10, r1)
		deck2 := quiz.BuildBlendedDeck(pool, nil, 10, r2)

		different := false
		for i := range deck1 {
			if deck1[i].ID != deck2[i].ID {
				different = true
				break
			}
		}
		if !different {
			t.Log("Note: decks happened to be identical (unlikely with 20 items)")
		}
	})
}
