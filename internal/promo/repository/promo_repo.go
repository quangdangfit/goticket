// Package repository implements promo.Repository.
package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/quangdangfit/goticket/internal/dbs"
	"github.com/quangdangfit/goticket/internal/promo"
	"github.com/quangdangfit/goticket/internal/promo/model"
	apperr "github.com/quangdangfit/goticket/pkg/errors"
)

type mysqlPromoRepository struct{ db dbs.MySQL }

// New returns a promo.Repository backed by MySQL.
func New(db dbs.MySQL) promo.Repository { return &mysqlPromoRepository{db: db} }

func (r *mysqlPromoRepository) Create(ctx context.Context, c *model.Code) error {
	if err := r.db.DB(ctx).Create(c).Error; err != nil {
		return fmt.Errorf("create promo: %w", err)
	}
	return nil
}

func (r *mysqlPromoRepository) Get(ctx context.Context, code string) (*model.Code, error) {
	var c model.Code
	err := r.db.DB(ctx).Where("code = ?", code).First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get promo: %w", err)
	}
	return &c, nil
}

func (r *mysqlPromoRepository) UserUseCount(ctx context.Context, code, userID string) (int, error) {
	var n int64
	err := r.db.DB(ctx).Model(&model.Redemption{}).
		Where("code = ? AND user_id = ?", code, userID).Count(&n).Error
	if err != nil {
		return 0, fmt.Errorf("user use count: %w", err)
	}
	return int(n), nil
}

func (r *mysqlPromoRepository) IncrementUsed(ctx context.Context, code string) error {
	res := r.db.DB(ctx).Model(&model.Code{}).
		Where("code = ? AND used < max_uses", code).
		UpdateColumn("used", gorm.Expr("used + 1"))
	if res.Error != nil {
		return fmt.Errorf("increment used: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return apperr.ErrSoldOut
	}
	return nil
}

func (r *mysqlPromoRepository) RecordRedemption(ctx context.Context, red *model.Redemption) error {
	if err := r.db.DB(ctx).Create(red).Error; err != nil {
		return fmt.Errorf("record redemption: %w", err)
	}
	return nil
}
