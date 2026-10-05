package model

import "time"

// Group represents an idol group entity (e.g. Hinatazaka46, Sakurazaka46, Nogizaka46, Aobazaka46).
// It is designed to be fully open for dynamic extension without any enum hardcoding or code changes.
type Group struct {
	ID            ID        `json:"id" db:"id,pk"`                     // grp_... (UUID v7)
	SeriesID      ID        `json:"series_id" db:"series_id,fk"`       // ser_... (UUID v7)
	Name          string    `json:"name" db:"name"`                   // Formal name, e.g. "日向坂46"
	ShortName     string    `json:"short_name" db:"short_name"`       // Short display name, e.g. "日向坂"
	Slug          string    `json:"slug" db:"slug,uk"`                 // URL-safe identifier, e.g. "hinatazaka46", "aobazaka46"
	ThemeColorHex string    `json:"theme_color_hex" db:"theme_color_hex"` // Official theme color, e.g. "#7CC7E8"
	DisplayOrder  int       `json:"display_order" db:"display_order"`   // UI sort order
	IsActive      bool      `json:"is_active" db:"is_active"`           // Active status
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}
