package model

import "time"

// CurrentMasterVersion defines the canonical version of master data (seeds/data/*.json).
const CurrentMasterVersion = "2026.10.05-3"

// MasterVersion tracks the current master data synchronization version (Ref: ADR-0021).
type MasterVersion struct {
	ID        string    `json:"id"`         // "current"
	Version   string    `json:"version"`    // Git commit hash or semantic version
	UpdatedAt time.Time `json:"updated_at"`
}
