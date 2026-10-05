package quiz

import (
	"math/rand"

	"github.com/aobaiwaki/penlight-v2/pkg/model"
)

// DefaultUnseenRatio is the target ratio of unseen members in a deck (Ref: ADR-0020).
const DefaultUnseenRatio = 0.7

// BuildBlendedDeck creates a quiz deck by blending unseen items and review (seen) items
// based on past answer logs for ANY QuizTarget (Ref: ADR-0020, ADR-0031).
// It guarantees zero duplicate items within the generated deck.
func BuildBlendedDeck[T QuizTarget](
	pool []T,
	history []model.AnswerLog,
	deckSize int,
	rng *rand.Rand,
) []T {
	if len(pool) == 0 || deckSize <= 0 {
		return nil
	}

	// 1. Build set of seen target IDs from history
	seenIDs := make(map[model.ID]bool, len(history))
	for _, log := range history {
		if targetID := log.GetTargetID(); targetID != "" {
			seenIDs[targetID] = true
		}
	}

	// 2. Partition pool into unseen and seen items
	var unseen, seen []T
	for _, item := range pool {
		if seenIDs[item.GetID()] {
			seen = append(seen, item)
		} else {
			unseen = append(unseen, item)
		}
	}

	// 3. Shuffle unseen and seen groups separately using Fisher-Yates
	shuffleItems(unseen, rng)
	shuffleItems(seen, rng)

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

	// If we haven't reached actualDeckSize and there are leftover unseen items, take more unseen
	if takeUnseen+takeSeen < actualDeckSize && len(unseen) > takeUnseen {
		additionalUnseen := actualDeckSize - (takeUnseen + takeSeen)
		leftoverUnseen := len(unseen) - takeUnseen
		if additionalUnseen > leftoverUnseen {
			additionalUnseen = leftoverUnseen
		}
		takeUnseen += additionalUnseen
	}

	// Combine into final deck
	deck := make([]T, 0, actualDeckSize)
	deck = append(deck, unseen[:takeUnseen]...)
	deck = append(deck, seen[:takeSeen]...)

	// 5. Final shuffle to randomize presentation order
	shuffleItems(deck, rng)

	return deck
}

// shuffleItems performs an in-place Fisher-Yates shuffle on any slice.
func shuffleItems[T any](items []T, rng *rand.Rand) {
	for i := len(items) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		items[i], items[j] = items[j], items[i]
	}
}

// SelectQuestionImage selects the presentation image for a member based on the quiz filter (Ref: ADR-0020).
// If the filter specifies photo types, an image matching one of those photo types is chosen.
// Otherwise, an image is randomly selected from the member's available images (falling back to primary image).
func SelectQuestionImage(member model.Member, filter model.QuizFilter, rng *rand.Rand) *model.MemberImage {
	if len(member.Images) == 0 {
		return member.PrimaryImage()
	}

	phtSet := make(map[model.ID]bool, len(filter.PhotoTypeIDs))
	for _, id := range filter.PhotoTypeIDs {
		phtSet[id] = true
	}

	if len(phtSet) > 0 {
		var matched []model.MemberImage
		for _, img := range member.Images {
			if phtSet[img.PhotoTypeID] {
				matched = append(matched, img)
			}
		}
		if len(matched) > 0 {
			idx := rng.Intn(len(matched))
			return &matched[idx]
		}
	}

	// When no costume filter is specified, randomly select among available images
	idx := rng.Intn(len(member.Images))
	return &member.Images[idx]
}

