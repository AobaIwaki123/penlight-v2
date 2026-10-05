package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aobaiwaki/penlight-v2/pkg/model"
	_ "modernc.org/sqlite" // Pure Go SQLite driver (CGO_ENABLED=0 compatible, Ref: ADR-0005)
)

var _ model.Repository = (*SQLiteRepository)(nil)

// SQLiteRepository implements model.Repository using modernc.org/sqlite in WAL mode.
type SQLiteRepository struct {
	db *sql.DB
}

// NewSQLiteRepository opens or creates the SQLite database at dbPath and applies pragmatic defaults.
func NewSQLiteRepository(dbPath string) (*SQLiteRepository, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// SQLite connection string with WAL mode, 5s busy timeout, and foreign keys (Ref: ADR-0005)
	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// Connection pool tuning for SQLite single-writer semantics
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping sqlite database: %w", err)
	}

	return &SQLiteRepository{db: db}, nil
}

// DB returns the underlying sql.DB instance for testing or transaction management.
func (r *SQLiteRepository) DB() *sql.DB {
	return r.db
}

// Close closes the underlying database handle.
func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}

// ListGroups returns all active groups ordered by display_order.
func (r *SQLiteRepository) ListGroups(ctx context.Context) ([]model.Group, error) {
	const query = `
		SELECT id, name, short_name, slug, theme_color_hex, display_order, is_active, created_at, updated_at
		FROM groups
		WHERE is_active = 1
		ORDER BY display_order ASC;
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query groups: %w", err)
	}
	defer rows.Close()

	var groups []model.Group
	for rows.Next() {
		var g model.Group
		var isActiveInt int
		var createdAtStr, updatedAtStr string

		if err := rows.Scan(
			&g.ID,
			&g.Name,
			&g.ShortName,
			&g.Slug,
			&g.ThemeColorHex,
			&g.DisplayOrder,
			&isActiveInt,
			&createdAtStr,
			&updatedAtStr,
		); err != nil {
			return nil, fmt.Errorf("failed to scan group: %w", err)
		}

		g.IsActive = isActiveInt == 1
		g.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		g.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
		groups = append(groups, g)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error in groups: %w", err)
	}
	return groups, nil
}

// ListColors returns all official colors ordered by display_order.
func (r *SQLiteRepository) ListColors(ctx context.Context) ([]model.Color, error) {
	const query = `
		SELECT id, group_id, name, hex_code, display_order, created_at, updated_at
		FROM colors
		ORDER BY display_order ASC;
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query colors: %w", err)
	}
	defer rows.Close()

	var colors []model.Color
	for rows.Next() {
		var c model.Color
		var groupID sql.NullString
		var createdAtStr, updatedAtStr string

		if err := rows.Scan(
			&c.ID,
			&groupID,
			&c.Name,
			&c.HexCode,
			&c.DisplayOrder,
			&createdAtStr,
			&updatedAtStr,
		); err != nil {
			return nil, fmt.Errorf("failed to scan color: %w", err)
		}

		if groupID.Valid {
			gid := model.ID(groupID.String)
			c.GroupID = &gid
		}
		c.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		c.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
		colors = append(colors, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error in colors: %w", err)
	}
	return colors, nil
}

// ListPhotoTypes returns all costume/photo categories for a group (Ref: ADR-0021).
func (r *SQLiteRepository) ListPhotoTypes(ctx context.Context, groupID model.ID) ([]model.PhotoType, error) {
	const query = `
		SELECT id, group_id, slug, name, display_order, created_at, updated_at
		FROM photo_types
		WHERE (? = '' OR group_id = ?)
		ORDER BY display_order ASC;
	`
	rows, err := r.db.QueryContext(ctx, query, string(groupID), string(groupID))
	if err != nil {
		return nil, fmt.Errorf("failed to query photo types: %w", err)
	}
	defer rows.Close()

	var photoTypes []model.PhotoType
	for rows.Next() {
		var pt model.PhotoType
		var createdAtStr, updatedAtStr string

		if err := rows.Scan(
			&pt.ID,
			&pt.GroupID,
			&pt.Slug,
			&pt.Name,
			&pt.DisplayOrder,
			&createdAtStr,
			&updatedAtStr,
		); err != nil {
			return nil, fmt.Errorf("failed to scan photo type: %w", err)
		}

		pt.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		pt.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
		photoTypes = append(photoTypes, pt)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error in photo types: %w", err)
	}
	return photoTypes, nil
}

