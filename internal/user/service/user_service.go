// Package service implements the user.Service port.
package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/quangdangfit/goticket/internal/user"
	"github.com/quangdangfit/goticket/internal/user/dto"
	"github.com/quangdangfit/goticket/internal/user/model"
	apperr "github.com/quangdangfit/goticket/pkg/errors"
	"github.com/quangdangfit/goticket/pkg/hash"
	"github.com/quangdangfit/goticket/pkg/jwt"
)

type userService struct {
	repo user.Repository
	jwt  jwt.Manager
}

// New constructs a user.Service.
func New(repo user.Repository, j jwt.Manager) user.Service {
	return &userService{repo: repo, jwt: j}
}

func (s *userService) Register(ctx context.Context, in dto.RegisterInput) (*dto.AuthOutput, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))
	if existing, err := s.repo.GetByEmail(ctx, email); err == nil && existing != nil {
		return nil, fmt.Errorf("register: %w", apperr.ErrConflict)
	} else if err != nil && !errors.Is(err, apperr.ErrNotFound) {
		return nil, err
	}
	ph, err := hash.Password(in.Password)
	if err != nil {
		return nil, fmt.Errorf("register: %w", err)
	}
	now := time.Now().UTC()
	u := &model.User{
		ID:           strings.ReplaceAll(uuid.NewString(), "-", "")[:26],
		Email:        email,
		PasswordHash: ph,
		Name:         in.Name,
		Phone:        in.Phone,
		Role:         model.RoleUser,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}
	return s.issueTokens(ctx, u)
}

func (s *userService) Login(ctx context.Context, in dto.LoginInput) (*dto.AuthOutput, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))
	u, err := s.repo.GetByEmail(ctx, email)
	if errors.Is(err, apperr.ErrNotFound) {
		return nil, apperr.ErrUnauthorized
	}
	if err != nil {
		return nil, err
	}
	if err := hash.Compare(u.PasswordHash, in.Password); err != nil {
		return nil, apperr.ErrUnauthorized
	}
	return s.issueTokens(ctx, u)
}

func (s *userService) Refresh(ctx context.Context, refreshToken string) (*dto.AuthOutput, error) {
	h := s.jwt.HashRefresh(refreshToken)
	rt, err := s.repo.GetRefresh(ctx, h)
	if errors.Is(err, apperr.ErrNotFound) {
		return nil, apperr.ErrUnauthorized
	}
	if err != nil {
		return nil, err
	}
	if rt.RevokedAt != nil || time.Now().UTC().After(rt.ExpiresAt) {
		return nil, apperr.ErrUnauthorized
	}
	if err := s.repo.RevokeRefresh(ctx, h, time.Now().UTC()); err != nil {
		return nil, err
	}
	u, err := s.repo.GetByID(ctx, rt.UserID)
	if err != nil {
		return nil, err
	}
	return s.issueTokens(ctx, u)
}

func (s *userService) Profile(ctx context.Context, userID string) (*dto.Profile, error) {
	u, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	p := toProfile(u)
	return &p, nil
}

func (s *userService) Logout(ctx context.Context, refreshToken string) error {
	return s.repo.RevokeRefresh(ctx, s.jwt.HashRefresh(refreshToken), time.Now().UTC())
}

func (s *userService) issueTokens(ctx context.Context, u *model.User) (*dto.AuthOutput, error) {
	access, accessExp, err := s.jwt.IssueAccess(u.ID, u.Role)
	if err != nil {
		return nil, fmt.Errorf("issue access: %w", err)
	}
	refresh, refHash, refExp, err := s.jwt.IssueRefresh(u.ID)
	if err != nil {
		return nil, fmt.Errorf("issue refresh: %w", err)
	}
	rt := &model.RefreshToken{
		ID:        strings.ReplaceAll(uuid.NewString(), "-", "")[:26],
		UserID:    u.ID,
		TokenHash: refHash,
		ExpiresAt: refExp,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.repo.StoreRefresh(ctx, rt); err != nil {
		return nil, err
	}
	return &dto.AuthOutput{
		AccessToken:      access,
		RefreshToken:     refresh,
		AccessExpiresAt:  accessExp,
		RefreshExpiresAt: refExp,
		User:             toProfile(u),
	}, nil
}

func toProfile(u *model.User) dto.Profile {
	return dto.Profile{
		ID: u.ID, Email: u.Email, Name: u.Name, Phone: u.Phone,
		Role: u.Role, CreatedAt: u.CreatedAt,
	}
}
