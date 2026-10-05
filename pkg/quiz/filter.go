package quiz

import (
	"fmt"
	"net/http"

	"github.com/aobaiwaki/penlight-v2/pkg/model"
)

// MinCandidateCount defines the minimum number of candidates required for a quiz pool.
const MinCandidateCount = 4

// FilterMembers filters members according to the specified criteria (Ref: ADR-0018, ADR-0026).
// It supports series-level isolation via filter.SeriesID when groups metadata is provided.
// It returns an AppError with CodeInsufficientMembers if fewer than MinCandidateCount members match.
func FilterMembers(members []model.Member, filter model.QuizFilter, groups ...model.Group) ([]model.Member, error) {
	// Prepare series group lookup if series isolation is requested
	var allowedGroupIDs map[model.ID]bool
	if filter.SeriesID != nil && len(groups) > 0 {
		allowedGroupIDs = make(map[model.ID]bool)
		for _, g := range groups {
			if g.SeriesID == *filter.SeriesID {
				allowedGroupIDs[g.ID] = true
			}
		}
	}

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
		// 1. Series filter (Ref: ADR-0026)
		if allowedGroupIDs != nil && !allowedGroupIDs[m.GroupID] {
			continue
		}

		// 2. Status filter: skip graduated members unless explicitly requested
		if !filter.IncludeGraduated && m.Status != "active" {
			continue
		}

		// 3. Group filter
		if filter.GroupID != nil && m.GroupID != *filter.GroupID {
			continue
		}

		// 4. Generation filter
		if len(genSet) > 0 && !genSet[m.Generation] {
			continue
		}

		// 5. Photo type (costume) filter: member must have at least one image matching the photo types
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

// FilterSongs filters songs according to the specified criteria (Ref: ADR-0026, ADR-0027, ADR-0029, ADR-0031).
// It supports series-level isolation via filter.SeriesID when groups metadata is provided.
// It returns an AppError with CodeInsufficientMembers if fewer than MinCandidateCount songs match.
func FilterSongs(songs []model.Song, filter model.QuizFilter, groups ...model.Group) ([]model.Song, error) {
	// Prepare series group lookup if series isolation is requested
	var allowedGroupIDs map[model.ID]bool
	if filter.SeriesID != nil && len(groups) > 0 {
		allowedGroupIDs = make(map[model.ID]bool)
		for _, g := range groups {
			if g.SeriesID == *filter.SeriesID {
				allowedGroupIDs[g.ID] = true
			}
		}
	}

	var filtered []model.Song

	for _, s := range songs {
		// 1. Series filter (Ref: ADR-0026)
		if allowedGroupIDs != nil && !allowedGroupIDs[s.GroupID] {
			continue
		}

		// 2. Group filter
		if filter.GroupID != nil && s.GroupID != *filter.GroupID {
			continue
		}

		filtered = append(filtered, s)
	}

	if len(filtered) < MinCandidateCount {
		return nil, model.AppError{
			Status: http.StatusBadRequest,
			Code:   model.CodeInsufficientMembers,
			Title:  "Insufficient Candidate Songs",
			Detail: fmt.Sprintf("Filtered songs count (%d) is below the minimum required (%d)", len(filtered), MinCandidateCount),
		}
	}

	return filtered, nil
}

// FilterColorsBySeries filters colors to ensure colors belonging to other series never contaminate the palette (Ref: ADR-0026).
// Standard common colors (GroupID == nil) and colors belonging to the specified series' groups are retained.
func FilterColorsBySeries(colors []model.Color, groups []model.Group, seriesID model.ID) []model.Color {
	seriesGroupIDs := make(map[model.ID]bool)
	for _, g := range groups {
		if g.SeriesID == seriesID {
			seriesGroupIDs[g.ID] = true
		}
	}

	var filtered []model.Color
	for _, c := range colors {
		if c.GroupID == nil || seriesGroupIDs[*c.GroupID] {
			filtered = append(filtered, c)
		}
	}
	return filtered
}
