package model

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewListingModel_ValidInput(t *testing.T) {
	ownerUserID := uint64(123)
	title := "Vintage Watch"
	description := "A beautiful vintage watch from the 1950s"

	listing, err := NewListingModel(ownerUserID, title, description)

	require.NoError(t, err)
	assert.Equal(t, uint64(0), listing.ID())
	assert.Equal(t, ownerUserID, listing.OwnerUserID())
	assert.Equal(t, title, listing.Title())
	assert.Equal(t, description, listing.Description())
	assert.False(t, listing.CreatedAt().IsZero())
	assert.False(t, listing.UpdatedAt().IsZero())
	assert.Equal(t, listing.CreatedAt(), listing.UpdatedAt())
}

func TestNewListingModel_EmptyDescription(t *testing.T) {
	ownerUserID := uint64(123)
	title := "Vintage Watch"
	description := ""

	listing, err := NewListingModel(ownerUserID, title, description)

	require.NoError(t, err)
	assert.Equal(t, "", listing.Description())
	assert.Equal(t, title, listing.Title())
}

func TestNewListingModel_TitleTrimsWhitespace(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		expected string
	}{
		{
			name:     "leading whitespace",
			title:    "  Vintage Watch",
			expected: "Vintage Watch",
		},
		{
			name:     "trailing whitespace",
			title:    "Vintage Watch  ",
			expected: "Vintage Watch",
		},
		{
			name:     "both sides whitespace",
			title:    "  Vintage Watch  ",
			expected: "Vintage Watch",
		},
		{
			name:     "tabs and newlines",
			title:    "\t\nVintage Watch\t\n",
			expected: "Vintage Watch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listing, err := NewListingModel(123, tt.title, "description")

			require.NoError(t, err)
			assert.Equal(t, tt.expected, listing.Title())
		})
	}
}

func TestNewListingModel_DescriptionTrimsWhitespace(t *testing.T) {
	tests := []struct {
		name        string
		description string
		expected    string
	}{
		{
			name:        "leading whitespace",
			description: "  A beautiful vintage watch",
			expected:    "A beautiful vintage watch",
		},
		{
			name:        "trailing whitespace",
			description: "A beautiful vintage watch  ",
			expected:    "A beautiful vintage watch",
		},
		{
			name:        "both sides whitespace",
			description: "  A beautiful vintage watch  ",
			expected:    "A beautiful vintage watch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listing, err := NewListingModel(123, "Title", tt.description)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, listing.Description())
		})
	}
}

func TestNewListingModel_EmptyTitle(t *testing.T) {
	tests := []struct {
		name  string
		title string
	}{
		{
			name:  "empty string",
			title: "",
		},
		{
			name:  "only spaces",
			title: "   ",
		},
		{
			name:  "only tabs",
			title: "\t\t",
		},
		{
			name:  "only newlines",
			title: "\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listing, err := NewListingModel(123, tt.title, "description")

			require.Error(t, err)
			assert.Equal(t, "listing title cannot be empty", err.Error())
			assert.Equal(t, ListingModel{}, listing)
		})
	}
}

func TestNewListingModel_TitleTooLong(t *testing.T) {
	// Create title with exactly 201 characters
	longTitle := strings.Repeat("a", 201)

	listing, err := NewListingModel(123, longTitle, "description")

	require.Error(t, err)
	assert.Equal(t, "listing title cannot exceed 200 characters", err.Error())
	assert.Equal(t, ListingModel{}, listing)
}

func TestNewListingModel_TitleMaxLength(t *testing.T) {
	// Create title with exactly 200 characters (max allowed)
	maxTitle := strings.Repeat("a", 200)

	listing, err := NewListingModel(123, maxTitle, "description")

	require.NoError(t, err)
	assert.Equal(t, maxTitle, listing.Title())
	assert.Equal(t, 200, len(listing.Title()))
}

func TestNewListingModel_TitleWithUnicodeCharacters(t *testing.T) {
	tests := []struct {
		name       string
		title      string
		shouldPass bool
	}{
		{
			name:       "valid unicode title under limit",
			title:      "古董手表 Vintage Watch",
			shouldPass: true,
		},
		{
			name:       "unicode title at 200 runes",
			title:      strings.Repeat("表", 200),
			shouldPass: true,
		},
		{
			name:       "unicode title over 200 runes",
			title:      strings.Repeat("表", 201),
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listing, err := NewListingModel(123, tt.title, "description")

			if tt.shouldPass {
				require.NoError(t, err)
				assert.Equal(t, tt.title, listing.Title())
			} else {
				require.Error(t, err)
				assert.Equal(t, "listing title cannot exceed 200 characters", err.Error())
			}
		})
	}
}