// ListMembers returns all non-graduated members with their primary image.
func (r *SQLiteRepository) ListMembers(ctx context.Context) ([]model.Member, error) {
	const query = `
		SELECT m.id, m.group_id, m.family_name, m.given_name, m.family_name_kana, m.given_name_kana,
		       m.generation, m.status, m.left_color_id, m.right_color_id, m.ordered,
		       m.joined_at, m.graduated_at, m.created_at, m.updated_at,
		       mi.id, mi.photo_type_id, mi.image_key, mi.is_primary, mi.display_order, mi.created_at, mi.updated_at,
		       pt.id, pt.group_id, pt.slug, pt.name, pt.display_order, pt.created_at, pt.updated_at
		FROM members m
		LEFT JOIN member_images mi ON m.id = mi.member_id AND mi.is_primary = 1
		LEFT JOIN photo_types pt ON mi.photo_type_id = pt.id
		WHERE m.status != 'graduated'
		ORDER BY m.generation ASC, m.family_name_kana ASC;
	`
	return r.queryMembers(ctx, query)
}

// ListMembersByGroup returns non-graduated members belonging to a specific group with their primary image.
func (r *SQLiteRepository) ListMembersByGroup(ctx context.Context, groupID model.ID) ([]model.Member, error) {
	const query = `
		SELECT m.id, m.group_id, m.family_name, m.given_name, m.family_name_kana, m.given_name_kana,
		       m.generation, m.status, m.left_color_id, m.right_color_id, m.ordered,
		       m.joined_at, m.graduated_at, m.created_at, m.updated_at,
		       mi.id, mi.photo_type_id, mi.image_key, mi.is_primary, mi.display_order, mi.created_at, mi.updated_at,
		       pt.id, pt.group_id, pt.slug, pt.name, pt.display_order, pt.created_at, pt.updated_at
		FROM members m
		LEFT JOIN member_images mi ON m.id = mi.member_id AND mi.is_primary = 1
		LEFT JOIN photo_types pt ON mi.photo_type_id = pt.id
		WHERE m.group_id = ? AND m.status != 'graduated'
		ORDER BY m.generation ASC, m.family_name_kana ASC;
	`
	return r.queryMembers(ctx, query, groupID)
}

// ListMemberImages returns all images associated with a member.
func (r *SQLiteRepository) ListMemberImages(ctx context.Context, memberID model.ID) ([]model.MemberImage, error) {
	const query = `
		SELECT mi.id, mi.member_id, mi.photo_type_id, mi.image_key, mi.is_primary, mi.display_order, mi.created_at, mi.updated_at,
		       pt.id, pt.group_id, pt.slug, pt.name, pt.display_order, pt.created_at, pt.updated_at
		FROM member_images mi
		JOIN photo_types pt ON mi.photo_type_id = pt.id
		WHERE mi.member_id = ?
		ORDER BY mi.display_order ASC, mi.created_at ASC;
	`
	rows, err := r.db.QueryContext(ctx, query, memberID)
	if err != nil {
		return nil, fmt.Errorf("failed to query member images: %w", err)
	}
	defer rows.Close()

	var images []model.MemberImage
	for rows.Next() {
		var img model.MemberImage
		var pt model.PhotoType
		var isPrimaryInt int
		var createdAtStr, updatedAtStr string
		var ptCreatedAtStr, ptUpdatedAtStr string

		if err := rows.Scan(
			&img.ID,
			&img.MemberID,
			&img.PhotoTypeID,
			&img.ImageKey,
			&isPrimaryInt,
			&img.DisplayOrder,
			&createdAtStr,
			&updatedAtStr,
			&pt.ID,
			&pt.GroupID,
			&pt.Slug,
			&pt.Name,
			&pt.DisplayOrder,
			&ptCreatedAtStr,
			&ptUpdatedAtStr,
		); err != nil {
			return nil, fmt.Errorf("failed to scan member image: %w", err)
		}

		img.IsPrimary = isPrimaryInt == 1
		img.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		img.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
		pt.CreatedAt, _ = time.Parse(time.RFC3339, ptCreatedAtStr)
		pt.UpdatedAt, _ = time.Parse(time.RFC3339, ptUpdatedAtStr)
		img.PhotoType = &pt

		images = append(images, img)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error in member images: %w", err)
	}
	return images, nil
}

