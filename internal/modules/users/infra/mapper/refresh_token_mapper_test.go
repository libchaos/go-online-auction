package mapper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"auction/internal/modules/users/domain/model"
	"auction/internal/modules/users/infra/mapper"
	"auction/internal/modules/users/infra/sqlcgen"
)

type RefreshTokenMapperTestSuite struct {
	suite.Suite
	sut *mapper.RefreshTokenMapper
}

func (s *RefreshTokenMapperTestSuite) SetupTest() {
	s.sut = mapper.NewRefreshTokenMapper()
}

func TestRefreshTokenMapperSuite(t *testing.T) {
	suite.Run(t, new(RefreshTokenMapperTestSuite))
}

func (s *RefreshTokenMapperTestSuite) TestToDomain_ValidRow_ReturnsModel() {
	// Arrange
	now := time.Now().UTC()
	revokedAt := now.Add(-time.Hour)
	replacedBy := int64(9)
	t := sqlcgen.RefreshToken{
		ID:         1,
		UserID:     2,
		TokenHash:  "token-hash",
		ExpiresAt:  now.Add(time.Hour),
		RevokedAt:  &revokedAt,
		ReplacedBy: &replacedBy,
		CreatedAt:  now,
	}

	// Act
	token, err := s.sut.ToDomain(t)

	// Assert
	s.Require().NoError(err)
	s.Equal(uint64(t.ID), token.ID())
	s.Equal(uint64(t.UserID), token.UserID())
	s.Equal(t.TokenHash, token.TokenHash())
	s.True(token.IsRevoked())
	s.Equal(uint64(replacedBy), *token.ReplacedBy())
}

func (s *RefreshTokenMapperTestSuite) TestToDomain_ZeroID_ReturnsError() {
	// Arrange
	now := time.Now().UTC()
	t := sqlcgen.RefreshToken{
		ID:        0,
		UserID:    2,
		TokenHash: "token-hash",
		ExpiresAt: now,
		CreatedAt: now,
	}

	// Act
	_, err := s.sut.ToDomain(t)

	// Assert
	s.Require().Error(err)
}

func (s *RefreshTokenMapperTestSuite) TestToCreateParams_ValidModel_ReturnsParams() {
	// Arrange
	expiresAt := time.Now().UTC().Add(time.Hour)
	token, _ := model.NewRefreshTokenModel(2, "token-hash", expiresAt)

	// Act
	params := s.sut.ToCreateParams(token)

	// Assert
	s.Equal(int64(token.UserID()), params.UserID)
	s.Equal(token.TokenHash(), params.TokenHash)
	s.Equal(token.ExpiresAt(), params.ExpiresAt)
	s.Nil(params.RevokedAt)
	s.Nil(params.ReplacedBy)
}

func (s *RefreshTokenMapperTestSuite) TestToUpdateParams_RevokedModel_ReturnsParams() {
	// Arrange
	now := time.Now().UTC()
	revokedAt := now.Add(-time.Minute)
	replacedBy := uint64(9)
	token, err := model.RestoreRefreshTokenModel(
		1,
		2,
		"token-hash",
		now.Add(time.Hour),
		&revokedAt,
		&replacedBy,
		now,
	)
	s.Require().NoError(err)

	// Act
	params := s.sut.ToUpdateParams(token)

	// Assert
	s.Equal(int64(token.ID()), params.ID)
	s.Equal(&revokedAt, params.RevokedAt)
	s.Equal(int64(replacedBy), *params.ReplacedBy)
}
