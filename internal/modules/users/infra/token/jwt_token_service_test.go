package token_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"auction/internal/modules/users/infra/token"
	"auction/internal/shared/modules/config"
)

type JWTTokenServiceTestSuite struct {
	suite.Suite
	sut *token.JWTTokenService
	cfg config.Config
}

func (s *JWTTokenServiceTestSuite) SetupTest() {
	s.cfg = config.Config{
		App: config.App{Name: "go-online-auction-test"},
		JWT: config.JWT{
			Secret:                "test-secret-with-enough-length-123",
			AccessTokenTTLMinutes: 15,
			RefreshTokenTTLHours:  168,
		},
	}

	sut, err := token.NewJWTTokenService(s.cfg)
	s.Require().NoError(err)
	s.sut = sut
}

func TestJWTTokenServiceSuite(t *testing.T) {
	suite.Run(t, new(JWTTokenServiceTestSuite))
}

func (s *JWTTokenServiceTestSuite) TestNewJWTTokenService_EmptySecret_ReturnsError() {
	// Arrange
	cfg := config.Config{}

	// Act
	_, err := token.NewJWTTokenService(cfg)

	// Assert
	s.Require().ErrorIs(err, token.ErrEmptyJWTSecret)
}

func (s *JWTTokenServiceTestSuite) TestGenerateAccessToken_ThenVerify_ReturnsClaims() {
	// Arrange
	userID := uint64(42)
	role := "seller"
	email := "john@example.com"

	// Act
	accessToken, expiresAt, err := s.sut.GenerateAccessToken(userID, role, email)

	// Assert
	s.Require().NoError(err)
	s.NotEmpty(accessToken)
	s.True(expiresAt.After(time.Now().UTC()))

	claims, err := s.sut.Verify(accessToken)
	s.Require().NoError(err)
	s.Equal(userID, claims.UserID)
	s.Equal(role, claims.Role)
	s.Equal(email, claims.Email)
}

func (s *JWTTokenServiceTestSuite) TestVerify_WrongSecret_ReturnsError() {
	// Arrange
	otherConfig := s.cfg
	otherConfig.JWT.Secret = "a-completely-different-secret-456"
	otherService, err := token.NewJWTTokenService(otherConfig)
	s.Require().NoError(err)

	accessToken, _, err := otherService.GenerateAccessToken(1, "bidder", "a@b.com")
	s.Require().NoError(err)

	// Act
	_, verifyErr := s.sut.Verify(accessToken)

	// Assert
	s.Require().Error(verifyErr)
}

func (s *JWTTokenServiceTestSuite) TestVerify_MalformedToken_ReturnsError() {
	// Act
	_, err := s.sut.Verify("not-a-jwt")

	// Assert
	s.Require().Error(err)
}

func (s *JWTTokenServiceTestSuite) TestVerify_UnsignedAlgorithm_ReturnsError() {
	// Arrange: alg=none token (header {"alg":"none","typ":"JWT"})
	noneToken := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0." +
		"eyJzdWIiOiIxIiwicm9sZSI6ImFkbWluIn0."

	// Act
	_, err := s.sut.Verify(noneToken)

	// Assert
	s.Require().Error(err)
}

func (s *JWTTokenServiceTestSuite) TestGenerateRefreshToken_ReturnsRawAndHash() {
	// Act
	raw, hash, expiresAt, err := s.sut.GenerateRefreshToken()

	// Assert
	s.Require().NoError(err)
	s.NotEmpty(raw)
	s.NotEmpty(hash)
	s.NotEqual(raw, hash)
	s.Equal(hash, s.sut.HashRefreshToken(raw))
	s.True(expiresAt.After(time.Now().UTC()))
}
