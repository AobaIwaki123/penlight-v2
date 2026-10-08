package model

import "time"

// QuizOption represents one of the choices in a 4-choice penlight quiz.
type QuizOption struct {
	MemberID    ID           `json:"member_id"`
	MemberName  string       `json:"member_name"`
	Generation  int          `json:"generation"`
	Penlight    PenlightPair `json:"penlight"`
	LeftHex     string       `json:"left_hex"`
	RightHex    string       `json:"right_hex"`
	LeftName    string       `json:"left_name"`
	RightName   string       `json:"right_name"`
	IsCorrect   bool         `json:"is_correct"`
}

// QuizQuestion represents a 4-choice penlight quiz question.
type QuizQuestion struct {
	ID             ID           `json:"id"`               // quiz_... (UUID v7)
	TargetMemberID ID           `json:"target_member_id"` // mem_...
	TargetMember   Member       `json:"target_member"`
	Options        []QuizOption `json:"options"`          // Exactly 4 options
	CorrectIndex   int          `json:"correct_index"`    // 0-3
	GeneratedAt    time.Time    `json:"generated_at"`
}

// QuizFilter specifies candidate pool filtering criteria (Ref: ADR-0018, ADR-0026, ADR-0029).
type QuizFilter struct {
	SeriesID         *ID   `json:"series_id,omitempty"`
	GroupID          *ID   `json:"group_id,omitempty"`
	Generations      []int `json:"generations,omitempty"`
	PhotoTypeIDs     []ID  `json:"photo_type_ids,omitempty"`
	IncludeGraduated bool  `json:"include_graduated"`
	SongMode         bool  `json:"song_mode"` // true: 楽曲カラークイズ, false: メンバーカラークイズ (ADR-0029)
}
