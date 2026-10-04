package quiz

import (
	"math/rand"

	"github.com/aobaiwaki/penlight-v2/pkg/model"
)

// DefaultUnseenRatio is the target ratio of unseen members in a deck (Ref: ADR-0020).
const DefaultUnseenRatio = 0.7

// BuildBlendedDeck creates a quiz deck by blending unseen members and review (seen) members
// based on past answer logs (Ref: ADR-0020).
// It guarantees zero duplicate members within the generated deck.
func BuildBlendedDeck(
	pool []model.Member,
	history []model.AnswerLog,
	deckSize int,
	rng *rand.Rand,
) []model.Member {
	if len(pool) == 0 || deckSize <= 0 {
		return nil
	}

	// 1. Build set of seen member IDs from history
	seenIDs := make(map[model.ID]bool, len(history))
	for _, log := range history {
		seenIDs[log.TargetMemberID] = true
	}

	// 2. Partition pool into unseen and seen members
	var unseen, seen []model.Member
	for _, m := range pool {
		if seenIDs[m.ID] {
			seen = append(seen, m)
		} else {
			unseen = append(unseen, m)
		}
	}

	// 3. Shuffle unseen and seen groups separately using Fisher-Yates
	shuffleMembers(unseen, rng)
	shuffleMembers(seen, rng)

	// 4. Calculate target allocation
	actualDeckSize := deckSize
	if len(pool) < actualDeckSize {
		actualDeckSize = len(pool)
	}

	targetUnseen := int(float64(actualDeckSize) * DefaultUnseenRatio)
	takeUnseen := len(unseen)
	if takeUnseen > targetUnseen {
		takeUnseen = targetUnseen
	}

	remainingNeeded := actualDeckSize - takeUnseen
	takeSeen := len(seen)
	if takeSeen > remainingNeeded {
		takeSeen = remainingNeeded
	}

	// If we haven't reached actualDeckSize and there are leftover unseen members, take more unseen
	if takeUnseen+takeSeen < actualDeckSize && len(unseen) > takeUnseen {
		additionalUnseen := actualDeckSize - (takeUnseen + takeSeen)
		leftoverUnseen := len(unseen) - takeUnseen
		if additionalUnseen > leftoverUnseen {
			additionalUnseen = leftoverUnseen
		}
		takeUnseen += additionalUnseen
	}

	// Combine into final deck
	deck := make([]model.Member, 0, actualDeckSize)
	deck = append(deck, unseen[:takeUnseen]...)
	deck = append(deck, seen[:takeSeen]...)

	// 5. Final shuffle to randomize presentation order
	shuffleMembers(deck, rng)

	return deck
}

// shuffleMembers performs an in-place Fisher-Yates shuffle on a slice of members.
func shuffleMembers(members []model.Member, rng *rand.Rand) {
	for i := len(members) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		members[i], members[j] = members[j], members[i]
	}
}
