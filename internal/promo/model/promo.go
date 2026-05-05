// Package model holds GORM entities for the promo bounded context.
package model

import "time"

// Discount kinds.
const (
	TypePercent = "percent"
	TypeFixed   = "fixed"
)

// Code is a promo code definition.
type Code struct {
	Code         string `gorm:"primaryKey;size:64"`
	Type         string `gorm:"size:16"`
	ValueMinor   int64
	Percent      int
	MaxUses      int
	Used         int
	PerUserLimit int
	StartsAt     time.Time
	ExpiresAt    time.Time
	CreatedAt    time.Time
}

// TableName overrides GORM's pluralizer.
func (Code) TableName() string { return "promo_codes" }

// Redemption logs a successful application.
type Redemption struct {
	ID            string `gorm:"primaryKey;type:char(26)"`
	Code          string `gorm:"size:64;index"`
	UserID        string `gorm:"type:char(26)"`
	OrderID       string `gorm:"type:char(26)"`
	DiscountMinor int64
	CreatedAt     time.Time
}

// TableName overrides GORM's pluralizer.
func (Redemption) TableName() string { return "promo_redemptions" }
