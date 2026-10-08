package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/aobaiwaki/penlight-v2/pkg/model"
)

// sqlQueryer is implemented by both *sql.Conn and *sql.Tx. Keeping the
// scanner against this small interface lets proposal approval read and write
// through the same transaction without introducing repository DTOs.
type sqlQueryer interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type sqlScanner interface {
	Scan(...any) error
}

func (r *SQLiteRepository) withImmediateProposalTransaction(ctx context.Context, fn func(*sql.Conn) (model.MetadataEditProposal, error)) (model.MetadataEditProposal, error) {
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return model.MetadataEditProposal{}, fmt.Errorf("failed to acquire sqlite connection: %w", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, "BEGIN IMMEDIATE;"); err != nil {
		return model.MetadataEditProposal{}, fmt.Errorf("failed to begin immediate transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = conn.ExecContext(context.Background(), "ROLLBACK;")
		}
	}()

	proposal, err := fn(conn)
	if err != nil {
		return model.MetadataEditProposal{}, err
	}
	if _, err := conn.ExecContext(ctx, "COMMIT;"); err != nil {
		return model.MetadataEditProposal{}, fmt.Errorf("failed to commit metadata proposal transaction: %w", err)
	}
	committed = true
	return proposal, nil
}

func parseMetadataProposalTime(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid metadata proposal timestamp %q: %w", value, err)
	}
	return parsed, nil
}

func scanMetadataEditProposal(scanner sqlScanner) (*model.MetadataEditProposal, error) {
	var proposal model.MetadataEditProposal
	var changesJSON string
	var status string
	var submittedAt string
	var approvedAt, rejectedAt, proposerUserID, approverUserID, rejectionReason sql.NullString
	if err := scanner.Scan(
		&proposal.ID,
		&proposal.MemberID,
		&proposal.BaseRevision,
		&changesJSON,
		&status,
		&submittedAt,
		&approvedAt,
		&rejectedAt,
		&proposerUserID,
		&approverUserID,
		&rejectionReason,
	); err != nil {
		return nil, err
	}
	proposal.Status = model.MetadataEditProposalStatus(status)
	if err := json.Unmarshal([]byte(changesJSON), &proposal.Changes); err != nil {
		return nil, fmt.Errorf("failed to decode metadata proposal changes: %w", err)
	}
	var err error
	if proposal.SubmittedAt, err = parseMetadataProposalTime(submittedAt); err != nil {
		return nil, err
	}
	if approvedAt.Valid {
		value, parseErr := parseMetadataProposalTime(approvedAt.String)
		if parseErr != nil {
			return nil, parseErr
		}
		proposal.ApprovedAt = &value
	}
	if rejectedAt.Valid {
		value, parseErr := parseMetadataProposalTime(rejectedAt.String)
		if parseErr != nil {
			return nil, parseErr
		}
		proposal.RejectedAt = &value
	}
	if proposerUserID.Valid {
		value := model.ID(proposerUserID.String)
		proposal.ProposerUserID = &value
	}
	if approverUserID.Valid {
		value := model.ID(approverUserID.String)
		proposal.ApproverUserID = &value
	}
	if rejectionReason.Valid {
		proposal.RejectionReason = &rejectionReason.String
	}
	return &proposal, nil
}

const metadataProposalSelect = `
	SELECT id, member_id, base_revision, changes_json, status, submitted_at,
	       approved_at, rejected_at, proposer_user_id, approver_user_id, rejection_reason
	FROM metadata_edit_proposals`

func loadMetadataEditProposal(ctx context.Context, q sqlQueryer, id model.ID) (*model.MetadataEditProposal, error) {
	return scanMetadataEditProposal(q.QueryRowContext(ctx, metadataProposalSelect+` WHERE id = ?;`, string(id)))
}

func proposalIdentityMatches(existing model.MetadataEditProposal, requested model.MetadataEditProposal) (bool, error) {
	if existing.MemberID != requested.MemberID || existing.BaseRevision != requested.BaseRevision {
		return false, nil
	}
	existingJSON, err := existing.Changes.CanonicalJSON()
	if err != nil {
		return false, fmt.Errorf("failed to canonicalize stored proposal changes: %w", err)
	}
	requestedJSON, err := requested.Changes.CanonicalJSON()
	if err != nil {
		return false, fmt.Errorf("failed to canonicalize requested proposal changes: %w", err)
	}
	return string(existingJSON) == string(requestedJSON), nil
}

