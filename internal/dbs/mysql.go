package dbs

import (
	"context"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/quangdangfit/goticket/config"
)

// MySQL is the narrow port that domain repositories depend on.
type MySQL interface {
	DB(ctx context.Context) *gorm.DB
	Close() error
}

type gormMySQL struct{ db *gorm.DB }

// NewMySQL opens a GORM/MySQL connection pool from cfg.
func NewMySQL(cfg config.MySQLConfig) (MySQL, error) {
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("sql db: %w", err)
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	return &gormMySQL{db: db}, nil
}

func (g *gormMySQL) DB(ctx context.Context) *gorm.DB { return g.db.WithContext(ctx) }

func (g *gormMySQL) Close() error {
	sqlDB, err := g.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
