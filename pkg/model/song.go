package model

import "time"

// Song represents a musical track entity and its official/live penlight colors.
type Song struct {
	ID        ID        `json:"id" db:"id,pk"`                          // sng_... (UUID v7)
	GroupID   ID        `json:"group_id" db:"group_id,fk"`              // 所属グループ (grp_...)
	Title     string    `json:"title" db:"title"`                       // 楽曲タイトル (例: "絶対アイドル辞めないで")
	Kana      *string   `json:"kana,omitempty" db:"kana"`               // 読み仮名 (任意・未設定可, ソート補助用)
	Color1ID  ID        `json:"color1_id" db:"color1_id,fk"`            // 1色目 (必須, col_...)
	Color2ID  *ID       `json:"color2_id,omitempty" db:"color2_id,fk"`  // 2色目 (任意, col_...。1色の曲は nil)
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// GetID returns the song's surrogate key ID (implements QuizTarget, Ref: ADR-0031).
func (s Song) GetID() ID {
	return s.ID
}

// GetCorrectColors returns the 1 or 2 penlight color IDs for the song (implements QuizTarget, Ref: ADR-0027, ADR-0031).
func (s Song) GetCorrectColors() []ID {
	if s.Color2ID != nil {
		return []ID{s.Color1ID, *s.Color2ID}
	}
	return []ID{s.Color1ID}
}