func queryMemberMetadata(ctx context.Context, q sqlQueryer, id model.ID) (*model.Member, error) {
	const query = `
		SELECT m.id, m.group_id, m.family_name, m.given_name, m.family_name_kana, m.given_name_kana,
		       m.generation, m.status, m.left_color_id, m.right_color_id, m.ordered,
		       m.joined_at, m.graduated_at, m.verified_at, m.created_at, m.updated_at, m.metadata_revision
		FROM members m
		WHERE m.id = ?;`
	var member model.Member
	var ordered int
	var joinedAt, graduatedAt, verifiedAt sql.NullString
	var createdAt, updatedAt string
	row := q.QueryRowContext(ctx, query, string(id))
	if err := row.Scan(
		&member.ID,
		&member.GroupID,
		&member.FamilyName,
		&member.GivenName,
		&member.FamilyNameKana,
		&member.GivenNameKana,
		&member.Generation,
		&member.Status,
		&member.Penlight.LeftColorID,
		&member.Penlight.RightColorID,
		&ordered,
		&joinedAt,
		&graduatedAt,
		&verifiedAt,
		&createdAt,
		&updatedAt,
		&member.MetadataRevision,
	); err != nil {
		return nil, err
	}
	member.Penlight.Ordered = ordered == 1
	if joinedAt.Valid {
		value, err := time.Parse(time.RFC3339, joinedAt.String)
		if err != nil {
			return nil, fmt.Errorf("failed to parse member joined_at: %w", err)
		}
		member.JoinedAt = &value
	}
	if graduatedAt.Valid {
		value, err := time.Parse(time.RFC3339, graduatedAt.String)
		if err != nil {
			return nil, fmt.Errorf("failed to parse member graduated_at: %w", err)
		}
		member.GraduatedAt = &value
	}
	if verifiedAt.Valid {
		value, err := time.Parse(time.RFC3339, verifiedAt.String)
		if err != nil {
			return nil, fmt.Errorf("failed to parse member verified_at: %w", err)
		}
		member.VerifiedAt = &value
	}
	var err error
	if member.CreatedAt, err = time.Parse(time.RFC3339, createdAt); err != nil {
		return nil, fmt.Errorf("failed to parse member created_at: %w", err)
	}
	if member.UpdatedAt, err = time.Parse(time.RFC3339, updatedAt); err != nil {
		return nil, fmt.Errorf("failed to parse member updated_at: %w", err)
	}
	images, err := queryMemberImages(ctx, q, id)
	if err != nil {
		return nil, err
	}
	member.Images = images
	return &member, nil
}

func queryMemberImages(ctx context.Context, q sqlQueryer, memberID model.ID) ([]model.MemberImage, error) {
	const query = `
		SELECT mi.id, mi.member_id, mi.photo_type_id, mi.image_key, mi.is_primary, mi.display_order,
		       mi.created_at, mi.updated_at,
		       pt.id, pt.group_id, pt.slug, pt.name, pt.display_order, pt.created_at, pt.updated_at
		FROM member_images mi
		JOIN photo_types pt ON mi.photo_type_id = pt.id
		WHERE mi.member_id = ?
		ORDER BY mi.display_order ASC, mi.created_at ASC;`
	rows, err := q.QueryContext(ctx, query, string(memberID))
	if err != nil {
		return nil, fmt.Errorf("failed to query member images: %w", err)
	}
	defer rows.Close()

	images := make([]model.MemberImage, 0)
	for rows.Next() {
		var image model.MemberImage
		var photoType model.PhotoType
		var isPrimary int
		var imageCreatedAt, imageUpdatedAt, photoTypeCreatedAt, photoTypeUpdatedAt string
		if err := rows.Scan(
			&image.ID,
			&image.MemberID,
			&image.PhotoTypeID,
			&image.ImageKey,
			&isPrimary,
			&image.DisplayOrder,
			&imageCreatedAt,
			&imageUpdatedAt,
			&photoType.ID,
			&photoType.GroupID,
			&photoType.Slug,
			&photoType.Name,
			&photoType.DisplayOrder,
			&photoTypeCreatedAt,
			&photoTypeUpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan member image: %w", err)
		}
		image.IsPrimary = isPrimary == 1
		image.CreatedAt, err = time.Parse(time.RFC3339, imageCreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to parse image created_at: %w", err)
		}
		image.UpdatedAt, err = time.Parse(time.RFC3339, imageUpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to parse image updated_at: %w", err)
		}
		photoType.CreatedAt, err = time.Parse(time.RFC3339, photoTypeCreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to parse photo type created_at: %w", err)
		}
		photoType.UpdatedAt, err = time.Parse(time.RFC3339, photoTypeUpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to parse photo type updated_at: %w", err)
		}
		image.PhotoType = &photoType
		images = append(images, image)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error in member images: %w", err)
	}
	return images, nil
}

