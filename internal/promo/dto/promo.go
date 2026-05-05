// Package dto holds shapes for the promo HTTP port.
package dto

import "time"

// CreatePromoInput is the body of POST /admin/promos.
type CreatePromoInput struct {
	Code         string    `json:"code"           validate:"required,min=3,max=64"`
	Type         string    `json:"type"           validate:"required,oneof=percent fixed"`
	ValueMinor   int64     `json:"value_minor"    validate:"required_if=Type fixed,min=0"`
	Percent      int       `json:"percent"        validate:"required_if=Type percent,min=0,max=100"`
	MaxUses      int       `json:"max_uses"       validate:"required,min=1"`
	PerUserLimit int       `json:"per_user_limit" validate:"required,min=1"`
	StartsAt     time.Time `json:"starts_at"`
	ExpiresAt    time.Time `json:"expires_at"     validate:"required"`
}
