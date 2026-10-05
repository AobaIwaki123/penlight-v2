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