// GetMasterVersion returns the current master data synchronization version (Ref: ADR-0021).
func (r *SQLiteRepository) GetMasterVersion(ctx context.Context) (*model.MasterVersion, error) {
	const query = `
		SELECT id, version, updated_at
		FROM master_versions
		WHERE id = 'current';
	`
	var mv model.MasterVersion
	var updatedAtStr string
	err := r.db.QueryRowContext(ctx, query).Scan(&mv.ID, &mv.Version, &updatedAtStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query master version: %w", err)
	}
	mv.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
	return &mv, nil
}

func (r *SQLiteRepository) queryMembers(ctx context.Context, query string, args ...any) ([]model.Member, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query members: %w", err)
	}
	defer rows.Close()

	var members []model.Member
	for rows.Next() {
		var m model.Member
		var orderedInt int
		var joinedAtStr, graduatedAtStr sql.NullString
		var createdAtStr, updatedAtStr string
		var imgID, imgPhotoTypeID, imgKey, imgCreatedAtStr, imgUpdatedAtStr sql.NullString
		var imgIsPrimary, imgDisplayOrder sql.NullInt64
		var ptID, ptGroupID, ptSlug, ptName, ptCreatedAtStr, ptUpdatedAtStr sql.NullString
		var ptDisplayOrder sql.NullInt64

		if err := rows.Scan(
			&m.ID,
			&m.GroupID,
			&m.FamilyName,
			&m.GivenName,
			&m.FamilyNameKana,
			&m.GivenNameKana,
			&m.Generation,
			&m.Status,
			&m.Penlight.LeftColorID,
			&m.Penlight.RightColorID,
			&orderedInt,
			&joinedAtStr,
			&graduatedAtStr,
			&createdAtStr,
			&updatedAtStr,
			&imgID,
			&imgPhotoTypeID,
			&imgKey,
			&imgIsPrimary,
			&imgDisplayOrder,
			&imgCreatedAtStr,
			&imgUpdatedAtStr,
			&ptID,
			&ptGroupID,
			&ptSlug,
			&ptName,
			&ptDisplayOrder,
			&ptCreatedAtStr,
			&ptUpdatedAtStr,
		); err != nil {
			return nil, fmt.Errorf("failed to scan member: %w", err)
		}

		m.Penlight.Ordered = orderedInt == 1
		if joinedAtStr.Valid {
			t, _ := time.Parse(time.RFC3339, joinedAtStr.String)
			m.JoinedAt = &t
		}
		if graduatedAtStr.Valid {
			t, _ := time.Parse(time.RFC3339, graduatedAtStr.String)
			m.GraduatedAt = &t
		}
		m.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		m.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)

		if imgID.Valid {
			imgCreated, _ := time.Parse(time.RFC3339, imgCreatedAtStr.String)
			imgUpdated, _ := time.Parse(time.RFC3339, imgUpdatedAtStr.String)
			memberImg := model.MemberImage{
				ID:           model.ID(imgID.String),
				MemberID:     m.ID,
				PhotoTypeID:  model.ID(imgPhotoTypeID.String),
				ImageKey:     imgKey.String,
				IsPrimary:    imgIsPrimary.Int64 == 1,
				DisplayOrder: int(imgDisplayOrder.Int64),
				CreatedAt:    imgCreated,
				UpdatedAt:    imgUpdated,
			}
			if ptID.Valid {
				ptCreated, _ := time.Parse(time.RFC3339, ptCreatedAtStr.String)
				ptUpdated, _ := time.Parse(time.RFC3339, ptUpdatedAtStr.String)
				memberImg.PhotoType = &model.PhotoType{
					ID:           model.ID(ptID.String),
					GroupID:      model.ID(ptGroupID.String),
					Slug:         ptSlug.String,
					Name:         ptName.String,
					DisplayOrder: int(ptDisplayOrder.Int64),
					CreatedAt:    ptCreated,
					UpdatedAt:    ptUpdated,
				}
			}
			m.Images = append(m.Images, memberImg)
		}

		members = append(members, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error in members: %w", err)
	}
	return members, nil
}



