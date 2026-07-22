package model_test

import (
	"testing"

	"auction/internal/modules/notification/domain/enum"
	"auction/internal/modules/notification/domain/errs"
	"auction/internal/modules/notification/domain/model"
	"github.com/stretchr/testify/suite"
)

type PreferenceModelTestSuite struct {
	suite.Suite
}

func TestPreferenceModelSuite(t *testing.T) {
	suite.Run(t, new(PreferenceModelTestSuite))
}

func (s *PreferenceModelTestSuite) TestNewNotificationPreference_ZeroUser_ReturnsError() {
	// Act
	_, err := model.NewNotificationPreference(0, nil)

	// Assert
	s.Require().ErrorIs(err, errs.ErrPreferencesUserRequired)
}

func (s *PreferenceModelTestSuite) TestNewNotificationPreference_NilSettings_AppliesDefaults() {
	// Act
	preference, err := model.NewNotificationPreference(42, nil)

	// Assert
	s.Require().NoError(err)
	s.Equal(uint64(42), preference.UserID())
	s.True(preference.IsChannelEnabled(enum.EnumNotificationCategoryPayment, enum.EnumNotificationChannelInApp))
	s.True(preference.IsChannelEnabled(enum.EnumNotificationCategoryDeposit, enum.EnumNotificationChannelInApp))
	s.True(preference.IsChannelEnabled(enum.EnumNotificationCategoryAuction, enum.EnumNotificationChannelInApp))
	s.True(preference.IsChannelEnabled(enum.EnumNotificationCategoryListing, enum.EnumNotificationChannelInApp))
	s.True(preference.IsChannelEnabled(enum.EnumNotificationCategorySystem, enum.EnumNotificationChannelInApp))
}

func (s *PreferenceModelTestSuite) TestDefaultPreferences_EnablesInAppForAllCategories() {
	// Act
	defaults := model.DefaultPreferences()

	// Assert
	s.Len(defaults, 5)
	for category, channels := range defaults {
		s.Truef(channels[enum.EnumNotificationChannelInApp], "expected in_app enabled for %s", category)
	}
}

func (s *PreferenceModelTestSuite) TestIsChannelEnabled_ExplicitlyDisabled_ReturnsFalse() {
	// Arrange
	settings := model.PreferenceSettings{
		enum.EnumNotificationCategoryAuction: {enum.EnumNotificationChannelInApp: false},
	}
	preference, err := model.NewNotificationPreference(7, settings)
	s.Require().NoError(err)

	// Act
	enabled := preference.IsChannelEnabled(enum.EnumNotificationCategoryAuction, enum.EnumNotificationChannelInApp)

	// Assert
	s.False(enabled)
}

func (s *PreferenceModelTestSuite) TestIsChannelEnabled_UnknownCategory_FallsBackToDefault() {
	// Arrange
	settings := model.PreferenceSettings{
		enum.EnumNotificationCategoryAuction: {enum.EnumNotificationChannelInApp: true},
	}
	preference, err := model.NewNotificationPreference(7, settings)
	s.Require().NoError(err)

	// Act
	enabled := preference.IsChannelEnabled(enum.EnumNotificationCategoryPayment, enum.EnumNotificationChannelInApp)

	// Assert
	s.True(enabled)
}

func (s *PreferenceModelTestSuite) TestIsChannelEnabled_UnknownChannel_ReturnsFalse() {
	// Arrange
	preference, err := model.NewNotificationPreference(7, nil)
	s.Require().NoError(err)

	// Act
	enabled := preference.IsChannelEnabled(enum.EnumNotificationCategoryPayment, "push")

	// Assert
	s.False(enabled)
}
