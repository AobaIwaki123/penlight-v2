package model

import "time"

// AnswerLog records an individual quiz question response by a user for statistics and review.
type AnswerLog struct {
	ID             ID        `json:"id" db:"id,pk"`                             // ans_<uuidv7>
	UserID         ID        `json:"user_id" db:"user_id,fk"`                   // usr_<uuidv7>
	QuizQuestionID ID        `json:"quiz_question_id" db:"quiz_question_id"`       // quiz_<uuidv7>
	TargetMemberID ID        `json:"target_member_id" db:"target_member_id,fk"` // mem_<uuidv7>
	GroupID        ID        `json:"group_id" db:"group_id,fk"`                 // grp_<uuidv7> (for fast aggregation)
	IsCorrect      bool      `json:"is_correct" db:"is_correct"`               // whether the answer was correct
	ResponseTimeMs int       `json:"response_time_ms" db:"response_time_ms"`   // reaction time in milliseconds
	AnsweredAt     time.Time `json:"answered_at" db:"answered_at"`             // timestamp of answer
}
