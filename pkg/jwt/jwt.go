// Package jwt issues and verifies HS256 access + refresh tokens.
package jwt

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

// Manager signs and verifies tokens.
type Manager interface {
	IssueAccess(userID, role string) (token string, expiresAt time.Time, err error)
	IssueRefresh(userID string) (token, hash string, expiresAt time.Time, err error)
	Verify(token string) (userID, role string, err error)
	HashRefresh(token string) string
}

type manager struct {
	access     []byte
	refresh    []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// New returns a Manager configured with HS256 secrets and TTLs.
func New(accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration) Manager {
	return &manager{
		access:     []byte(accessSecret),
		refresh:    []byte(refreshSecret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

type accessClaims struct {
	UserID string `json:"sub"`
	Role   string `json:"role"`
	jwtv5.RegisteredClaims
}

func (m *manager) IssueAccess(userID, role string) (string, time.Time, error) {
	exp := time.Now().UTC().Add(m.accessTTL)
	c := accessClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwtv5.RegisteredClaims{
			ExpiresAt: jwtv5.NewNumericDate(exp),
			IssuedAt:  jwtv5.NewNumericDate(time.Now().UTC()),
		},
	}
	t, err := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, c).SignedString(m.access)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign access: %w", err)
	}
	return t, exp, nil
}

func (m *manager) IssueRefresh(userID string) (string, string, time.Time, error) {
	exp := time.Now().UTC().Add(m.refreshTTL)
	c := accessClaims{
		UserID: userID,
		RegisteredClaims: jwtv5.RegisteredClaims{
			ExpiresAt: jwtv5.NewNumericDate(exp),
			IssuedAt:  jwtv5.NewNumericDate(time.Now().UTC()),
		},
	}
	t, err := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, c).SignedString(m.refresh)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("sign refresh: %w", err)
	}
	return t, m.HashRefresh(t), exp, nil
}

func (m *manager) Verify(token string) (string, string, error) {
	c := &accessClaims{}
	t, err := jwtv5.ParseWithClaims(token, c, func(t *jwtv5.Token) (any, error) {
		if _, ok := t.Method.(*jwtv5.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.access, nil
	})
	if err != nil || !t.Valid {
		return "", "", errors.New("invalid token")
	}
	return c.UserID, c.Role, nil
}

func (m *manager) HashRefresh(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
