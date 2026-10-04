package quiz

import (
	"fmt"
	"net/http"

	"github.com/aobaiwaki/penlight-v2/pkg/model"
)

// MinCandidateCount defines the minimum number of candidates required for a quiz pool.
const MinCandidateCount = 4

// FilterMembers filters members according to the specified criteria (Ref: ADR-0018).
// It returns an AppError with CodeInsufficientMembers if fewer than MinCandidateCount members match.
func FilterMembers(members []model.Member, filter model.QuizFilter) ([]model.Member, error) {
	// Prepare lookup sets for fast matching
	genSet := make(map[int]bool, len(filter.Generations))
	for _, g := range filter.Generations {
		genSet[g] = true
	}

	phtSet := make(map[model.ID]bool, len(filter.PhotoTypeIDs))
	for _, p := range filter.PhotoTypeIDs {
		phtSet[p] = true
	}

	var filtered []model.Member

	for _, m := range members {
		// 1. Status filter: skip graduated members unless explicitly requested
		if !filter.IncludeGraduated && m.Status != "active" {
			continue
		}

		// 2. Group filter
		if filter.GroupID != nil && m.GroupID != *filter.GroupID {
			continue
		}

		// 3. Generation filter
		if len(genSet) > 0 && !genSet[m.Generation] {
			continue
		}

		// 4. Photo type (costume) filter: member must have at least one image matching the photo types
		if len(phtSet) > 0 {
			hasMatchingPhoto := false
			for _, img := range m.Images {
				if phtSet[img.PhotoTypeID] {
					hasMatchingPhoto = true
					break
				}
			}
			if !hasMatchingPhoto {
				continue
			}
		}

		filtered = append(filtered, m)
	}

	if len(filtered) < MinCandidateCount {
		return nil, model.AppError{
			Status: http.StatusBadRequest,
			Code:   model.CodeInsufficientMembers,
			Title:  "Insufficient Candidate Members",
			Detail: fmt.Sprintf("Filtered candidates count (%d) is below the minimum required (%d)", len(filtered), MinCandidateCount),
		}
	}

	return filtered, nil
}
