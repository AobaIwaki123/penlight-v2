package model

import "time"

// BootstrapResponse contains all master records required for offline PWA operation (Ref: ADR-0007, ADR-0026, ADR-0027).
type BootstrapResponse struct {
	Series      []Series  `json:"series"`
	Groups      []Group   `json:"groups"`
	Colors      []Color   `json:"colors"`
	Members     []Member  `json:"members"`
	Songs       []Song    `json:"songs"`
	GeneratedAt time.Time `json:"generated_at"`
}

// GenerateQuizRequest holds query parameters for quiz generation.
type GenerateQuizRequest struct {
	GroupID    *ID  `json:"group_id,omitempty"`
	Generation *int `json:"generation,omitempty"`
	Strategy   string `json:"strategy"` // random, similar_color, spaced_repetition
	Count      int    `json:"count"`    // 1-10
}

// GenerateQuizResponse returns dynamically generated 4-choice questions.
type GenerateQuizResponse struct {
	Questions []QuizQuestion `json:"questions"`
}

// SubmitAnswerRequest represents a single quiz answer submission.
type SubmitAnswerRequest struct {
	QuizQuestionID   ID `json:"quiz_question_id"`
	TargetMemberID   ID `json:"target_member_id"`
	SelectedMemberID ID `json:"selected_member_id"`
	ResponseTimeMs   int `json:"response_time_ms"`
}

// SubmitAnswerResponse returns immediate grading and explanation.
type SubmitAnswerResponse struct {
	IsCorrect       bool `json:"is_correct"`
	CorrectMemberID ID   `json:"correct_member_id"`
	AnswerLogID     ID   `json:"answer_log_id"`
}

// BatchAnswerItem represents an offline queued answer record (Ref: ADR-0007, ADR-0032).
type BatchAnswerItem struct {
	ID             ID        `json:"id"` // Pre-generated ans_<uuidv7> (idempotency key)
	UserID         ID        `json:"user_id,omitempty"`
	QuizQuestionID ID        `json:"quiz_question_id"`
	TargetMemberID *ID       `json:"target_member_id,omitempty"` // mem_... (メンバー問題時)
	TargetSongID   *ID       `json:"target_song_id,omitempty"`   // sng_... (楽曲問題時)
	GroupID        ID        `json:"group_id"`
	IsCorrect      bool      `json:"is_correct"`
	ResponseTimeMs int       `json:"response_time_ms"`
	AnsweredAt     time.Time `json:"answered_at"`
}

// BatchAnswerRequest holds multiple offline answer logs.
type BatchAnswerRequest struct {
	Answers []BatchAnswerItem `json:"answers"`
}

// BatchAnswerResponse returns the count of successfully synchronized items.
type BatchAnswerResponse struct {
	SyncedCount int  `json:"synced_count"`
	SyncedIDs   []ID `json:"synced_ids"`
}

// UserStatisticsResponse returns aggregated metrics for a user.
type UserStatisticsResponse struct {
	TotalAnswers          int     `json:"total_answers"`
	TotalCorrect          int     `json:"total_correct"`
	AccuracyRate          float64 `json:"accuracy_rate"`
	AverageResponseTimeMs int     `json:"average_response_time_ms"`
}

// TargetType indicates the entity type for quiz statistics (member / song / etc.).
type TargetType string

const (
	TargetTypeMember TargetType = "member"
	TargetTypeSong   TargetType = "song"
)

// TargetStat represents performance statistics for a specific target.
type TargetStat struct {
	TargetID              ID             `json:"target_id"`              // mem_... or sng_...
	TargetType            TargetType     `json:"target_type"`            // member / song
	GroupID               ID             `json:"group_id"`               // grp_...
	Name                  string         `json:"name"`                   // メンバー名 / 楽曲名
	TotalAnswers          int            `json:"total_answers"`
	CorrectAnswers        int            `json:"correct_answers"`
	AccuracyRate          float64        `json:"accuracy_rate"`          // 0.0 ~ 1.0
	AverageResponseTimeMs int            `json:"average_response_time_ms"`
	Metadata              map[string]any `json:"metadata,omitempty"`     // 任意付加情報 (kana, last_answered_at, streak等)
}

// GroupStat represents aggregated metrics for an idol group.
type GroupStat struct {
	GroupID               ID             `json:"group_id"`
	GroupName             string         `json:"group_name"`
	TotalAnswers          int            `json:"total_answers"`
	CorrectAnswers        int            `json:"correct_answers"`
	AccuracyRate          float64        `json:"accuracy_rate"`
	AverageResponseTimeMs int            `json:"average_response_time_ms"`
	Metadata              map[string]any `json:"metadata,omitempty"`     // 任意付加情報 (series_id, rank等)
}

// QuizStatisticsFilter specifies criteria for filtering quiz statistics.
type QuizStatisticsFilter struct {
	UserID     *ID
	GroupID    *ID
	TargetType *TargetType
	Limit      int // limit for weak targets (default: 5)
}

// QuizStatisticsResponse holds core analytical data and flexible extras (Ref: ADR-0032).
type QuizStatisticsResponse struct {
	TotalAnswers          int            `json:"total_answers"`
	TotalCorrect          int            `json:"total_correct"`
	AccuracyRate          float64        `json:"accuracy_rate"`
	AverageResponseTimeMs int            `json:"average_response_time_ms"`
	Groups                []GroupStat    `json:"groups,omitempty"`
	WeakTargets           []TargetStat   `json:"weak_targets,omitempty"` // 苦手克服向けワーストN件
	Extra                 map[string]any `json:"extra,omitempty"`        // 将来の拡張・実験的集計データ
}

