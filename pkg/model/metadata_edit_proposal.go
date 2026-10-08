package model

import "time"

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
