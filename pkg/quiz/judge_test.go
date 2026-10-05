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
	pink := model.ID("col_pink")

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

	// Song with 1 color (Ref: ADR-0027)
	song1Color := model.Song{
		ID:       "sng_01",
		Title:    "絶対アイドル辞めないで",
		Color1ID: pink,
	}

	// Song with 2 colors (Ref: ADR-0027)
	song2Color := model.Song{
		ID:       "sng_02",
		Title:    "あの子コンプレックス",
		Color1ID: blue,
		Color2ID: &white,
	}

	tests := []struct {
		name     string
		target   quiz.QuizTarget
		selected []model.ID
		expected bool
	}{
		// Member 2-color tests
		{
			name:     "Member: Direct order match (blue, white)",
			target:   memberDiff,
			selected: []model.ID{blue, white},
			expected: true,
		},
		{
			name:     "Member: Inverted order match (white, blue) - ADR-0019",
			target:   memberDiff,
			selected: []model.ID{white, blue},
			expected: true,
		},
		{
			name:     "Member: One color incorrect (blue, red)",
			target:   memberDiff,
			selected: []model.ID{blue, red},
			expected: false,
		},
		{
			name:     "Member: Both colors incorrect (red, red)",
			target:   memberDiff,
			selected: []model.ID{red, red},
			expected: false,
		},
		{
			name:     "Member: Same-color match (white, white)",
			target:   memberSame,
			selected: []model.ID{white, white},
			expected: true,
		},
		{
			name:     "Member: Same-color mismatch (white, blue)",
			target:   memberSame,
			selected: []model.ID{white, blue},
			expected: false,
		},
		{
			name:     "Member: Length mismatch (1 color for 2-color member)",
			target:   memberDiff,
			selected: []model.ID{blue},
			expected: false,
		},

		// Song 1-color tests (Ref: ADR-0027, ADR-0031)
		{
			name:     "Song 1-color: Correct color match (pink)",
			target:   song1Color,
			selected: []model.ID{pink},
			expected: true,
		},
		{
			name:     "Song 1-color: Incorrect color (blue)",
			target:   song1Color,
			selected: []model.ID{blue},
			expected: false,
		},
		{
			name:     "Song 1-color: Length mismatch (2 colors selected for 1-color song)",
			target:   song1Color,
			selected: []model.ID{pink, pink},
			expected: false,
		},

		// Song 2-color tests (Ref: ADR-0027, ADR-0031)
		{
			name:     "Song 2-color: Direct match (blue, white)",
			target:   song2Color,
			selected: []model.ID{blue, white},
			expected: true,
		},
		{
			name:     "Song 2-color: Inverted match (white, blue)",
			target:   song2Color,
			selected: []model.ID{white, blue},
			expected: true,
		},
		{
			name:     "Song 2-color: Mismatched color (blue, red)",
			target:   song2Color,
			selected: []model.ID{blue, red},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := quiz.JudgeAnswer(tc.selected, tc.target)
			if result != tc.expected {
				t.Fatalf("JudgeAnswer(%v, %s) = %v, expected %v", tc.selected, tc.target.GetID(), result, tc.expected)
			}
		})
	}

	// Test JudgeAnswerPair backward compatibility
	t.Run("JudgeAnswerPair compatibility helper", func(t *testing.T) {
		if !quiz.JudgeAnswerPair(blue, white, memberDiff) {
			t.Fatal("expected JudgeAnswerPair to return true for member direct match")
		}
		if !quiz.JudgeAnswerPair(white, blue, memberDiff) {
			t.Fatal("expected JudgeAnswerPair to return true for member inverted match")
		}
		if quiz.JudgeAnswerPair(blue, red, memberDiff) {
			t.Fatal("expected JudgeAnswerPair to return false for mismatch")
		}
	})
}
