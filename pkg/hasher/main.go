package hasher

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type Hasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, password string) error
}

// NewHasher creates a new Hasher instance
func NewHasher() Hasher {
	return &hasher{}
}

type hasher struct {
}

func (h *hasher) Hash(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hashing password: %w", err)
	}
	return string(hashedPassword), nil
}

func (h *hasher) Compare(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
