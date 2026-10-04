package model

import "context"

// Repository defines data access operations required by the domain and HTTP handlers.
// Following ADR-0004 and ADR-0017, this interface uses only Go standard library and
// model package types to preserve zero external dependencies in the model layer.
type Repository interface {
	// Master data access
	ListGroups(ctx context.Context) ([]Group, error)
	ListColors(ctx context.Context) ([]Color, error)
	ListMembers(ctx context.Context) ([]Member, error)
	ListMembersByGroup(ctx context.Context, groupID ID) ([]Member, error)
	ListMemberImages(ctx context.Context, memberID ID) ([]MemberImage, error)
	GetMasterVersion(ctx context.Context) (*MasterVersion, error)

	// Answer log operations
	InsertAnswerLog(ctx context.Context, log AnswerLog) error
	BatchInsertAnswerLogs(ctx context.Context, logs []AnswerLog) error
}
