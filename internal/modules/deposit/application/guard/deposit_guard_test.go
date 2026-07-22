package guard_test

import (
	"context"
	"testing"
	"time"

	"auction/internal/modules/deposit/application/guard"
	"auction/internal/modules/deposit/domain/enum"
	"auction/internal/modules/deposit/domain/errs"
	"auction/internal/modules/deposit/domain/model"
	"auction/internal/modules/deposit/ports"
	"auction/tests/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const (
	guardUserID    = uint64(100)
	guardAuctionID = uint64(200)
)

type DepositGuardTestSuite struct {
	suite.Suite
	sut         *guard.DepositGuard
	repository  *mocks.MockDepositRepository
	auctionConfig *mocks.MockAuctionConfigPort
}

func (s *DepositGuardTestSuite) SetupTest() {
	s.repository = mocks.NewMockDepositRepository(s.T())
	s.auctionConfig = mocks.NewMockAuctionConfigPort(s.T())
	s.sut = guard.NewDepositGuard(s.repository, s.auctionConfig)
}

func TestDepositGuardSuite(t *testing.T) {
	suite.Run(t, new(DepositGuardTestSuite))
}

func (s *DepositGuardTestSuite) heldDeposit(amount uint64) model.DepositModel {
	deposit, err := model.RestoreDepositModel(
		1,
		guardUserID,
		guardAuctionID,
		amount,
		"CNY",
		enum.EnumDepositStatusHeld,
		"ext-ref",
		"ref",
		2,
		time.Now(),
		time.Now(),
	)
	s.Require().NoError(err)

	return deposit
}

func (s *DepositGuardTestSuite) TestEnsureEligible_NotRequired_ReturnsNil() {
	// Arrange
	s.auctionConfig.On("GetRequiredDeposit", mock.Anything, guardAuctionID).
		Return(ports.AuctionDepositConfig{Required: false}, nil)

	// Act
	err := s.sut.EnsureEligible(context.Background(), guardUserID, guardAuctionID)

	// Assert
	s.Require().NoError(err)
}

func (s *DepositGuardTestSuite) TestEnsureEligible_RequiredAndHeldSufficient_ReturnsNil() {
	// Arrange
	s.auctionConfig.On("GetRequiredDeposit", mock.Anything, guardAuctionID).
		Return(ports.AuctionDepositConfig{Required: true, Amount: model.NewMoneyModel(500)}, nil)
	s.repository.On("FindByUserAndAuction", mock.Anything, guardUserID, guardAuctionID).
		Return(s.heldDeposit(1000), nil)

	// Act
	err := s.sut.EnsureEligible(context.Background(), guardUserID, guardAuctionID)

	// Assert
	s.Require().NoError(err)
}

func (s *DepositGuardTestSuite) TestEnsureEligible_RequiredAndNotFound_ReturnsNotHeld() {
	// Arrange
	s.auctionConfig.On("GetRequiredDeposit", mock.Anything, guardAuctionID).
		Return(ports.AuctionDepositConfig{Required: true, Amount: model.NewMoneyModel(500)}, nil)
	s.repository.On("FindByUserAndAuction", mock.Anything, guardUserID, guardAuctionID).
		Return(model.DepositModel{}, errs.ErrDepositNotFound)

	// Act
	err := s.sut.EnsureEligible(context.Background(), guardUserID, guardAuctionID)

	// Assert
	s.Require().ErrorIs(err, errs.ErrDepositNotHeld)
}

func (s *DepositGuardTestSuite) TestEnsureEligible_RequiredButNotHeld_ReturnsNotHeld() {
	// Arrange
	released, err := model.RestoreDepositModel(
		1, guardUserID, guardAuctionID, 1000, "CNY",
		enum.EnumDepositStatusReleased, "ext-ref", "ref", 2, time.Now(), time.Now(),
	)
	s.Require().NoError(err)
	s.auctionConfig.On("GetRequiredDeposit", mock.Anything, guardAuctionID).
		Return(ports.AuctionDepositConfig{Required: true, Amount: model.NewMoneyModel(500)}, nil)
	s.repository.On("FindByUserAndAuction", mock.Anything, guardUserID, guardAuctionID).
		Return(released, nil)

	// Act
	err = s.sut.EnsureEligible(context.Background(), guardUserID, guardAuctionID)

	// Assert
	s.Require().ErrorIs(err, errs.ErrDepositNotHeld)
}

func (s *DepositGuardTestSuite) TestEnsureEligible_RequiredButInsufficient_ReturnsInsufficient() {
	// Arrange
	s.auctionConfig.On("GetRequiredDeposit", mock.Anything, guardAuctionID).
		Return(ports.AuctionDepositConfig{Required: true, Amount: model.NewMoneyModel(5000)}, nil)
	s.repository.On("FindByUserAndAuction", mock.Anything, guardUserID, guardAuctionID).
		Return(s.heldDeposit(1000), nil)

	// Act
	err := s.sut.EnsureEligible(context.Background(), guardUserID, guardAuctionID)

	// Assert
	s.Require().ErrorIs(err, errs.ErrDepositInsufficient)
}

func (s *DepositGuardTestSuite) TestEnsureEligible_ConfigError_Propagates() {
	// Arrange
	configErr := errs.ErrAuctionConfigNotFound
	s.auctionConfig.On("GetRequiredDeposit", mock.Anything, guardAuctionID).
		Return(ports.AuctionDepositConfig{}, configErr)

	// Act
	err := s.sut.EnsureEligible(context.Background(), guardUserID, guardAuctionID)

	// Assert
	s.Require().ErrorIs(err, configErr)
}
