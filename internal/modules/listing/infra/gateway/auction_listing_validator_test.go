package gateway_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"auction/internal/modules/listing/domain/enum"
	"auction/internal/modules/listing/domain/errs"
	"auction/internal/modules/listing/domain/model"
	"auction/internal/modules/listing/infra/gateway"
	"auction/tests/mocks"
)

type AuctionListingValidatorTestSuite struct {
	suite.Suite
	sut               *gateway.AuctionListingValidator
	skuRepositoryMock *mocks.MockSkuRepository
	spuRepositoryMock *mocks.MockSpuRepository
}

func (s *AuctionListingValidatorTestSuite) SetupTest() {
	s.skuRepositoryMock = mocks.NewMockSkuRepository(s.T())
	s.spuRepositoryMock = mocks.NewMockSpuRepository(s.T())

	s.sut = gateway.NewAuctionListingValidator(
		s.skuRepositoryMock,
		s.spuRepositoryMock,
	)
}

func TestAuctionListingValidatorSuite(t *testing.T) {
	suite.Run(t, new(AuctionListingValidatorTestSuite))
}

func (s *AuctionListingValidatorTestSuite) buildSku(status string, quantity uint64) model.SkuModel {
	statusEnum, _ := enum.NewListingStatusEnum(status)
	now := time.Now().UTC()
	sku, _ := model.RestoreSkuModel(
		10, 1, map[string]string{"颜色": "红"}, 19900, quantity, statusEnum, 1, now, now,
	)
	return sku
}

func (s *AuctionListingValidatorTestSuite) buildSpu(status string) model.SpuModel {
	statusEnum, _ := enum.NewListingStatusEnum(status)
	now := time.Now().UTC()
	spu, _ := model.RestoreSpuModel(1, "iPhone 15", "", 1, nil, nil, statusEnum, 1, now, now)
	return spu
}

func (s *AuctionListingValidatorTestSuite) TestIsAuctionable_PublishedSkuAndSpu_ReturnsTrue() {
	// Arrange
	ctx := context.Background()
	sku := s.buildSku(enum.EnumListingStatusPublished, 5)
	spu := s.buildSpu(enum.EnumListingStatusPublished)

	s.skuRepositoryMock.On("FindByID", mock.Anything, uint64(10)).Return(sku, nil)
	s.spuRepositoryMock.On("FindByID", mock.Anything, uint64(1)).Return(spu, nil)

	// Act
	auctionable, err := s.sut.IsAuctionable(ctx, 10)

	// Assert
	s.Require().NoError(err)
	s.True(auctionable)
}

func (s *AuctionListingValidatorTestSuite) TestIsAuctionable_SkuNotFound_ReturnsFalse() {
	// Arrange
	ctx := context.Background()

	s.skuRepositoryMock.
		On("FindByID", mock.Anything, uint64(99)).
		Return(model.SkuModel{}, errs.ErrSkuNotFound)

	// Act
	auctionable, err := s.sut.IsAuctionable(ctx, 99)

	// Assert
	s.Require().NoError(err)
	s.False(auctionable)
}

func (s *AuctionListingValidatorTestSuite) TestIsAuctionable_DraftSku_ReturnsFalse() {
	// Arrange
	ctx := context.Background()
	sku := s.buildSku(enum.EnumListingStatusDraft, 5)

	s.skuRepositoryMock.On("FindByID", mock.Anything, uint64(10)).Return(sku, nil)

	// Act
	auctionable, err := s.sut.IsAuctionable(ctx, 10)

	// Assert
	s.Require().NoError(err)
	s.False(auctionable)
	s.spuRepositoryMock.AssertNotCalled(s.T(), "FindByID")
}

func (s *AuctionListingValidatorTestSuite) TestIsAuctionable_ZeroQuantity_ReturnsFalse() {
	// Arrange
	ctx := context.Background()
	sku := s.buildSku(enum.EnumListingStatusPublished, 0)

	s.skuRepositoryMock.On("FindByID", mock.Anything, uint64(10)).Return(sku, nil)

	// Act
	auctionable, err := s.sut.IsAuctionable(ctx, 10)

	// Assert
	s.Require().NoError(err)
	s.False(auctionable)
}

func (s *AuctionListingValidatorTestSuite) TestIsAuctionable_SpuNotPublished_ReturnsFalse() {
	// Arrange
	ctx := context.Background()
	sku := s.buildSku(enum.EnumListingStatusPublished, 5)
	spu := s.buildSpu(enum.EnumListingStatusOffShelf)

	s.skuRepositoryMock.On("FindByID", mock.Anything, uint64(10)).Return(sku, nil)
	s.spuRepositoryMock.On("FindByID", mock.Anything, uint64(1)).Return(spu, nil)

	// Act
	auctionable, err := s.sut.IsAuctionable(ctx, 10)

	// Assert
	s.Require().NoError(err)
	s.False(auctionable)
}

func (s *AuctionListingValidatorTestSuite) TestIsAuctionable_RepositoryError_ReturnsError() {
	// Arrange
	ctx := context.Background()
	repositoryErr := errors.New("repository error")

	s.skuRepositoryMock.
		On("FindByID", mock.Anything, uint64(10)).
		Return(model.SkuModel{}, repositoryErr)

	// Act
	auctionable, err := s.sut.IsAuctionable(ctx, 10)

	// Assert
	s.Require().ErrorIs(err, repositoryErr)
	s.False(auctionable)
}
