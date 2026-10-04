package quiz_test

import (
	"testing"

	"github.com/aobaiwaki/penlight-v2/pkg/model"
	"github.com/aobaiwaki/penlight-v2/pkg/quiz"
)

func createMockMembers() []model.Member {
	grp1 := model.ID("grp_hinata")
	grp2 := model.ID("grp_sakura")
	pht1 := model.ID("pht_13th_uniform")
	pht2 := model.ID("pht_1st_album")

	return []model.Member{
		{
			ID: "mem_01", GroupID: grp1, Generation: 4, Status: "active",
			Images: []model.MemberImage{{PhotoTypeID: pht1}},
		},
		{
			ID: "mem_02", GroupID: grp1, Generation: 4, Status: "active",
			Images: []model.MemberImage{{PhotoTypeID: pht1}},
		},
		{
			ID: "mem_03", GroupID: grp1, Generation: 4, Status: "active",
			Images: []model.MemberImage{{PhotoTypeID: pht1}},
		},
		{
			ID: "mem_04", GroupID: grp1, Generation: 4, Status: "active",
			Images: []model.MemberImage{{PhotoTypeID: pht2}},
		},
		{
			ID: "mem_05", GroupID: grp1, Generation: 3, Status: "active",
			Images: []model.MemberImage{{PhotoTypeID: pht2}},
		},
		{
			ID: "mem_06", GroupID: grp1, Generation: 1, Status: "graduated",
			Images: []model.MemberImage{{PhotoTypeID: pht1}},
		},
		{
			ID: "mem_07", GroupID: grp2, Generation: 2, Status: "active",
			Images: []model.MemberImage{{PhotoTypeID: pht2}},
		},
		{
			ID: "mem_08", GroupID: grp2, Generation: 2, Status: "active",
			Images: []model.MemberImage{{PhotoTypeID: pht2}},
		},
		{
			ID: "mem_09", GroupID: grp2, Generation: 2, Status: "active",
			Images: []model.MemberImage{{PhotoTypeID: pht2}},
		},
		{
			ID: "mem_10", GroupID: grp2, Generation: 2, Status: "active",
			Images: []model.MemberImage{{PhotoTypeID: pht2}},
		},
	}
}

func TestFilterMembers(t *testing.T) {
	members := createMockMembers()
	grp1 := model.ID("grp_hinata")
	pht1 := model.ID("pht_13th_uniform")

	tests := []struct {
		name          string
		filter        model.QuizFilter
		expectedCount int
		expectError   bool
	}{
		{
			name:          "No filter (all active members)",
			filter:        model.QuizFilter{},
			expectedCount: 9, // mem_06 is graduated
			expectError:   false,
		},
		{
			name: "Include graduated members",
			filter: model.QuizFilter{
				IncludeGraduated: true,
			},
			expectedCount: 10,
			expectError:   false,
		},
		{
			name: "Filter by group 1",
			filter: model.QuizFilter{
				GroupID: &grp1,
			},
			expectedCount: 5, // mem_01 to mem_05
			expectError:   false,
		},
		{
			name: "Filter by group 1 and generation 4",
			filter: model.QuizFilter{
				GroupID:     &grp1,
				Generations: []int{4},
			},
			expectedCount: 4, // mem_01, mem_02, mem_03, mem_04
			expectError:   false,
		},
		{
			name: "Filter by photo type pht1",
			filter: model.QuizFilter{
				GroupID:      &grp1,
				PhotoTypeIDs: []model.ID{pht1},
			},
			expectedCount: 3, // mem_01, mem_02, mem_03 (active only) -> insufficient!
			expectError:   true,
		},
		{
			name: "Filter by photo type pht1 with graduated",
			filter: model.QuizFilter{
				GroupID:          &grp1,
				PhotoTypeIDs:     []model.ID{pht1},
				IncludeGraduated: true,
			},
			expectedCount: 4, // mem_01, mem_02, mem_03, mem_06
			expectError:   false,
		},
		{
			name: "Insufficient candidates (fewer than 4)",
			filter: model.QuizFilter{
				GroupID:     &grp1,
				Generations: []int{3},
			},
			expectError: true, // only 1 member (mem_05)
		},
		{
			name: "Non-existent group",
			filter: model.QuizFilter{
				GroupID: func() *model.ID { id := model.ID("grp_none"); return &id }(),
			},
			expectError: true,
		},
		{
			name: "Multiple generations (2 and 3)",
			filter: model.QuizFilter{
				Generations: []int{2, 3},
			},
			expectedCount: 5, // mem_05, mem_07, mem_08, mem_09, mem_10
			expectError:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res, err := quiz.FilterMembers(members, tc.filter)
			if tc.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil with %d members", len(res))
				}
				appErr, ok := err.(model.AppError)
				if !ok {
					t.Fatalf("expected model.AppError, got %T: %v", err, err)
				}
				if appErr.Code != model.CodeInsufficientMembers {
					t.Fatalf("expected code %s, got %s", model.CodeInsufficientMembers, appErr.Code)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(res) != tc.expectedCount {
					t.Fatalf("expected %d members, got %d", tc.expectedCount, len(res))
				}
			}
		})
	}
}
