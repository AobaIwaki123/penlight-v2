package model

import "time"

// Series represents an idol franchise or series (e.g. "坂道シリーズ", "=LOVE系列").
type Series struct {
	ID           ID        `json:"id" db:"id,pk"`                       // ser_... (UUID v7)
	Name         string    `json:"name" db:"name"`                     // Formal name, e.g. "坂道シリーズ", "=LOVE系列"
	Slug         string    `json:"slug" db:"slug,uk"`                   // URL-safe identifier, e.g. "sakamichi", "ikolove"
	DisplayOrder int       `json:"display_order" db:"display_order"`   // UI sort order
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}
