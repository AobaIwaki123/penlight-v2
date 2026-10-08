package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

var (
	// ErrMetadataProposalConflict means the proposal's before values or base
	// revision no longer match the public member.
	ErrMetadataProposalConflict = errors.New("metadata edit proposal conflicts with current member")
	// ErrMetadataProposalContentMismatch means an idempotency key was reused for
	// different proposal content.
	ErrMetadataProposalContentMismatch = errors.New("metadata edit proposal content mismatch")
	// ErrMetadataProposalInvalidState means a terminal proposal was asked to move
	// to the opposite terminal state.
	ErrMetadataProposalInvalidState = errors.New("metadata edit proposal has an invalid state transition")
	// ErrMetadataProposalInvalid means a typed proposal payload or one of its
	// references is invalid and should be reported as INVALID_PARAMS.
	ErrMetadataProposalInvalid = errors.New("metadata edit proposal is invalid")
)

// MetadataEditProposalStatus is the immutable decision state (Ref: ADR-0037).
type MetadataEditProposalStatus string

const (
	ProposalPending  MetadataEditProposalStatus = "pending"
	ProposalApproved MetadataEditProposalStatus = "approved"
	ProposalRejected MetadataEditProposalStatus = "rejected"
)

// PenlightChange records the complete color pair before and after an edit.
type PenlightChange struct {
	Before PenlightPair `json:"before" db:"-"`
	After  PenlightPair `json:"after" db:"-"`
}

type GenerationChange struct {
	Before int `json:"before" db:"-"`
	After  int `json:"after" db:"-"`
}

type MemberStatusChange struct {
	Before MemberStatus `json:"before" db:"-"`
	After  MemberStatus `json:"after" db:"-"`
}

// PrimaryImageChange permits an initially unset primary image.
type PrimaryImageChange struct {
	Before *ID `json:"before" db:"-"`
	After  ID  `json:"after" db:"-"`
}

type ImagePhotoTypeChange struct {
	ImageID ID `json:"image_id" db:"-"`
	Before  ID `json:"before" db:"-"`
	After   ID `json:"after" db:"-"`
}

// MetadataEditChanges contains only the fields allowed by ADR-0036/0037.
type MetadataEditChanges struct {
	Penlight        *PenlightChange        `json:"penlight,omitempty" db:"-"`
	Generation      *GenerationChange      `json:"generation,omitempty" db:"-"`
	Status          *MemberStatusChange    `json:"status,omitempty" db:"-"`
	PrimaryImageID  *PrimaryImageChange    `json:"primary_image_id,omitempty" db:"-"`
	ImagePhotoTypes []ImagePhotoTypeChange `json:"image_photo_types,omitempty" db:"-"`
}

// SubmitMetadataEditProposalRequest is the user-facing proposal submission
// payload. The member is identified by the URL path.
type SubmitMetadataEditProposalRequest struct {
	ID           ID                  `json:"id"`
	BaseRevision int                 `json:"base_revision"`
	Changes      MetadataEditChanges `json:"changes"`
}

// RejectMetadataEditProposalRequest is the optional rejection reason payload.
type RejectMetadataEditProposalRequest struct {
	Reason *string `json:"reason,omitempty"`
}

// Normalize returns a copy with order-insensitive image changes sorted by
// image ID. All other fields retain their typed representation.
func (c MetadataEditChanges) Normalize() MetadataEditChanges {
	n := c
	if c.Penlight != nil {
		v := *c.Penlight
		n.Penlight = &v
	}
	if c.Generation != nil {
		v := *c.Generation
		n.Generation = &v
	}
	if c.Status != nil {
		v := *c.Status
		n.Status = &v
	}
	if c.PrimaryImageID != nil {
		v := *c.PrimaryImageID
		if c.PrimaryImageID.Before != nil {
			before := *c.PrimaryImageID.Before
			v.Before = &before
		}
		n.PrimaryImageID = &v
	}
	n.ImagePhotoTypes = append([]ImagePhotoTypeChange(nil), c.ImagePhotoTypes...)
	sort.Slice(n.ImagePhotoTypes, func(i, j int) bool {
		return n.ImagePhotoTypes[i].ImageID < n.ImagePhotoTypes[j].ImageID
	})
	return n
}