func validateProposalReferences(ctx context.Context, q sqlQueryer, member model.Member, changes model.MetadataEditChanges) error {
	if changes.Penlight != nil {
		for _, colorID := range []model.ID{changes.Penlight.After.LeftColorID, changes.Penlight.After.RightColorID} {
			var count int
			if err := q.QueryRowContext(ctx, `
				SELECT COUNT(*) FROM colors
				WHERE id = ? AND (group_id IS NULL OR group_id = ?);`, string(colorID), string(member.GroupID)).Scan(&count); err != nil {
				return fmt.Errorf("failed to validate color %s: %w", colorID, err)
			}
			if count != 1 {
				return fmt.Errorf("color %s is not available for member group", colorID)
			}
		}
	}
	if changes.PrimaryImageID != nil {
		var count int
		if err := q.QueryRowContext(ctx, `SELECT COUNT(*) FROM member_images WHERE id = ? AND member_id = ?;`, string(changes.PrimaryImageID.After), string(member.ID)).Scan(&count); err != nil {
			return fmt.Errorf("failed to validate primary image: %w", err)
		}
		if count != 1 {
			return fmt.Errorf("primary image %s does not belong to member", changes.PrimaryImageID.After)
		}
	}
	for _, change := range changes.ImagePhotoTypes {
		var imageCount int
		if err := q.QueryRowContext(ctx, `SELECT COUNT(*) FROM member_images WHERE id = ? AND member_id = ?;`, string(change.ImageID), string(member.ID)).Scan(&imageCount); err != nil {
			return fmt.Errorf("failed to validate image %s: %w", change.ImageID, err)
		}
		if imageCount != 1 {
			return fmt.Errorf("image %s does not belong to member", change.ImageID)
		}
		var photoTypeGroup model.ID
		if err := q.QueryRowContext(ctx, `SELECT group_id FROM photo_types WHERE id = ?;`, string(change.After)).Scan(&photoTypeGroup); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("photo type %s does not exist", change.After)
			}
			return fmt.Errorf("failed to validate photo type %s: %w", change.After, err)
		}
		if photoTypeGroup != member.GroupID {
			return fmt.Errorf("photo type %s is not available for member group", change.After)
		}
	}
	return nil
}