func TestNewListingModel_DescriptionTooLong(t *testing.T) {
	// Create description with exactly 2001 characters
	longDescription := strings.Repeat("a", 2001)

	listing, err := NewListingModel(123, "Title", longDescription)

	require.Error(t, err)
	assert.Equal(t, "listing description cannot exceed 2000 characters", err.Error())
	assert.Equal(t, ListingModel{}, listing)
}

func TestNewListingModel_DescriptionMaxLength(t *testing.T) {
	// Create description with exactly 2000 characters (max allowed)
	maxDescription := strings.Repeat("a", 2000)

	listing, err := NewListingModel(123, "Title", maxDescription)

	require.NoError(t, err)
	assert.Equal(t, maxDescription, listing.Description())
	assert.Equal(t, 2000, len(listing.Description()))
}

func TestNewListingModel_DescriptionWithUnicodeCharacters(t *testing.T) {
	tests := []struct {
		name        string
		description string
		shouldPass  bool
	}{
		{
			name:        "valid unicode description under limit",
			description: "这是一个美丽的古董手表 This is a beautiful vintage watch",
			shouldPass:  true,
		},
		{
			name:        "unicode description at 2000 runes",
			description: strings.Repeat("表", 2000),
			shouldPass:  true,
		},
		{
			name:        "unicode description over 2000 runes",
			description: strings.Repeat("表", 2001),
			shouldPass:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listing, err := NewListingModel(123, "Title", tt.description)

			if tt.shouldPass {
				require.NoError(t, err)
				assert.Equal(t, tt.description, listing.Description())
			} else {
				require.Error(t, err)
				assert.Equal(t, "listing description cannot exceed 2000 characters", err.Error())
			}
		})
	}
}

func TestRestoreListingModel_ValidInput(t *testing.T) {
	id := uint64(456)
	ownerUserID := uint64(123)
	title := "Restored Listing"
	description := "This is a restored listing"
	createdAt := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)

	listing, err := RestoreListingModel(id, ownerUserID, title, description, createdAt, updatedAt)

	require.NoError(t, err)
	assert.Equal(t, id, listing.ID())
	assert.Equal(t, ownerUserID, listing.OwnerUserID())
	assert.Equal(t, title, listing.Title())
	assert.Equal(t, description, listing.Description())
	assert.Equal(t, createdAt, listing.CreatedAt())
	assert.Equal(t, updatedAt, listing.UpdatedAt())
}

func TestRestoreListingModel_TitleTrimsWhitespace(t *testing.T) {
	id := uint64(456)
	ownerUserID := uint64(123)
	title := "  Restored Listing  "
	description := "description"
	createdAt := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)

	listing, err := RestoreListingModel(id, ownerUserID, title, description, createdAt, updatedAt)

	require.NoError(t, err)
	assert.Equal(t, "Restored Listing", listing.Title())
}

func TestRestoreListingModel_EmptyTitle(t *testing.T) {
	id := uint64(456)
	ownerUserID := uint64(123)
	title := "   "
	description := "description"
	createdAt := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)

	listing, err := RestoreListingModel(id, ownerUserID, title, description, createdAt, updatedAt)

	require.Error(t, err)
	assert.Equal(t, "listing title cannot be empty", err.Error())
	assert.Equal(t, ListingModel{}, listing)
}

func TestRestoreListingModel_TitleTooLong(t *testing.T) {
	id := uint64(456)
	ownerUserID := uint64(123)
	title := strings.Repeat("a", 201)
	description := "description"
	createdAt := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)

	listing, err := RestoreListingModel(id, ownerUserID, title, description, createdAt, updatedAt)

	require.Error(t, err)
	assert.Equal(t, "listing title cannot exceed 200 characters", err.Error())
	assert.Equal(t, ListingModel{}, listing)
}

func TestRestoreListingModel_DescriptionTooLong(t *testing.T) {
	id := uint64(456)
	ownerUserID := uint64(123)
	title := "Title"
	description := strings.Repeat("a", 2001)
	createdAt := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)

	listing, err := RestoreListingModel(id, ownerUserID, title, description, createdAt, updatedAt)

	require.Error(t, err)
	assert.Equal(t, "listing description cannot exceed 2000 characters", err.Error())
	assert.Equal(t, ListingModel{}, listing)
}

func TestListingModel_Getters(t *testing.T) {
	id := uint64(789)
	ownerUserID := uint64(456)
	title := "Test Listing"
	description := "Test Description"
	createdAt := time.Date(2023, 5, 10, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2023, 5, 11, 12, 0, 0, 0, time.UTC)

	listing, err := RestoreListingModel(id, ownerUserID, title, description, createdAt, updatedAt)
	require.NoError(t, err)

	assert.Equal(t, id, listing.ID())
	assert.Equal(t, ownerUserID, listing.OwnerUserID())
	assert.Equal(t, title, listing.Title())
	assert.Equal(t, description, listing.Description())
	assert.Equal(t, createdAt, listing.CreatedAt())
	assert.Equal(t, updatedAt, listing.UpdatedAt())
}
