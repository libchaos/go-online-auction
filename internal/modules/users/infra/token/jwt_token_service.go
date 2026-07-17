package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"auction/internal/modules/users/ports"
	"auction/internal/shared/modules/authn"
	"auction/internal/shared/modules/config"
)

const (
	signingMethod         = "HS256"
	refreshTokenByteSize  = 32
	defaultAccessTTLMin   = 15
	defaultRefreshTTLHour = 168
)

var (
	ErrEmptyJWTSecret = errors.New("JWT secret must not be empty")
	ErrInvalidToken   = errors.New("invalid access token")
)

var (
	_ ports.TokenService  = (*JWTTokenService)(nil)
	_ authn.TokenVerifier = (*JWTTokenService)(nil)
)

type JWTTokenService struct {
	secret     []byte
	issuer     string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewJWTTokenService(cfg config.Config) (*JWTTokenService, error) {
	if cfg.JWT.Secret == "" {
		return nil, ErrEmptyJWTSecret
	}

	accessTTLMinutes := cfg.JWT.AccessTokenTTLMinutes
	if accessTTLMinutes <= 0 {
		accessTTLMinutes = defaultAccessTTLMin
	}

	refreshTTLHours := cfg.JWT.RefreshTokenTTLHours
	if refreshTTLHours <= 0 {
		refreshTTLHours = defaultRefreshTTLHour
	}

	return &JWTTokenService{
		secret:     []byte(cfg.JWT.Secret),
		issuer:     cfg.App.Name,
		accessTTL:  time.Duration(accessTTLMinutes) * time.Minute,
		refreshTTL: time.Duration(refreshTTLHours) * time.Hour,
	}, nil
}

func (s *JWTTokenService) GenerateAccessToken(
	userID uint64,
	role, email string,
) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(s.accessTTL)

	claims := jwt.MapClaims{
		"sub":   strconv.FormatUint(userID, 10),
		"role":  role,
		"email": email,
		"iss":   s.issuer,
		"iat":   now.Unix(),
		"exp":   expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", time.Time{}, err
	}

	return signed, expiresAt, nil
}

func (s *JWTTokenService) GenerateRefreshToken() (string, string, time.Time, error) {
	buf := make([]byte, refreshTokenByteSize)
	if _, err := rand.Read(buf); err != nil {
		return "", "", time.Time{}, err
	}

	raw := base64.RawURLEncoding.EncodeToString(buf)
	expiresAt := time.Now().UTC().Add(s.refreshTTL)

	return raw, s.HashRefreshToken(raw), expiresAt, nil
}

func (s *JWTTokenService) HashRefreshToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// Verify implements authn.TokenVerifier
func (s *JWTTokenService) Verify(tokenString string) (authn.Claims, error) {
	token, err := jwt.Parse(
		tokenString,
		func(_ *jwt.Token) (any, error) { return s.secret, nil },
		jwt.WithValidMethods([]string{signingMethod}),
		jwt.WithIssuer(s.issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return authn.Claims{}, err
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return authn.Claims{}, ErrInvalidToken
	}

	sub, err := mapClaims.GetSubject()
	if err != nil {
		return authn.Claims{}, ErrInvalidToken
	}

	userID, err := strconv.ParseUint(sub, 10, 64)
	if err != nil || userID == 0 {
		return authn.Claims{}, ErrInvalidToken
	}

	role, ok := mapClaims["role"].(string)
	if !ok || role == "" {
		return authn.Claims{}, ErrInvalidToken
	}

	email, _ := mapClaims["email"].(string)

	return authn.Claims{
		UserID: userID,
		Role:   role,
		Email:  email,
	}, nil
}
