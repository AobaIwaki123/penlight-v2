package model

import "time"

// User represents an authenticated or registered user.
type User struct {
	ID          ID        `json:"id" db:"id,pk"`                       // usr_<uuidv7>
	GoogleSub   *string   `json:"google_sub,omitempty" db:"google_sub"` // Google OIDC unique subject ID
	Email       *string   `json:"email,omitempty" db:"email"`           // Google account email address
	DisplayName string    `json:"display_name" db:"display_name"`       // User public display name
	AvatarURL   string    `json:"avatar_url" db:"avatar_url"`           // Avatar image URL
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	LastLoginAt time.Time `json:"last_login_at" db:"last_login_at"`
}