// InsertAnswerLog inserts an individual quiz answer record idempotently (Ref: ADR-0007, ADR-0032).
func (r *SQLiteRepository) InsertAnswerLog(ctx context.Context, log model.AnswerLog) error {
	const query = `
		INSERT OR IGNORE INTO answer_logs (
			id, user_id, quiz_question_id, target_member_id, target_song_id, group_id, is_correct, response_time_ms, answered_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);
	`
	isCorrectInt := 0
	if log.IsCorrect {
		isCorrectInt = 1
	}

	var userIDArg any = nil
	if log.UserID != "" {
		userIDArg = string(log.UserID)
	}

	var memberIDArg any = nil
	if log.TargetMemberID != nil {
		memberIDArg = string(*log.TargetMemberID)
	}
	var songIDArg any = nil
	if log.TargetSongID != nil {
		songIDArg = string(*log.TargetSongID)
	}

	_, err := r.db.ExecContext(
		ctx,
		query,
		string(log.ID),
		userIDArg,
		string(log.QuizQuestionID),
		memberIDArg,
		songIDArg,
		string(log.GroupID),
		isCorrectInt,
		log.ResponseTimeMs,
		log.AnsweredAt.Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("failed to insert answer log: %w", err)
	}
	return nil
}

// BatchInsertAnswerLogs inserts multiple answer logs in a single atomic transaction.
func (r *SQLiteRepository) BatchInsertAnswerLogs(ctx context.Context, logs []model.AnswerLog) error {
	if len(logs) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	const query = `
		INSERT OR IGNORE INTO answer_logs (
			id, user_id, quiz_question_id, target_member_id, target_song_id, group_id, is_correct, response_time_ms, answered_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);
	`
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, log := range logs {
		isCorrectInt := 0
		if log.IsCorrect {
			isCorrectInt = 1
		}
		var userIDArg any = nil
		if log.UserID != "" {
			userIDArg = string(log.UserID)
		}
		var memberIDArg any = nil
		if log.TargetMemberID != nil {
			memberIDArg = string(*log.TargetMemberID)
		}
		var songIDArg any = nil
		if log.TargetSongID != nil {
			songIDArg = string(*log.TargetSongID)
		}

		if _, err := stmt.ExecContext(
			ctx,
			string(log.ID),
			userIDArg,
			string(log.QuizQuestionID),
			memberIDArg,
			songIDArg,
			string(log.GroupID),
			isCorrectInt,
			log.ResponseTimeMs,
			log.AnsweredAt.Format(time.RFC3339),
		); err != nil {
			return fmt.Errorf("failed to insert answer log %s: %w", log.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit batch answer logs: %w", err)
	}
	return nil
}
