package auth

import (
	"time"

	"vocabulary-service/internal/constants"
)

type User struct {
	ID                uint64         `gorm:"primaryKey;autoIncrement;column:user_id" json:"id"`
	Firstname         string         `gorm:"column:first_name" json:"firstname"`
	Lastname          string         `gorm:"column:last_name" json:"lastname"`
	Email             string         `gorm:"column:email;not null" json:"email"`
	Password          *string        `gorm:"column:password" json:"-"`
	Role              constants.Role `gorm:"column:role;not null;default:user" json:"role"`
	PasswordChangedAt *time.Time     `gorm:"column:password_changed_at" json:"-"`
	LastLoginAt       *time.Time     `gorm:"column:last_login_at" json:"lastLoginAt"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
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