// CanonicalJSON returns the stable JSON representation used for proposal
// idempotency comparisons.
func (c MetadataEditChanges) CanonicalJSON() ([]byte, error) {
	return json.Marshal(c.Normalize())
}

func hasIDPrefix(id ID, prefix Prefix) bool {
	return strings.HasPrefix(string(id), string(prefix)+"_")
}

func validMemberStatus(status MemberStatus) bool {
	return status == StatusActive || status == StatusGraduated || status == StatusHiatus
}

func samePenlight(a, b PenlightPair) bool {
	return a.LeftColorID == b.LeftColorID && a.RightColorID == b.RightColorID && a.Ordered == b.Ordered
}

// Validate checks the typed proposal shape and rejects empty or no-op edits.
func (c MetadataEditChanges) Validate() error {
	changed := 0
	if c.Penlight != nil {
		if !hasIDPrefix(c.Penlight.Before.LeftColorID, PrefixColor) || !hasIDPrefix(c.Penlight.Before.RightColorID, PrefixColor) ||
			!hasIDPrefix(c.Penlight.After.LeftColorID, PrefixColor) || !hasIDPrefix(c.Penlight.After.RightColorID, PrefixColor) {
			return fmt.Errorf("penlight color IDs must use %s_ prefix", PrefixColor)
		}
		if samePenlight(c.Penlight.Before, c.Penlight.After) {
			return errors.New("penlight change is a no-op")
		}
		changed++
	}
	if c.Generation != nil {
		if c.Generation.Before < 1 || c.Generation.After < 1 {
			return errors.New("generation must be positive")
		}
		if c.Generation.Before == c.Generation.After {
			return errors.New("generation change is a no-op")
		}
		changed++
	}
	if c.Status != nil {
		if !validMemberStatus(c.Status.Before) || !validMemberStatus(c.Status.After) {
			return errors.New("status must be active, graduated, or hiatus")
		}
		if c.Status.Before == c.Status.After {
			return errors.New("status change is a no-op")
		}
		changed++
	}
	if c.PrimaryImageID != nil {
		if c.PrimaryImageID.Before != nil && !hasIDPrefix(*c.PrimaryImageID.Before, PrefixImage) {
			return fmt.Errorf("primary image before ID must use %s_ prefix", PrefixImage)
		}
		if !hasIDPrefix(c.PrimaryImageID.After, PrefixImage) {
			return fmt.Errorf("primary image after ID must use %s_ prefix", PrefixImage)
		}
		if c.PrimaryImageID.Before != nil && *c.PrimaryImageID.Before == c.PrimaryImageID.After {
			return errors.New("primary image change is a no-op")
		}
		changed++
	}
	seenImages := make(map[ID]struct{}, len(c.ImagePhotoTypes))
	for _, change := range c.ImagePhotoTypes {
		if !hasIDPrefix(change.ImageID, PrefixImage) || !hasIDPrefix(change.Before, PrefixPhotoType) || !hasIDPrefix(change.After, PrefixPhotoType) {
			return fmt.Errorf("image and photo type IDs must use their TypeID prefixes")
		}
		if _, ok := seenImages[change.ImageID]; ok {
			return fmt.Errorf("image photo type %s is duplicated", change.ImageID)
		}
		seenImages[change.ImageID] = struct{}{}
		if change.Before == change.After {
			return fmt.Errorf("image photo type change for %s is a no-op", change.ImageID)
		}
		changed++
	}
	if changed == 0 {
		return errors.New("metadata edit proposal has no changes")
	}
	return nil
}

func memberPrimaryImageID(member Member) *ID {
	for _, image := range member.Images {
		if image.IsPrimary {
			id := image.ID
			return &id
		}
	}
	return nil
}

