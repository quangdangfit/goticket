// Package mocks holds gomock-generated test doubles, one subpackage per
// domain. Regenerate via `make mocks`.
package mocks

//go:generate mockgen -package usermock -destination usermock/user.go github.com/quangdangfit/goticket/internal/user Repository,Service
