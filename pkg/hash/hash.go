// Package hash wraps bcrypt password hashing.
package hash

import "golang.org/x/crypto/bcrypt"

// Password hashes a plaintext password.
func Password(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Compare reports whether plain matches the stored bcrypt hash.
func Compare(stored, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(stored), []byte(plain))
}
