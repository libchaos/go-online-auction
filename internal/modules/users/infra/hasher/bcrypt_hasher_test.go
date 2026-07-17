package hasher_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"auction/internal/modules/users/domain/errs"
	"auction/internal/modules/users/infra/hasher"
)

type BcryptPasswordHasherTestSuite struct {
	suite.Suite
	sut *hasher.BcryptPasswordHasher
}

func (s *BcryptPasswordHasherTestSuite) SetupTest() {
	s.sut = hasher.NewBcryptPasswordHasher()
}

func TestBcryptPasswordHasherSuite(t *testing.T) {
	suite.Run(t, new(BcryptPasswordHasherTestSuite))
}

func (s *BcryptPasswordHasherTestSuite) TestHashAndCompare_ValidPassword_Succeeds() {
	// Arrange
	password := "super-secret-password"

	// Act
	hash, err := s.sut.Hash(password)

	// Assert
	s.Require().NoError(err)
	s.NotEmpty(hash)
	s.NotEqual(password, hash)
	s.Require().NoError(s.sut.Compare(hash, password))
}

func (s *BcryptPasswordHasherTestSuite) TestCompare_WrongPassword_ReturnsInvalidCredentials() {
	// Arrange
	hash, err := s.sut.Hash("correct-password")
	s.Require().NoError(err)

	// Act
	compareErr := s.sut.Compare(hash, "wrong-password")

	// Assert
	s.Require().ErrorIs(compareErr, errs.ErrInvalidCredentials)
}
