package model

import "time"

// Color represents an official penlight color.
// Colors can be group-specific or shared standard colors (when GroupID is nil).
type Color struct {
	ID           ID        `json:"id"`                     // col_... (UUID v7)
	GroupID      *ID       `json:"group_id,omitempty"`     // Optional group owner, nil for common/standard colors
	Name         string    `json:"name"`                   // Color name, e.g. "スカイブルー", "パステルブルー", "青"
	HexCode      string    `json:"hex_code"`               // Hex color code, e.g. "#00BFFF", "#7CC7E8"
	DisplayOrder int       `json:"display_order"`          // UI sort order within the color selector
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
