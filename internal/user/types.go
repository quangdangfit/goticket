// Package user is the identity & auth bounded context.
package user

import (
	"context"
	"time"

	"github.com/quangdangfit/goticket/internal/user/dto"
	"github.com/quangdangfit/goticket/internal/user/model"
)

// Repository persists users and refresh tokens.
type Repository interface {
	Create(ctx context.Context, u *model.User) error
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByID(ctx context.Context, id string) (*model.User, error)

	StoreRefresh(ctx context.Context, t *model.RefreshToken) error
	GetRefresh(ctx context.Context, tokenHash string) (*model.RefreshToken, error)
	RevokeRefresh(ctx context.Context, tokenHash string, at time.Time) error
}

// Service is the user-facing port mounted on HTTP.
type Service interface {
	Register(ctx context.Context, in dto.RegisterInput) (*dto.AuthOutput, error)
	Login(ctx context.Context, in dto.LoginInput) (*dto.AuthOutput, error)
	Refresh(ctx context.Context, refreshToken string) (*dto.AuthOutput, error)
	Profile(ctx context.Context, userID string) (*dto.Profile, error)
	Logout(ctx context.Context, refreshToken string) error
}
