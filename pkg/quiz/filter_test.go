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

func TestFilterMembers_SeriesIsolation(t *testing.T) {
	serSakamichi := model.ID("ser_sakamichi")
	serIkolove := model.ID("ser_ikolove")

	groups := []model.Group{
		{ID: "grp_hinata", SeriesID: serSakamichi, Name: "日向坂46"},
		{ID: "grp_sakura", SeriesID: serSakamichi, Name: "櫻坂46"},
		{ID: "grp_equal", SeriesID: serIkolove, Name: "=LOVE"},
	}

	members := []model.Member{
		{ID: "mem_h1", GroupID: "grp_hinata", Status: "active"},
		{ID: "mem_h2", GroupID: "grp_hinata", Status: "active"},
		{ID: "mem_s1", GroupID: "grp_sakura", Status: "active"},
		{ID: "mem_s2", GroupID: "grp_sakura", Status: "active"},
		{ID: "mem_e1", GroupID: "grp_equal", Status: "active"},
		{ID: "mem_e2", GroupID: "grp_equal", Status: "active"},
		{ID: "mem_e3", GroupID: "grp_equal", Status: "active"},
		{ID: "mem_e4", GroupID: "grp_equal", Status: "active"},
	}

	t.Run("ADR-0026: Sakamichi series isolation filters out =LOVE members", func(t *testing.T) {
		res, err := quiz.FilterMembers(members, model.QuizFilter{SeriesID: &serSakamichi}, groups...)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(res) != 4 {
			t.Fatalf("expected 4 sakamichi members, got %d", len(res))
		}
		for _, m := range res {
			if m.GroupID == "grp_equal" {
				t.Fatalf("=LOVE member %s leaked into sakamichi pool", m.ID)
			}
		}
	})

	t.Run("ADR-0026: Ikolove series isolation filters out Sakamichi members", func(t *testing.T) {
		res, err := quiz.FilterMembers(members, model.QuizFilter{SeriesID: &serIkolove}, groups...)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(res) != 4 {
			t.Fatalf("expected 4 ikolove members, got %d", len(res))
		}
		for _, m := range res {
			if m.GroupID != "grp_equal" {
				t.Fatalf("non-ikolove member %s leaked into ikolove pool", m.ID)
			}
		}
	})
}

func TestFilterSongs(t *testing.T) {
	serSakamichi := model.ID("ser_sakamichi")
	serIkolove := model.ID("ser_ikolove")

	groups := []model.Group{
		{ID: "grp_hinata", SeriesID: serSakamichi, Name: "日向坂46"},
		{ID: "grp_equal", SeriesID: serIkolove, Name: "=LOVE"},
	}

	songs := []model.Song{
		{ID: "sng_h1", GroupID: "grp_hinata", Title: "キュン", Color1ID: "col_sky"},
		{ID: "sng_h2", GroupID: "grp_hinata", Title: "ドレミソラシド", Color1ID: "col_sky"},
		{ID: "sng_h3", GroupID: "grp_hinata", Title: "こんなに好きになっちゃっていいの？", Color1ID: "col_sky"},
		{ID: "sng_h4", GroupID: "grp_hinata", Title: "ソンナコトナイヨ", Color1ID: "col_sky"},
		{ID: "sng_e1", GroupID: "grp_equal", Title: "絶対アイドル辞めないで", Color1ID: "col_pink"},
		{ID: "sng_e2", GroupID: "grp_equal", Title: "あの子コンプレックス", Color1ID: "col_blue"},
		{ID: "sng_e3", GroupID: "grp_equal", Title: "探せ ダイヤモンドリリー", Color1ID: "col_orange"},
		{ID: "sng_e4", GroupID: "grp_equal", Title: "青春”サブリミナル”", Color1ID: "col_yellow"},
	}

	t.Run("ADR-0026 & ADR-0027: Filter songs by Series", func(t *testing.T) {
		res, err := quiz.FilterSongs(songs, model.QuizFilter{SeriesID: &serSakamichi}, groups...)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(res) != 4 {
			t.Fatalf("expected 4 sakamichi songs, got %d", len(res))
		}
		for _, s := range res {
			if s.GroupID != "grp_hinata" {
				t.Fatalf("leaked song from other series: %+v", s)
			}
		}
	})

	t.Run("Filter songs with insufficient candidates", func(t *testing.T) {
		dummyGroup := model.ID("grp_dummy")
		_, err := quiz.FilterSongs(songs, model.QuizFilter{GroupID: &dummyGroup}, groups...)
		if err == nil {
			t.Fatal("expected error for insufficient candidate songs")
		}
		appErr, ok := err.(model.AppError)
		if !ok || appErr.Code != model.CodeInsufficientMembers {
			t.Fatalf("expected CodeInsufficientMembers, got %v", err)
		}
	})
}

func TestFilterColorsBySeries(t *testing.T) {
	serSakamichi := model.ID("ser_sakamichi")
	serIkolove := model.ID("ser_ikolove")

	grpSakamichi := model.ID("grp_hinata")
	grpIkolove := model.ID("grp_equal")

	groups := []model.Group{
		{ID: grpSakamichi, SeriesID: serSakamichi},
		{ID: grpIkolove, SeriesID: serIkolove},
	}

	colors := []model.Color{
		{ID: "col_common_white", GroupID: nil, Name: "白"},
		{ID: "col_hinata_sky", GroupID: &grpSakamichi, Name: "スカイブルー"},
		{ID: "col_ikolove_pink", GroupID: &grpIkolove, Name: "イコラブピンク"},
	}

	t.Run("ADR-0026: Palette excludes other series colors", func(t *testing.T) {
		sakamichiColors := quiz.FilterColorsBySeries(colors, groups, serSakamichi)
		if len(sakamichiColors) != 2 {
			t.Fatalf("expected 2 colors (common white + hinata sky), got %d", len(sakamichiColors))
		}
		for _, c := range sakamichiColors {
			if c.ID == "col_ikolove_pink" {
				t.Fatal("=LOVE color leaked into sakamichi palette")
			}
		}

		ikoloveColors := quiz.FilterColorsBySeries(colors, groups, serIkolove)
		if len(ikoloveColors) != 2 {
			t.Fatalf("expected 2 colors (common white + ikolove pink), got %d", len(ikoloveColors))
		}
		for _, c := range ikoloveColors {
			if c.ID == "col_hinata_sky" {
				t.Fatal("hinata color leaked into ikolove palette")
			}
		}
	})
}
