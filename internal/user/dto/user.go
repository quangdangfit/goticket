// Package dto holds request/response shapes for the user HTTP port.
package dto

import "time"

// RegisterInput is the body of POST /auth/register.
type RegisterInput struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
	Name     string `json:"name"     validate:"required,min=1,max=255"`
	Phone    string `json:"phone"    validate:"omitempty,max=32"`
}

// LoginInput is the body of POST /auth/login.
type LoginInput struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RefreshInput is the body of POST /auth/refresh.
type RefreshInput struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// AuthOutput is returned by register/login/refresh.
type AuthOutput struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	AccessExpiresAt  time.Time `json:"access_expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
	User             Profile   `json:"user"`
}

// Profile is the public-facing view of a user.
type Profile struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}
