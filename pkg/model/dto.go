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

// BatchAnswerItem represents an offline queued answer record.
type BatchAnswerItem struct {
	ID             ID        `json:"id"` // Pre-generated ans_<uuidv7> (idempotency key)
	QuizQuestionID ID        `json:"quiz_question_id"`
	TargetMemberID ID        `json:"target_member_id"`
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
