package quiz

import (
	"github.com/aobaiwaki/penlight-v2/pkg/model"
)

// JudgeAnswer verifies whether the user-selected two colors match the member's official penlight colors.
// In accordance with ADR-0019, order is not considered (either left/right orientation is accepted).
func JudgeAnswer(color1ID, color2ID model.ID, target model.Member) bool {
	correctLeft := target.Penlight.LeftColorID
	correctRight := target.Penlight.RightColorID

	// 1. Direct match (L==ans.L && R==ans.R)
	if color1ID == correctLeft && color2ID == correctRight {
		return true
	}

	// 2. Inverted match (L==ans.R && R==ans.L)
	if color1ID == correctRight && color2ID == correctLeft {
		return true
	}

	return false
}
