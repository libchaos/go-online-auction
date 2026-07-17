package hasher

import (
	"errors"

	"golang.org/x/crypto/bcrypt"

	"auction/internal/modules/users/domain/errs"
	"auction/internal/modules/users/ports"
)

var _ ports.PasswordHasher = (*BcryptPasswordHasher)(nil)

type BcryptPasswordHasher struct{}

func NewBcryptPasswordHasher() *BcryptPasswordHasher {
	return &BcryptPasswordHasher{}
}

func (h *BcryptPasswordHasher) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func (h *BcryptPasswordHasher) Compare(hash, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return errs.ErrInvalidCredentials
		}
		return err
	}

	return nil
}
