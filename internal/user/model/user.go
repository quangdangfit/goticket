// Package model holds GORM entities for the user bounded context.
package model

import "time"

// Role values.
const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

// User is the persisted account record.
type User struct {
	ID           string    `gorm:"primaryKey;type:char(26)"`
	Email        string    `gorm:"uniqueIndex;size:255"`
	PasswordHash string    `gorm:"size:255"`
	Name         string    `gorm:"size:255"`
	Phone        string    `gorm:"size:32"`
	Role         string    `gorm:"size:16;default:user"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// TableName overrides GORM's pluralizer.
func (User) TableName() string { return "users" }

// RefreshToken stores a hashed refresh token for revocation/rotation.
type RefreshToken struct {
	ID        string    `gorm:"primaryKey;type:char(26)"`
	UserID    string    `gorm:"type:char(26);index"`
	TokenHash string    `gorm:"uniqueIndex;type:char(64)"`
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

// TableName overrides GORM's pluralizer.
func (RefreshToken) TableName() string { return "refresh_tokens" }
