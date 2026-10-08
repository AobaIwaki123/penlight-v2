package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// ListSeries returns all idol series ordered by display_order (Ref: ADR-0026).
func (r *SQLiteRepository) ListSeries(ctx context.Context) ([]model.Series, error) {
	const query = `
		SELECT id, name, slug, display_order, created_at, updated_at
		FROM series
		ORDER BY display_order ASC;
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query series: %w", err)
	}
	defer rows.Close()

	var seriesList []model.Series
	for rows.Next() {
		var s model.Series
		var createdAtStr, updatedAtStr string
		if err := rows.Scan(
			&s.ID,
			&s.Name,
			&s.Slug,
			&s.DisplayOrder,
			&createdAtStr,
			&updatedAtStr,
		); err != nil {
			return nil, fmt.Errorf("failed to scan series: %w", err)
		}
		s.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		s.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
		seriesList = append(seriesList, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error in series: %w", err)
	}
	return seriesList, nil
}

// GetSeries returns an idol series by ID (Ref: ADR-0026).
func (r *SQLiteRepository) GetSeries(ctx context.Context, id model.ID) (*model.Series, error) {
	const query = `
		SELECT id, name, slug, display_order, created_at, updated_at
		FROM series
		WHERE id = ?;
	`
	var s model.Series
	var createdAtStr, updatedAtStr string
	err := r.db.QueryRowContext(ctx, query, string(id)).Scan(
		&s.ID,
		&s.Name,
		&s.Slug,
		&s.DisplayOrder,
		&createdAtStr,
		&updatedAtStr,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get series: %w", err)
	}
	s.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	s.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
	return &s, nil
}

// ListGroups returns all active groups ordered by display_order.
func (r *SQLiteRepository) ListGroups(ctx context.Context) ([]model.Group, error) {
	const query = `
		SELECT id, series_id, name, short_name, slug, theme_color_hex, display_order, is_active, created_at, updated_at
		FROM groups
		WHERE is_active = 1
		ORDER BY display_order ASC;
	`
	return r.queryGroups(ctx, query)
}

// ListGroupsBySeries returns active groups belonging to a specific series (Ref: ADR-0026).
func (r *SQLiteRepository) ListGroupsBySeries(ctx context.Context, seriesID model.ID) ([]model.Group, error) {
	const query = `
		SELECT id, series_id, name, short_name, slug, theme_color_hex, display_order, is_active, created_at, updated_at
		FROM groups
		WHERE series_id = ? AND is_active = 1
		ORDER BY display_order ASC;
	`
	return r.queryGroups(ctx, query, string(seriesID))
}

func (r *SQLiteRepository) queryGroups(ctx context.Context, query string, args ...any) ([]model.Group, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query groups: %w", err)
	}
	defer rows.Close()

	var groups []model.Group
	for rows.Next() {
		var g model.Group
		var seriesIDStr sql.NullString
		var isActiveInt int
		var createdAtStr, updatedAtStr string

		if err := rows.Scan(
			&g.ID,
			&seriesIDStr,
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

		if seriesIDStr.Valid {
			g.SeriesID = model.ID(seriesIDStr.String)
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

// ListMembers returns members with their primary image. Graduated members are
// excluded by default to preserve the public quiz bootstrap behavior.
func (r *SQLiteRepository) ListMembers(ctx context.Context, options ...model.MemberListOptions) ([]model.Member, error) {
	includeGraduated := len(options) > 0 && options[0].IncludeGraduated
	query := `
		SELECT m.id, m.group_id, m.family_name, m.given_name, m.family_name_kana, m.given_name_kana,
		       m.generation, m.status, m.left_color_id, m.right_color_id, m.ordered,
		       m.joined_at, m.graduated_at, m.verified_at, m.created_at, m.updated_at, m.metadata_revision,
		       mi.id, mi.photo_type_id, mi.image_key, mi.is_primary, mi.display_order, mi.created_at, mi.updated_at,
		       pt.id, pt.group_id, pt.slug, pt.name, pt.display_order, pt.created_at, pt.updated_at
		FROM members m
		LEFT JOIN member_images mi ON m.id = mi.member_id AND mi.is_primary = 1
		LEFT JOIN photo_types pt ON mi.photo_type_id = pt.id
`
	if !includeGraduated {
		query += "\t\tWHERE m.status != 'graduated'\n"
	}
	query += "\t\tORDER BY m.generation ASC, m.family_name_kana ASC;\n"
	return r.queryMembers(ctx, query)
}

// ListMembersByGroup returns non-graduated members belonging to a specific group with their primary image.
func (r *SQLiteRepository) ListMembersByGroup(ctx context.Context, groupID model.ID) ([]model.Member, error) {
	const query = `
		SELECT m.id, m.group_id, m.family_name, m.given_name, m.family_name_kana, m.given_name_kana,
		       m.generation, m.status, m.left_color_id, m.right_color_id, m.ordered,
		       m.joined_at, m.graduated_at, m.verified_at, m.created_at, m.updated_at, m.metadata_revision,
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

// ListMembersBySeries returns non-graduated members belonging to a specific series with their primary image (Ref: ADR-0026).
func (r *SQLiteRepository) ListMembersBySeries(ctx context.Context, seriesID model.ID) ([]model.Member, error) {
	const query = `
		SELECT m.id, m.group_id, m.family_name, m.given_name, m.family_name_kana, m.given_name_kana,
		       m.generation, m.status, m.left_color_id, m.right_color_id, m.ordered,
		       m.joined_at, m.graduated_at, m.verified_at, m.created_at, m.updated_at, m.metadata_revision,
		       mi.id, mi.photo_type_id, mi.image_key, mi.is_primary, mi.display_order, mi.created_at, mi.updated_at,
		       pt.id, pt.group_id, pt.slug, pt.name, pt.display_order, pt.created_at, pt.updated_at
		FROM members m
		JOIN groups g ON m.group_id = g.id
		LEFT JOIN member_images mi ON m.id = mi.member_id AND mi.is_primary = 1
		LEFT JOIN photo_types pt ON mi.photo_type_id = pt.id
		WHERE g.series_id = ? AND m.status != 'graduated'
		ORDER BY m.generation ASC, m.family_name_kana ASC;
	`
	return r.queryMembers(ctx, query, seriesID)
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
		SELECT id, version, updated_at, data_revision
		FROM master_versions
		WHERE id = 'current';
	`
	var mv model.MasterVersion
	var updatedAtStr string
	err := r.db.QueryRowContext(ctx, query).Scan(&mv.ID, &mv.Version, &updatedAtStr, &mv.DataRevision)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query master version: %w", err)
	}
	mv.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
	return &mv, nil
}

// ListSongs returns all musical tracks ordered by group_id and title (Ref: ADR-0027).
func (r *SQLiteRepository) ListSongs(ctx context.Context) ([]model.Song, error) {
	const query = `
		SELECT id, group_id, title, kana, color1_id, color2_id, created_at, updated_at
		FROM songs
		ORDER BY group_id ASC, title ASC;
	`
	return r.querySongs(ctx, query)
}

// ListSongsByGroup returns musical tracks belonging to a specific group (Ref: ADR-0027).
func (r *SQLiteRepository) ListSongsByGroup(ctx context.Context, groupID model.ID) ([]model.Song, error) {
	const query = `
		SELECT id, group_id, title, kana, color1_id, color2_id, created_at, updated_at
		FROM songs
		WHERE group_id = ?
		ORDER BY title ASC;
	`
	return r.querySongs(ctx, query, string(groupID))
}

// ListSongsBySeries returns musical tracks belonging to groups within a series (Ref: ADR-0026, ADR-0027).
func (r *SQLiteRepository) ListSongsBySeries(ctx context.Context, seriesID model.ID) ([]model.Song, error) {
	const query = `
		SELECT s.id, s.group_id, s.title, s.kana, s.color1_id, s.color2_id, s.created_at, s.updated_at
		FROM songs s
		JOIN groups g ON s.group_id = g.id
		WHERE g.series_id = ?
		ORDER BY g.display_order ASC, s.title ASC;
	`
	return r.querySongs(ctx, query, string(seriesID))
}

func (r *SQLiteRepository) querySongs(ctx context.Context, query string, args ...any) ([]model.Song, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query songs: %w", err)
	}
	defer rows.Close()

	var songs []model.Song
	for rows.Next() {
		var s model.Song
		var kanaStr, color2Str sql.NullString
		var createdAtStr, updatedAtStr string

		if err := rows.Scan(
			&s.ID,
			&s.GroupID,
			&s.Title,
			&kanaStr,
			&s.Color1ID,
			&color2Str,
			&createdAtStr,
			&updatedAtStr,
		); err != nil {
			return nil, fmt.Errorf("failed to scan song: %w", err)
		}

		if kanaStr.Valid {
			s.Kana = &kanaStr.String
		}
		if color2Str.Valid {
			c2 := model.ID(color2Str.String)
			s.Color2ID = &c2
		}
		s.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		s.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
		songs = append(songs, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error in songs: %w", err)
	}
	return songs, nil
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
		var joinedAtStr, graduatedAtStr, verifiedAtStr sql.NullString
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
			&verifiedAtStr,
			&createdAtStr,
			&updatedAtStr,
			&m.MetadataRevision,
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
		if verifiedAtStr.Valid {
			t, _ := time.Parse(time.RFC3339, verifiedAtStr.String)
			m.VerifiedAt = &t
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

// GetQuizStatistics calculates core aggregate statistics and weak targets (Ref: ADR-0032).
func (r *SQLiteRepository) GetQuizStatistics(ctx context.Context, filter model.QuizStatisticsFilter) (*model.QuizStatisticsResponse, error) {
	resp := &model.QuizStatisticsResponse{
		Groups:      make([]model.GroupStat, 0),
		WeakTargets: make([]model.TargetStat, 0),
		Extra:       make(map[string]any),
	}

	// 1. Overall summary
	var overallWhere []string
	var overallArgs []any

	if filter.UserID != nil {
		overallWhere = append(overallWhere, "user_id = ?")
		overallArgs = append(overallArgs, string(*filter.UserID))
	}
	if filter.GroupID != nil {
		overallWhere = append(overallWhere, "group_id = ?")
		overallArgs = append(overallArgs, string(*filter.GroupID))
	}

	whereClause := ""
	if len(overallWhere) > 0 {
		whereClause = "WHERE " + strings.Join(overallWhere, " AND ")
	}

	overallQuery := fmt.Sprintf(`
		SELECT
			COUNT(*) AS total_answers,
			COALESCE(SUM(is_correct), 0) AS total_correct,
			COALESCE(AVG(response_time_ms), 0) AS avg_time
		FROM answer_logs
		%s;
	`, whereClause)

	var avgTimeFloat float64
	err := r.db.QueryRowContext(ctx, overallQuery, overallArgs...).Scan(
		&resp.TotalAnswers,
		&resp.TotalCorrect,
		&avgTimeFloat,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query overall statistics: %w", err)
	}
	resp.AverageResponseTimeMs = int(avgTimeFloat)
	if resp.TotalAnswers > 0 {
		resp.AccuracyRate = float64(resp.TotalCorrect) / float64(resp.TotalAnswers)
	}

	// 2. Group statistics (joined with groups table)
	groupWhere := make([]string, 0)
	groupArgs := make([]any, 0)
	if filter.UserID != nil {
		groupWhere = append(groupWhere, "a.user_id = ?")
		groupArgs = append(groupArgs, string(*filter.UserID))
	}
	if filter.GroupID != nil {
		groupWhere = append(groupWhere, "a.group_id = ?")
		groupArgs = append(groupArgs, string(*filter.GroupID))
	}

	groupWhereClause := ""
	if len(groupWhere) > 0 {
		groupWhereClause = "WHERE " + strings.Join(groupWhere, " AND ")
	}

	groupQuery := fmt.Sprintf(`
		SELECT
			g.id,
			g.name,
			COUNT(a.id) AS total_answers,
			COALESCE(SUM(a.is_correct), 0) AS total_correct,
			COALESCE(AVG(a.response_time_ms), 0) AS avg_time
		FROM answer_logs a
		JOIN groups g ON a.group_id = g.id
		%s
		GROUP BY g.id, g.name
		ORDER BY total_answers DESC, g.display_order ASC;
	`, groupWhereClause)

	gRows, err := r.db.QueryContext(ctx, groupQuery, groupArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to query group statistics: %w", err)
	}
	defer gRows.Close()

	for gRows.Next() {
		var gs model.GroupStat
		var gidStr string
		var gAvgTime float64
		if err := gRows.Scan(&gidStr, &gs.GroupName, &gs.TotalAnswers, &gs.CorrectAnswers, &gAvgTime); err != nil {
			return nil, fmt.Errorf("failed to scan group statistics: %w", err)
		}
		gs.GroupID = model.ID(gidStr)
		gs.AverageResponseTimeMs = int(gAvgTime)
		if gs.TotalAnswers > 0 {
			gs.AccuracyRate = float64(gs.CorrectAnswers) / float64(gs.TotalAnswers)
		}
		resp.Groups = append(resp.Groups, gs)
	}

	// 3. Weak targets (polymorphic: members and songs)
	limit := filter.Limit
	if limit <= 0 {
		limit = 5
	}
	if limit > 50 {
		limit = 50
	}

	var weakTargets []model.TargetStat

	// 3a. Member targets
	if filter.TargetType == nil || *filter.TargetType == model.TargetTypeMember {
		mWhere := []string{"a.target_member_id IS NOT NULL"}
		mArgs := []any{}
		if filter.UserID != nil {
			mWhere = append(mWhere, "a.user_id = ?")
			mArgs = append(mArgs, string(*filter.UserID))
		}
		if filter.GroupID != nil {
			mWhere = append(mWhere, "a.group_id = ?")
			mArgs = append(mArgs, string(*filter.GroupID))
		}

		mQuery := fmt.Sprintf(`
			SELECT
				m.id,
				(m.family_name || ' ' || m.given_name) AS full_name,
				m.group_id,
				COUNT(a.id) AS total_answers,
				COALESCE(SUM(a.is_correct), 0) AS total_correct,
				COALESCE(AVG(a.response_time_ms), 0) AS avg_time
			FROM answer_logs a
			JOIN members m ON a.target_member_id = m.id
			WHERE %s
			GROUP BY m.id, full_name, m.group_id
			HAVING total_answers > 0
			ORDER BY (CAST(total_correct AS REAL) / total_answers) ASC, total_answers DESC
			LIMIT ?;
		`, strings.Join(mWhere, " AND "))

		mArgs = append(mArgs, limit)
		mRows, err := r.db.QueryContext(ctx, mQuery, mArgs...)
		if err != nil {
			return nil, fmt.Errorf("failed to query member statistics: %w", err)
		}
		defer mRows.Close()

		for mRows.Next() {
			var ts model.TargetStat
			var tidStr, gidStr string
			var avgTime float64
			if err := mRows.Scan(&tidStr, &ts.Name, &gidStr, &ts.TotalAnswers, &ts.CorrectAnswers, &avgTime); err != nil {
				return nil, fmt.Errorf("failed to scan member stat: %w", err)
			}
			ts.TargetID = model.ID(tidStr)
			ts.TargetType = model.TargetTypeMember
			ts.GroupID = model.ID(gidStr)
			ts.AverageResponseTimeMs = int(avgTime)
			if ts.TotalAnswers > 0 {
				ts.AccuracyRate = float64(ts.CorrectAnswers) / float64(ts.TotalAnswers)
			}
			weakTargets = append(weakTargets, ts)
		}
	}

	// 3b. Song targets
	if filter.TargetType == nil || *filter.TargetType == model.TargetTypeSong {
		sWhere := []string{"a.target_song_id IS NOT NULL"}
		sArgs := []any{}
		if filter.UserID != nil {
			sWhere = append(sWhere, "a.user_id = ?")
			sArgs = append(sArgs, string(*filter.UserID))
		}
		if filter.GroupID != nil {
			sWhere = append(sWhere, "a.group_id = ?")
			sArgs = append(sArgs, string(*filter.GroupID))
		}

		sQuery := fmt.Sprintf(`
			SELECT
				s.id,
				s.title,
				s.group_id,
				COUNT(a.id) AS total_answers,
				COALESCE(SUM(a.is_correct), 0) AS total_correct,
				COALESCE(AVG(a.response_time_ms), 0) AS avg_time
			FROM answer_logs a
			JOIN songs s ON a.target_song_id = s.id
			WHERE %s
			GROUP BY s.id, s.title, s.group_id
			HAVING total_answers > 0
			ORDER BY (CAST(total_correct AS REAL) / total_answers) ASC, total_answers DESC
			LIMIT ?;
		`, strings.Join(sWhere, " AND "))

		sArgs = append(sArgs, limit)
		sRows, err := r.db.QueryContext(ctx, sQuery, sArgs...)
		if err != nil {
			return nil, fmt.Errorf("failed to query song statistics: %w", err)
		}
		defer sRows.Close()

		for sRows.Next() {
			var ts model.TargetStat
			var tidStr, gidStr string
			var avgTime float64
			if err := sRows.Scan(&tidStr, &ts.Name, &gidStr, &ts.TotalAnswers, &ts.CorrectAnswers, &avgTime); err != nil {
				return nil, fmt.Errorf("failed to scan song stat: %w", err)
			}
			ts.TargetID = model.ID(tidStr)
			ts.TargetType = model.TargetTypeSong
			ts.GroupID = model.ID(gidStr)
			ts.AverageResponseTimeMs = int(avgTime)
			if ts.TotalAnswers > 0 {
				ts.AccuracyRate = float64(ts.CorrectAnswers) / float64(ts.TotalAnswers)
			}
			weakTargets = append(weakTargets, ts)
		}
	}

	// Sort combined weak targets by accuracy ascending, then total answers descending
	if filter.TargetType == nil && len(weakTargets) > 1 {
		for i := 0; i < len(weakTargets)-1; i++ {
			for j := i + 1; j < len(weakTargets); j++ {
				// sort by accuracy ascending, then total answers descending
				if weakTargets[i].AccuracyRate > weakTargets[j].AccuracyRate ||
					(weakTargets[i].AccuracyRate == weakTargets[j].AccuracyRate && weakTargets[i].TotalAnswers < weakTargets[j].TotalAnswers) {
					weakTargets[i], weakTargets[j] = weakTargets[j], weakTargets[i]
				}
			}
		}
		if len(weakTargets) > limit {
			weakTargets = weakTargets[:limit]
		}
	}

	resp.WeakTargets = weakTargets
	return resp, nil
}

// GetMember returns a single member by ID (including graduated), along with all associated images.
func (r *SQLiteRepository) GetMember(ctx context.Context, id model.ID) (*model.Member, error) {
	const query = `
		SELECT m.id, m.group_id, m.family_name, m.given_name, m.family_name_kana, m.given_name_kana,
		       m.generation, m.status, m.left_color_id, m.right_color_id, m.ordered,
		       m.joined_at, m.graduated_at, m.verified_at, m.created_at, m.updated_at, m.metadata_revision
		FROM members m
		WHERE m.id = ?;
	`
	var m model.Member
	var orderedInt int
	var joinedAtStr, graduatedAtStr, verifiedAtStr sql.NullString
	var createdAtStr, updatedAtStr string

	err := r.db.QueryRowContext(ctx, query, string(id)).Scan(
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
		&verifiedAtStr,
		&createdAtStr,
		&updatedAtStr,
		&m.MetadataRevision,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get member: %w", err)
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
	if verifiedAtStr.Valid {
		t, _ := time.Parse(time.RFC3339, verifiedAtStr.String)
		m.VerifiedAt = &t
	}
	m.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	m.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)

	// Fetch all images for member
	images, err := r.ListMemberImages(ctx, m.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to list member images: %w", err)
	}
	m.Images = images

	return &m, nil
}
