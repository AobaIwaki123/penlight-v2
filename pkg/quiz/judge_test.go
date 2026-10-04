package quiz_test

import (
	"testing"

	"github.com/aobaiwaki/penlight-v2/pkg/model"
	"github.com/aobaiwaki/penlight-v2/pkg/quiz"
)

func TestJudgeAnswer(t *testing.T) {
	blue := model.ID("col_blue")
	white := model.ID("col_white")
	red := model.ID("col_red")

	// Member with two different colors
	memberDiff := model.Member{
		ID: "mem_01",
		Penlight: model.PenlightPair{
			LeftColorID:  blue,
			RightColorID: white,
		},
	}

	// Member with identical colors (e.g. white x white)
	memberSame := model.Member{
		ID: "mem_02",
		Penlight: model.PenlightPair{
			LeftColorID:  white,
			RightColorID: white,
		},
	}

	tests := []struct {
		name     string
		target   model.Member
		c1       model.ID
		c2       model.ID
		expected bool
	}{
		{
			name:     "Direct order match (blue, white)",
			target:   memberDiff,
			c1:       blue,
			c2:       white,
			expected: true,
		},
		{
			name:     "Inverted order match (white, blue) - ADR-0019",
			target:   memberDiff,
			c1:       white,
			c2:       blue,
			expected: true,
		},
		{
			name:     "One color incorrect (blue, red)",
			target:   memberDiff,
			c1:       blue,
			c2:       red,
			expected: false,
		},
		{
			name:     "Both colors incorrect (red, red)",
			target:   memberDiff,
			c1:       red,
			c2:       red,
			expected: false,
		},
		{
			name:     "Same-color member match (white, white)",
			target:   memberSame,
			c1:       white,
			c2:       white,
			expected: true,
		},
		{
			name:     "Same-color member mismatch (white, blue)",
			target:   memberSame,
			c1:       white,
			c2:       blue,
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := quiz.JudgeAnswer(tc.c1, tc.c2, tc.target)
			if result != tc.expected {
				t.Fatalf("JudgeAnswer(%s, %s) = %v, expected %v", tc.c1, tc.c2, result, tc.expected)
			}
		})
	}
}