func (r *SQLiteRepository) CreateMetadataEditProposal(ctx context.Context, requested model.MetadataEditProposal) (*model.MetadataEditProposal, error) {
	if requested.Status == "" {
		requested.Status = model.ProposalPending
	}
	if err := requested.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", model.ErrMetadataProposalInvalid, err)
	}
	result, err := r.withImmediateProposalTransaction(ctx, func(conn *sql.Conn) (model.MetadataEditProposal, error) {
		existing, err := loadMetadataEditProposal(ctx, conn, requested.ID)
		if err == nil {
			matches, compareErr := proposalIdentityMatches(*existing, requested)
			if compareErr != nil {
				return model.MetadataEditProposal{}, compareErr
			}
			if !matches {
				return model.MetadataEditProposal{}, model.ErrMetadataProposalContentMismatch
			}
			return *existing, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return model.MetadataEditProposal{}, fmt.Errorf("failed to check proposal idempotency key: %w", err)
		}

		member, err := queryMemberMetadata(ctx, conn, requested.MemberID)
		if errors.Is(err, sql.ErrNoRows) {
			return model.MetadataEditProposal{}, sql.ErrNoRows
		}
		if err != nil {
			return model.MetadataEditProposal{}, fmt.Errorf("failed to load member for proposal: %w", err)
		}
		if requested.BaseRevision != member.MetadataRevision {
			return model.MetadataEditProposal{}, fmt.Errorf("base revision %d does not match current revision %d: %w", requested.BaseRevision, member.MetadataRevision, model.ErrMetadataProposalConflict)
		}
		if err := requested.Changes.ValidateAgainst(*member); err != nil {
			if !errors.Is(err, model.ErrMetadataProposalConflict) {
				return model.MetadataEditProposal{}, fmt.Errorf("%w: %v", model.ErrMetadataProposalInvalid, err)
			}
			return model.MetadataEditProposal{}, err
		}
		if err := validateProposalReferences(ctx, conn, *member, requested.Changes); err != nil {
			return model.MetadataEditProposal{}, fmt.Errorf("%w: %v", model.ErrMetadataProposalInvalid, err)
		}

		changesJSON, err := requested.Changes.CanonicalJSON()
		if err != nil {
			return model.MetadataEditProposal{}, fmt.Errorf("failed to encode proposal changes: %w", err)
		}
		submittedAt := time.Now().UTC()
		if requested.Status != model.ProposalPending {
			return model.MetadataEditProposal{}, fmt.Errorf("%w: new proposal status must be pending", model.ErrMetadataProposalInvalid)
		}
		if _, err := conn.ExecContext(ctx, `
			INSERT INTO metadata_edit_proposals
				(id, member_id, base_revision, changes_json, status, submitted_at, proposer_user_id)
			VALUES (?, ?, ?, ?, ?, ?, ?);`,
			string(requested.ID), string(requested.MemberID), requested.BaseRevision, string(changesJSON),
			string(model.ProposalPending), submittedAt.Format(time.RFC3339Nano), nullableID(requested.ProposerUserID)); err != nil {
			return model.MetadataEditProposal{}, fmt.Errorf("failed to insert metadata edit proposal: %w", err)
		}
		created, err := loadMetadataEditProposal(ctx, conn, requested.ID)
		if err != nil {
			return model.MetadataEditProposal{}, fmt.Errorf("failed to reload created metadata edit proposal: %w", err)
		}
		return *created, nil
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func nullableID(id *model.ID) any {
	if id == nil || *id == "" {
		return nil
	}
	return string(*id)
}

func (r *SQLiteRepository) ListMetadataEditProposals(ctx context.Context, status *model.MetadataEditProposalStatus) ([]model.MetadataEditProposal, error) {
	query := metadataProposalSelect
	args := make([]any, 0, 1)
	if status != nil {
		query += " WHERE status = ?"
		args = append(args, string(*status))
	}
	query += " ORDER BY submitted_at ASC, id ASC;"
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list metadata edit proposals: %w", err)
	}
	defer rows.Close()

	proposals := make([]model.MetadataEditProposal, 0)
	for rows.Next() {
		proposal, err := scanMetadataEditProposal(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metadata edit proposal: %w", err)
		}
		proposals = append(proposals, *proposal)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error in metadata edit proposals: %w", err)
	}
	return proposals, nil
}

func applyMetadataChanges(member *model.Member, changes model.MetadataEditChanges) {
	if changes.Penlight != nil {
		member.Penlight = changes.Penlight.After
	}
	if changes.Generation != nil {
		member.Generation = changes.Generation.After
	}
	if changes.Status != nil {
		member.Status = changes.Status.After
	}
	for _, change := range changes.ImagePhotoTypes {
		for i := range member.Images {
			if member.Images[i].ID == change.ImageID {
				member.Images[i].PhotoTypeID = change.After
			}
		}
	}
}

func (r *SQLiteRepository) ApproveMetadataEditProposal(ctx context.Context, id model.ID, approverUserID *model.ID) (*model.MetadataEditProposal, error) {
	result, err := r.withImmediateProposalTransaction(ctx, func(conn *sql.Conn) (model.MetadataEditProposal, error) {
		proposal, err := loadMetadataEditProposal(ctx, conn, id)
		if errors.Is(err, sql.ErrNoRows) {
			return model.MetadataEditProposal{}, sql.ErrNoRows
		}
		if err != nil {
			return model.MetadataEditProposal{}, fmt.Errorf("failed to load metadata edit proposal: %w", err)
		}
		if proposal.Status == model.ProposalApproved {
			return *proposal, nil
		}
		if proposal.Status == model.ProposalRejected {
			return model.MetadataEditProposal{}, model.ErrMetadataProposalInvalidState
		}

		member, err := queryMemberMetadata(ctx, conn, proposal.MemberID)
		if errors.Is(err, sql.ErrNoRows) {
			return model.MetadataEditProposal{}, sql.ErrNoRows
		}
		if err != nil {
			return model.MetadataEditProposal{}, fmt.Errorf("failed to load member for approval: %w", err)
		}
		if member.MetadataRevision != proposal.BaseRevision {
			return model.MetadataEditProposal{}, fmt.Errorf("member revision changed from %d to %d: %w", proposal.BaseRevision, member.MetadataRevision, model.ErrMetadataProposalConflict)
		}
		if err := proposal.Changes.ValidateAgainst(*member); err != nil {
			if !errors.Is(err, model.ErrMetadataProposalConflict) {
				return model.MetadataEditProposal{}, fmt.Errorf("%w: %v", model.ErrMetadataProposalInvalid, err)
			}
			return model.MetadataEditProposal{}, err
		}
		if err := validateProposalReferences(ctx, conn, *member, proposal.Changes); err != nil {
			return model.MetadataEditProposal{}, fmt.Errorf("%w: %v", model.ErrMetadataProposalInvalid, err)
		}

		updatedMember := *member
		applyMetadataChanges(&updatedMember, proposal.Changes)
		updatedMember.MetadataRevision++
		now := time.Now().UTC()
		nowString := now.Format(time.RFC3339Nano)
		ordered := 0
		if updatedMember.Penlight.Ordered {
			ordered = 1
		}
		result, err := conn.ExecContext(ctx, `
			UPDATE members
			SET left_color_id = ?, right_color_id = ?, ordered = ?, generation = ?, status = ?,
			    metadata_revision = ?, verified_at = ?, updated_at = ?
			WHERE id = ? AND metadata_revision = ?;`,
			string(updatedMember.Penlight.LeftColorID), string(updatedMember.Penlight.RightColorID), ordered,
			updatedMember.Generation, string(updatedMember.Status), updatedMember.MetadataRevision, nowString, nowString,
			string(updatedMember.ID), member.MetadataRevision)
		if err != nil {
			return model.MetadataEditProposal{}, fmt.Errorf("failed to update member during approval: %w", err)
		}
		if affected, affectedErr := result.RowsAffected(); affectedErr != nil || affected != 1 {
			if affectedErr != nil {
				return model.MetadataEditProposal{}, fmt.Errorf("failed to verify member update: %w", affectedErr)
			}
			return model.MetadataEditProposal{}, fmt.Errorf("member revision update affected %d rows", affected)
		}

		if proposal.Changes.PrimaryImageID != nil {
			if _, err := conn.ExecContext(ctx, `UPDATE member_images SET is_primary = 0, updated_at = ? WHERE member_id = ?;`, nowString, string(member.ID)); err != nil {
				return model.MetadataEditProposal{}, fmt.Errorf("failed to reset primary images: %w", err)
			}
			if _, err := conn.ExecContext(ctx, `UPDATE member_images SET is_primary = 1, updated_at = ? WHERE id = ? AND member_id = ?;`, nowString, string(proposal.Changes.PrimaryImageID.After), string(member.ID)); err != nil {
				return model.MetadataEditProposal{}, fmt.Errorf("failed to set primary image: %w", err)
			}
		}
		for _, change := range proposal.Changes.ImagePhotoTypes {
			result, err := conn.ExecContext(ctx, `
				UPDATE member_images SET photo_type_id = ?, updated_at = ?
				WHERE id = ? AND member_id = ?;`, string(change.After), nowString, string(change.ImageID), string(member.ID))
			if err != nil {
				return model.MetadataEditProposal{}, fmt.Errorf("failed to update image photo type: %w", err)
			}
			if affected, affectedErr := result.RowsAffected(); affectedErr != nil || affected != 1 {
				if affectedErr != nil {
					return model.MetadataEditProposal{}, fmt.Errorf("failed to verify image photo type update: %w", affectedErr)
				}
				return model.MetadataEditProposal{}, fmt.Errorf("image photo type update affected %d rows", affected)
			}
		}

		result, err = conn.ExecContext(ctx, `
			UPDATE metadata_edit_proposals
			SET status = ?, approved_at = ?, approver_user_id = ?
			WHERE id = ? AND status = ?;`,
			string(model.ProposalApproved), nowString, nullableID(approverUserID), string(id), string(model.ProposalPending))
		if err != nil {
			return model.MetadataEditProposal{}, fmt.Errorf("failed to mark proposal approved: %w", err)
		}
		if affected, affectedErr := result.RowsAffected(); affectedErr != nil || affected != 1 {
			if affectedErr != nil {
				return model.MetadataEditProposal{}, fmt.Errorf("failed to verify proposal approval: %w", affectedErr)
			}
			return model.MetadataEditProposal{}, fmt.Errorf("proposal approval affected %d rows", affected)
		}

		result, err = conn.ExecContext(ctx, `UPDATE master_versions SET data_revision = data_revision + 1, updated_at = ? WHERE id = 'current';`, nowString)
		if err != nil {
			return model.MetadataEditProposal{}, fmt.Errorf("failed to increment master data revision: %w", err)
		}
		if affected, affectedErr := result.RowsAffected(); affectedErr != nil || affected != 1 {
			if affectedErr != nil {
				return model.MetadataEditProposal{}, fmt.Errorf("failed to verify master data revision update: %w", affectedErr)
			}
			return model.MetadataEditProposal{}, fmt.Errorf("master data revision update affected %d rows", affected)
		}

		approved, err := loadMetadataEditProposal(ctx, conn, id)
		if err != nil {
			return model.MetadataEditProposal{}, fmt.Errorf("failed to reload approved proposal: %w", err)
		}
		return *approved, nil
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *SQLiteRepository) RejectMetadataEditProposal(ctx context.Context, id model.ID, approverUserID *model.ID, reason *string) (*model.MetadataEditProposal, error) {
	result, err := r.withImmediateProposalTransaction(ctx, func(conn *sql.Conn) (model.MetadataEditProposal, error) {
		proposal, err := loadMetadataEditProposal(ctx, conn, id)
		if errors.Is(err, sql.ErrNoRows) {
			return model.MetadataEditProposal{}, sql.ErrNoRows
		}
		if err != nil {
			return model.MetadataEditProposal{}, fmt.Errorf("failed to load metadata edit proposal: %w", err)
		}
		if proposal.Status == model.ProposalRejected {
			return *proposal, nil
		}
		if proposal.Status == model.ProposalApproved {
			return model.MetadataEditProposal{}, model.ErrMetadataProposalInvalidState
		}

		now := time.Now().UTC().Format(time.RFC3339Nano)
		result, err := conn.ExecContext(ctx, `
			UPDATE metadata_edit_proposals
			SET status = ?, rejected_at = ?, approver_user_id = ?, rejection_reason = ?
			WHERE id = ? AND status = ?;`,
			string(model.ProposalRejected), now, nullableID(approverUserID), nullableString(reason), string(id), string(model.ProposalPending))
		if err != nil {
			return model.MetadataEditProposal{}, fmt.Errorf("failed to mark proposal rejected: %w", err)
		}
		if affected, affectedErr := result.RowsAffected(); affectedErr != nil || affected != 1 {
			if affectedErr != nil {
				return model.MetadataEditProposal{}, fmt.Errorf("failed to verify proposal rejection: %w", affectedErr)
			}
			return model.MetadataEditProposal{}, fmt.Errorf("proposal rejection affected %d rows", affected)
		}
		rejected, err := loadMetadataEditProposal(ctx, conn, id)
		if err != nil {
			return model.MetadataEditProposal{}, fmt.Errorf("failed to reload rejected proposal: %w", err)
		}
		return *rejected, nil
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func nullableString(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}
