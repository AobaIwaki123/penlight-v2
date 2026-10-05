package model

import "time"

// AnswerLog records an individual quiz question response by a user for statistics and review.
type AnswerLog struct {
	ID             ID        `json:"id" db:"id,pk"`
	UserID         ID        `json:"user_id" db:"user_id,fk"`
	QuizQuestionID ID        `json:"quiz_question_id" db:"quiz_question_id"`
	TargetMemberID *ID       `json:"target_member_id,omitempty" db:"target_member_id,fk"` // mem_... (メンバー問題時)
	TargetSongID   *ID       `json:"target_song_id,omitempty" db:"target_song_id,fk"`     // sng_... (楽曲問題時)
	GroupID        ID        `json:"group_id" db:"group_id,fk"`
	IsCorrect      bool      `json:"is_correct" db:"is_correct"`
	ResponseTimeMs int       `json:"response_time_ms" db:"response_time_ms"`
	AnsweredAt     time.Time `json:"answered_at" db:"answered_at"`
}

// GetTargetID returns whichever target ID is populated.
func (a *AnswerLog) GetTargetID() ID {
	if a.TargetMemberID != nil {
		return *a.TargetMemberID
	}
	if a.TargetSongID != nil {
		return *a.TargetSongID
	}
	return ""
}
