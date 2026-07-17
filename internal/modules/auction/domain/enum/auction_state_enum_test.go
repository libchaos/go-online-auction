package enum_test

import (
	"testing"

	"auction/internal/modules/auction/domain/enum"
	"auction/internal/modules/auction/domain/errs"
	"github.com/stretchr/testify/suite"
)

type AuctionStateEnumTestSuite struct {
	suite.Suite
}

func TestAuctionStateEnumSuite(t *testing.T) {
	suite.Run(t, new(AuctionStateEnumTestSuite))
}

func (s *AuctionStateEnumTestSuite) TestNewAuctionStateEnum_ValidDraftState_ReturnsEnum() {
	// Arrange
	value := enum.EnumAuctionStateDraft

	// Act
	result, err := enum.NewAuctionStateEnum(value)

	// Assert
	s.Require().NoError(err)
	s.Equal(value, result.String())
}

func (s *AuctionStateEnumTestSuite) TestNewAuctionStateEnum_ValidActiveState_ReturnsEnum() {
	// Arrange
	value := enum.EnumAuctionStateActive

	// Act
	result, err := enum.NewAuctionStateEnum(value)

	// Assert
	s.Require().NoError(err)
	s.Equal(value, result.String())
}

func (s *AuctionStateEnumTestSuite) TestNewAuctionStateEnum_ValidClosedState_ReturnsEnum() {
	// Arrange
	value := enum.EnumAuctionStateClosed

	// Act
	result, err := enum.NewAuctionStateEnum(value)

	// Assert
	s.Require().NoError(err)
	s.Equal(value, result.String())
}

func (s *AuctionStateEnumTestSuite) TestNewAuctionStateEnum_ValidCancelledState_ReturnsEnum() {
	// Arrange
	value := enum.EnumAuctionStateCancelled

	// Act
	result, err := enum.NewAuctionStateEnum(value)

	// Assert
	s.Require().NoError(err)
	s.Equal(value, result.String())
}

func (s *AuctionStateEnumTestSuite) TestNewAuctionStateEnum_InvalidState_ReturnsError() {
	// Arrange
	value := "invalid"

	// Act
	_, err := enum.NewAuctionStateEnum(value)

	// Assert
	s.Require().ErrorIs(err, errs.ErrInvalidAuctionState)
}

func (s *AuctionStateEnumTestSuite) TestNewAuctionStateEnum_EmptyString_ReturnsError() {
	// Arrange
	value := ""

	// Act
	_, err := enum.NewAuctionStateEnum(value)

	// Assert
	s.Require().ErrorIs(err, errs.ErrInvalidAuctionState)
}

func (s *AuctionStateEnumTestSuite) TestString_ReturnsCorrectValue() {
	// Arrange
	value := enum.EnumAuctionStateActive
	auctionState, _ := enum.NewAuctionStateEnum(value)

	// Act
	result := auctionState.String()

	// Assert
	s.Equal(value, result)
}
