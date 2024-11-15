package hasher

import (
	"golang.org/x/crypto/bcrypt"
	"log"
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
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	log.Println("Hahshing password ", password, " to ", string(hashedPassword))
	return string(hashedPassword), nil

}

func (h *hasher) Compare(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