// ValidateAgainst verifies that every before value still describes the public
// member. A mismatch is a conflict and must leave the proposal pending.
func (c MetadataEditChanges) ValidateAgainst(member Member) error {
	if err := c.Validate(); err != nil {
		return err
	}
	if c.Penlight != nil && !samePenlight(c.Penlight.Before, member.Penlight) {
		return fmt.Errorf("penlight before value does not match: %w", ErrMetadataProposalConflict)
	}
	if c.Generation != nil && c.Generation.Before != member.Generation {
		return fmt.Errorf("generation before value does not match: %w", ErrMetadataProposalConflict)
	}
	if c.Status != nil && c.Status.Before != member.Status {
		return fmt.Errorf("status before value does not match: %w", ErrMetadataProposalConflict)
	}
	if c.PrimaryImageID != nil {
		current := memberPrimaryImageID(member)
		same := current == nil && c.PrimaryImageID.Before == nil
		if current != nil && c.PrimaryImageID.Before != nil {
			same = *current == *c.PrimaryImageID.Before
		}
		if !same {
			return fmt.Errorf("primary image before value does not match: %w", ErrMetadataProposalConflict)
		}
	}
	for _, change := range c.ImagePhotoTypes {
		var current *MemberImage
		for i := range member.Images {
			if member.Images[i].ID == change.ImageID {
				current = &member.Images[i]
				break
			}
		}
		if current == nil || current.PhotoTypeID != change.Before {
			return fmt.Errorf("photo type before value for %s does not match: %w", change.ImageID, ErrMetadataProposalConflict)
		}
	}
	return nil
}

// MetadataEditProposal retains a submission and its final decision (Ref: ADR-0037).
type MetadataEditProposal struct {
	ID              ID                         `json:"id" db:"id,pk" sql:"TEXT PRIMARY KEY"`
	MemberID        ID                         `json:"member_id" db:"member_id,fk" sql:"TEXT NOT NULL REFERENCES members(id) ON DELETE RESTRICT"`
	BaseRevision    int                        `json:"base_revision" db:"base_revision" sql:"INTEGER NOT NULL CHECK (base_revision >= 1)"`
	Changes         MetadataEditChanges        `json:"changes" db:"changes_json" sql:"TEXT NOT NULL CHECK (json_valid(changes_json))"`
	Status          MetadataEditProposalStatus `json:"status" db:"status" sql:"TEXT NOT NULL"`
	SubmittedAt     time.Time                  `json:"submitted_at" db:"submitted_at" sql:"TEXT NOT NULL"`
	ApprovedAt      *time.Time                 `json:"approved_at,omitempty" db:"approved_at" sql:"TEXT"`
	RejectedAt      *time.Time                 `json:"rejected_at,omitempty" db:"rejected_at" sql:"TEXT"`
	ProposerUserID  *ID                        `json:"proposer_user_id,omitempty" db:"proposer_user_id,fk" sql:"TEXT REFERENCES users(id) ON DELETE SET NULL"`
	ApproverUserID  *ID                        `json:"approver_user_id,omitempty" db:"approver_user_id,fk" sql:"TEXT REFERENCES users(id) ON DELETE SET NULL"`
	RejectionReason *string                    `json:"rejection_reason,omitempty" db:"rejection_reason" sql:"TEXT"`
}

// Validate checks proposal identity and lifecycle fields before persistence.
func (p MetadataEditProposal) Validate() error {
	if !hasIDPrefix(p.ID, PrefixProposal) {
		return fmt.Errorf("proposal ID must use %s_ prefix", PrefixProposal)
	}
	if !hasIDPrefix(p.MemberID, PrefixMember) {
		return fmt.Errorf("member ID must use %s_ prefix", PrefixMember)
	}
	if p.BaseRevision < 1 {
		return errors.New("base revision must be positive")
	}
	if p.Status != ProposalPending && p.Status != ProposalApproved && p.Status != ProposalRejected {
		return fmt.Errorf("invalid proposal status %q", p.Status)
	}
	return p.Changes.Validate()
}
