package quiz

import (
	"github.com/aobaiwaki/penlight-v2/pkg/model"
)

// JudgeAnswer checks whether user's selected colors match target's correct colors as an unordered set (Ref: ADR-0019, ADR-0031).
// It supports single-color targets (1-color songs) and two-color targets (members, 2-color songs).
func JudgeAnswer(selectedColors []model.ID, target QuizTarget) bool {
	correctColors := target.GetCorrectColors()
	if len(selectedColors) != len(correctColors) {
		return false
	}

	counts := make(map[model.ID]int, len(correctColors))
	for _, c := range correctColors {
		counts[c]++
	}

	for _, c := range selectedColors {
		counts[c]--
		if counts[c] < 0 {
			return false
		}
	}

	return true
}

// JudgeAnswerPair is a convenience wrapper for checking two selected colors against a member or song (Ref: ADR-0019).
func JudgeAnswerPair(color1ID, color2ID model.ID, target QuizTarget) bool {
	return JudgeAnswer([]model.ID{color1ID, color2ID}, target)
}
