package model

import "context"

// Repository defines data access operations required by the domain and HTTP handlers.
// Following ADR-0004 and ADR-0017, this interface uses only Go standard library and
// model package types to preserve zero external dependencies in the model layer.
type Repository interface {
	// Series operations (Ref: ADR-0026)
	ListSeries(ctx context.Context) ([]Series, error)
	GetSeries(ctx context.Context, id ID) (*Series, error)

	// Master data access
	ListGroups(ctx context.Context) ([]Group, error)
	ListGroupsBySeries(ctx context.Context, seriesID ID) ([]Group, error)
	ListColors(ctx context.Context) ([]Color, error)
	ListPhotoTypes(ctx context.Context, groupID ID) ([]PhotoType, error)
	ListMembers(ctx context.Context, options ...MemberListOptions) ([]Member, error)
	ListMembersByGroup(ctx context.Context, groupID ID) ([]Member, error)
	ListMembersBySeries(ctx context.Context, seriesID ID) ([]Member, error)
	ListMemberImages(ctx context.Context, memberID ID) ([]MemberImage, error)
	GetMasterVersion(ctx context.Context) (*MasterVersion, error)

	// Song operations (Ref: ADR-0027)
	ListSongs(ctx context.Context) ([]Song, error)
	ListSongsByGroup(ctx context.Context, groupID ID) ([]Song, error)
	ListSongsBySeries(ctx context.Context, seriesID ID) ([]Song, error)

	// Answer log operations
	InsertAnswerLog(ctx context.Context, log AnswerLog) error
	BatchInsertAnswerLogs(ctx context.Context, logs []AnswerLog) error
	GetQuizStatistics(ctx context.Context, filter QuizStatisticsFilter) (*QuizStatisticsResponse, error)

	// Member lookups and metadata edit proposals (Ref: ADR-0017, ADR-0037)
	GetMember(ctx context.Context, id ID) (*Member, error)
	CreateMetadataEditProposal(ctx context.Context, proposal MetadataEditProposal) (*MetadataEditProposal, error)
	ListMetadataEditProposals(ctx context.Context, status *MetadataEditProposalStatus) ([]MetadataEditProposal, error)
	ApproveMetadataEditProposal(ctx context.Context, id ID, approverUserID *ID) (*MetadataEditProposal, error)
	RejectMetadataEditProposal(ctx context.Context, id ID, approverUserID *ID, reason *string) (*MetadataEditProposal, error)
}
