// Package repository implements the user.Repository port against MySQL/GORM.
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	apperr "github.com/quangdangfit/goticket/pkg/errors"

	"github.com/quangdangfit/goticket/internal/dbs"
	"github.com/quangdangfit/goticket/internal/user"
	"github.com/quangdangfit/goticket/internal/user/model"
)

type mysqlUserRepository struct{ db dbs.MySQL }

// New returns a user.Repository backed by MySQL.
func New(db dbs.MySQL) user.Repository { return &mysqlUserRepository{db: db} }

func (r *mysqlUserRepository) Create(ctx context.Context, u *model.User) error {
	if err := r.db.DB(ctx).Create(u).Error; err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *mysqlUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	err := r.db.DB(ctx).Where("email = ?", email).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &u, nil
}

func (r *mysqlUserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	var u model.User
	err := r.db.DB(ctx).Where("id = ?", id).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &u, nil
}

func (r *mysqlUserRepository) StoreRefresh(ctx context.Context, t *model.RefreshToken) error {
	if err := r.db.DB(ctx).Create(t).Error; err != nil {
		return fmt.Errorf("store refresh: %w", err)
	}
	return nil
}

func (r *mysqlUserRepository) GetRefresh(ctx context.Context, tokenHash string) (*model.RefreshToken, error) {
	var t model.RefreshToken
	err := r.db.DB(ctx).Where("token_hash = ?", tokenHash).First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get refresh: %w", err)
	}
	return &t, nil
}

func (r *mysqlUserRepository) RevokeRefresh(ctx context.Context, tokenHash string, at time.Time) error {
	res := r.db.DB(ctx).Model(&model.RefreshToken{}).
		Where("token_hash = ? AND revoked_at IS NULL", tokenHash).
		Update("revoked_at", at)
	if res.Error != nil {
		return fmt.Errorf("revoke refresh: %w", res.Error)
	}
	return nil
}
