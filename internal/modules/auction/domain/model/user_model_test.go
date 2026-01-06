package model

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUserModel_ValidIdentifier(t *testing.T) {
	identifier := "user@example.com"

	user, err := NewUserModel(identifier)

	require.NoError(t, err)
	assert.Equal(t, identifier, user.Identifier())
	assert.Equal(t, uint64(0), user.ID())
	assert.False(t, user.CreatedAt().IsZero())
	assert.False(t, user.UpdatedAt().IsZero())
	assert.Equal(t, user.CreatedAt(), user.UpdatedAt())
}

func TestNewUserModel_TrimsWhitespace(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		expected   string
	}{
		{
			name:       "leading whitespace",
			identifier: "  user@example.com",
			expected:   "user@example.com",
		},
		{
			name:       "trailing whitespace",
			identifier: "user@example.com  ",
			expected:   "user@example.com",
		},
		{
			name:       "both sides whitespace",
			identifier: "  user@example.com  ",
			expected:   "user@example.com",
		},
		{
			name:       "tabs and spaces",
			identifier: "\t\nuser@example.com\t\n",
			expected:   "user@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := NewUserModel(tt.identifier)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, user.Identifier())
		})
	}
}

func TestNewUserModel_EmptyIdentifier(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
	}{
		{
			name:       "empty string",
			identifier: "",
		},
		{
			name:       "only spaces",
			identifier: "   ",
		},
		{
			name:       "only tabs",
			identifier: "\t\t",
		},
		{
			name:       "only newlines",
			identifier: "\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := NewUserModel(tt.identifier)

			require.Error(t, err)
			assert.Equal(t, "user identifier cannot be empty", err.Error())
			assert.Equal(t, UserModel{}, user)
		})
	}
}

func TestNewUserModel_IdentifierTooLong(t *testing.T) {
	// Create identifier with exactly 255 characters
	longIdentifier := strings.Repeat("a", 255)

	user, err := NewUserModel(longIdentifier)

	require.Error(t, err)
	assert.Equal(t, "user identifier cannot exceed 254 characters", err.Error())
	assert.Equal(t, UserModel{}, user)
}

func TestNewUserModel_IdentifierMaxLength(t *testing.T) {
	// Create identifier with exactly 254 characters (max allowed)
	maxIdentifier := strings.Repeat("a", 254)

	user, err := NewUserModel(maxIdentifier)

	require.NoError(t, err)
	assert.Equal(t, maxIdentifier, user.Identifier())
	assert.Equal(t, 254, len(user.Identifier()))
}

func TestNewUserModel_IdentifierWithUnicodeCharacters(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		shouldPass bool
	}{
		{
			name:       "valid unicode identifier under limit",
			identifier: "用户@example.com",
			shouldPass: true,
		},
		{
			name:       "unicode identifier at 254 runes",
			identifier: strings.Repeat("用", 254),
			shouldPass: true,
		},
		{
			name:       "unicode identifier over 254 runes",
			identifier: strings.Repeat("用", 255),
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := NewUserModel(tt.identifier)

			if tt.shouldPass {
				require.NoError(t, err)
				assert.Equal(t, tt.identifier, user.Identifier())
			} else {
				require.Error(t, err)
				assert.Equal(t, "user identifier cannot exceed 254 characters", err.Error())
			}
		})
	}
}

func TestRestoreUserModel_ValidInput(t *testing.T) {
	id := uint64(123)
	identifier := "restored@example.com"
	createdAt := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)

	user, err := RestoreUserModel(id, identifier, createdAt, updatedAt)

	require.NoError(t, err)
	assert.Equal(t, id, user.ID())
	assert.Equal(t, identifier, user.Identifier())
	assert.Equal(t, createdAt, user.CreatedAt())
	assert.Equal(t, updatedAt, user.UpdatedAt())
}

func TestRestoreUserModel_TrimsWhitespace(t *testing.T) {
	id := uint64(123)
	identifier := "  restored@example.com  "
	createdAt := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)

	user, err := RestoreUserModel(id, identifier, createdAt, updatedAt)

	require.NoError(t, err)
	assert.Equal(t, "restored@example.com", user.Identifier())
}

func TestRestoreUserModel_EmptyIdentifier(t *testing.T) {
	id := uint64(123)
	identifier := "   "
	createdAt := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)

	user, err := RestoreUserModel(id, identifier, createdAt, updatedAt)

	require.Error(t, err)
	assert.Equal(t, "user identifier cannot be empty", err.Error())
	assert.Equal(t, UserModel{}, user)
}

func TestRestoreUserModel_IdentifierTooLong(t *testing.T) {
	id := uint64(123)
	identifier := strings.Repeat("a", 255)
	createdAt := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)

	user, err := RestoreUserModel(id, identifier, createdAt, updatedAt)

	require.Error(t, err)
	assert.Equal(t, "user identifier cannot exceed 254 characters", err.Error())
	assert.Equal(t, UserModel{}, user)
}

func TestUserModel_Getters(t *testing.T) {
	id := uint64(456)
	identifier := "test@example.com"
	createdAt := time.Date(2023, 5, 10, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2023, 5, 11, 12, 0, 0, 0, time.UTC)

	user, err := RestoreUserModel(id, identifier, createdAt, updatedAt)
	require.NoError(t, err)

	assert.Equal(t, id, user.ID())
	assert.Equal(t, identifier, user.Identifier())
	assert.Equal(t, createdAt, user.CreatedAt())
	assert.Equal(t, updatedAt, user.UpdatedAt())
}
