package quiz

import "github.com/aobaiwaki/penlight-v2/pkg/model"

// QuizTarget represents an entity that can be quizzed for official penlight colors (Ref: ADR-0031).
// Both Member (member colors) and Song (song colors) implement this interface.
type QuizTarget interface {
	GetID() model.ID
	GetCorrectColors() []model.ID // Returns 1 or 2 official color IDs
}

// Compile-time interface verification
var (
	_ QuizTarget = (*model.Member)(nil)
	_ QuizTarget = (*model.Song)(nil)
)
