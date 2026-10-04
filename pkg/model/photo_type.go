package model

import "time"

// PhotoType represents a costume or photo category (Ref: ADR-0021).
type PhotoType struct {
	ID           ID        `json:"id"`            // pht_... (UUID v7 surrogate key)
	GroupID      ID        `json:"group_id"`      // grp_... (FK)
	Slug         string    `json:"slug"`          // e.g. "13th_single", "2nd_album"
	Name         string    `json:"name"`          // e.g. "13th Single 制服", "2nd Album アー写"
	DisplayOrder int       `json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
