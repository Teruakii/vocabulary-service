package auth

import (
	"time"

	"vocabulary-service/internal/constants"
)

type User struct {
	ID                uint64         `gorm:"primaryKey;autoIncrement;column:user_id" json:"id"`
	Firstname         string         `gorm:"column:first_name" json:"firstname"`
	Lastname          string         `gorm:"column:last_name" json:"lastname"`
	Email             string         `gorm:"column:email;type:varchar(255);not null;uniqueIndex:idx_users_email" json:"email"`
	Password          *string        `gorm:"column:password" json:"-"`
	Role              constants.Role `gorm:"column:role;not null;default:user" json:"role"`
	PasswordChangedAt *time.Time     `gorm:"column:password_changed_at" json:"-"`
	// EmailVerifiedAt is set once the user proves they own Email (Google
	// sign-in or an emailed reset link) and cleared when Email changes.
	EmailVerifiedAt *time.Time `gorm:"column:email_verified_at" json:"-"`
	LastLoginAt     *time.Time `gorm:"column:last_login_at" json:"lastLoginAt"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) HasPassword() bool {
	return u.Password != nil && *u.Password != ""
}

type Identity struct {
	ID         uint64             `gorm:"primaryKey;autoIncrement;column:identity_id" json:"id"`
	UserID     uint64             `gorm:"column:user_id;not null" json:"userId"`
	Provider   constants.Provider `gorm:"column:provider;type:varchar(50);not null;uniqueIndex:idx_provider_identifier" json:"provider"`
	Identifier string             `gorm:"column:identifier;type:varchar(255);not null;uniqueIndex:idx_provider_identifier" json:"identifier"`
	CreatedAt  time.Time          `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt  time.Time          `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (Identity) TableName() string {
	return "identities"
}

// Session is one issued refresh token. Refresh only works while its session
// exists and is not revoked, so logging out really ends it server-side.
type Session struct {
	ID        string     `gorm:"primaryKey;column:session_id;type:varchar(64)"`
	UserID    uint64     `gorm:"column:user_id;not null;index"`
	ExpiresAt time.Time  `gorm:"column:expires_at;not null"`
	RevokedAt *time.Time `gorm:"column:revoked_at"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
}

func (Session) TableName() string {
	return "refresh_sessions"
}

// Active reports whether the session can still be used to refresh.
func (s *Session) Active(now time.Time) bool {
	return s.RevokedAt == nil && now.Before(s.ExpiresAt)
}

// PasswordReset is one emailed reset link. Only the SHA-256 of the token is
// stored. Email is the address the link went to: if the account's email has
// changed since, the link no longer proves anything about the account.
type PasswordReset struct {
	TokenHash string    `gorm:"primaryKey;column:token_hash;type:varchar(64)"`
	UserID    uint64    `gorm:"column:user_id;not null;index"`
	Email     string    `gorm:"column:email;type:varchar(255);not null"`
	ExpiresAt time.Time `gorm:"column:expires_at;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (PasswordReset) TableName() string {
	return "password_resets"
}
