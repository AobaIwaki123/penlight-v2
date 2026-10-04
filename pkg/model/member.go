package model

import (
	"fmt"
	"time"
)

// MemberStatus represents the member's current active status.
type MemberStatus string

const (
	StatusActive    MemberStatus = "active"    // 現役活動中
	StatusGraduated MemberStatus = "graduated" // 卒業
	StatusHiatus    MemberStatus = "hiatus"    // 休業中
)

// PenlightPair holds the two penlight colors assigned to a member.
type PenlightPair struct {
	LeftColorID  ID   `json:"left_color_id"`  // col_... (Left hand / Color 1)
	RightColorID ID   `json:"right_color_id"` // col_... (Right hand / Color 2)
	Ordered      bool `json:"ordered"`        // true if position matters, false if symmetric
}

// MemberImage represents an image asset associated with a member (Ref: ADR-0021).
type MemberImage struct {
	ID           ID         `json:"id"`                    // img_... (UUID v7 surrogate key)
	MemberID     ID         `json:"member_id"`             // mem_... (FK)
	PhotoTypeID  ID         `json:"photo_type_id"`         // pht_... (FK, Ref: ADR-0021)
	PhotoType    *PhotoType `json:"photo_type,omitempty"`  // Associated costume/photo category
	ImageKey     string     `json:"image_key"`             // Immutable image filename: e.g. "img_<uuidv7>.webp"
	IsPrimary    bool       `json:"is_primary"`            // true for the primary/default image
	DisplayOrder int        `json:"display_order"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}


// Member represents an idol member.
// Natural keys (names) are never used as identifiers to prevent collisions on homonyms or name changes.
type Member struct {
	ID             ID            `json:"id"`                     // mem_... (UUID v7 immutable surrogate key)
	GroupID        ID            `json:"group_id"`               // grp_...
	FamilyName     string        `json:"family_name"`            // e.g. "加藤"
	GivenName      string        `json:"given_name"`             // e.g. "史帆"
	FamilyNameKana string        `json:"family_name_kana"`       // e.g. "かとう"
	GivenNameKana  string        `json:"given_name_kana"`        // e.g. "しほ"
	Generation     int           `json:"generation"`             // e.g. 1 (1期生)
	Status         MemberStatus  `json:"status"`                 // active, graduated, hiatus
	Penlight       PenlightPair  `json:"penlight"`               // Assigned penlight colors
	Images         []MemberImage `json:"images,omitempty"`       // Associated images (1:N, Ref: ADR-0021)
	JoinedAt       *time.Time    `json:"joined_at,omitempty"`    // Optional joining date
	GraduatedAt    *time.Time    `json:"graduated_at,omitempty"` // Optional graduation date
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

// PrimaryImage returns the primary image if loaded, or nil.
func (m Member) PrimaryImage() *MemberImage {
	for i := range m.Images {
		if m.Images[i].IsPrimary {
			return &m.Images[i]
		}
	}
	if len(m.Images) > 0 {
		return &m.Images[0]
	}
	return nil
}


// FullName returns the kanji name formatted as "姓 名".
func (m Member) FullName() string {
	return fmt.Sprintf("%s %s", m.FamilyName, m.GivenName)
}

// FullNameKana returns the furigana kana formatted as "せい めい".
func (m Member) FullNameKana() string {
	return fmt.Sprintf("%s %s", m.FamilyNameKana, m.GivenNameKana)
}

// DisambiguatedLabel returns a distinctive label when multiple members share the exact same name.
// e.g. "加藤 史帆 (1期生)" or "加藤 史帆 (1期生/卒業)"
func (m Member) DisambiguatedLabel() string {
	if m.Status == StatusGraduated {
		return fmt.Sprintf("%s (%d期生/卒業)", m.FullName(), m.Generation)
	}
	return fmt.Sprintf("%s (%d期生)", m.FullName(), m.Generation)
}
