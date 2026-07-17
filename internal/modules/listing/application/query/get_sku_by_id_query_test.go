package query_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"auction/internal/modules/listing/application/query"
	"auction/internal/modules/listing/domain/enum"
	"auction/internal/modules/listing/domain/errs"
	"auction/internal/modules/listing/domain/model"
	"auction/tests/mocks"
)

type GetSkuByIDQueryTestSuite struct {
	suite.Suite
	sut               *query.GetSkuByIDQuery
	skuRepositoryMock *mocks.MockSkuRepository
	loggerMock        *mocks.MockLogger
	mockSku           model.SkuModel
}

func (s *GetSkuByIDQueryTestSuite) SetupTest() {
	s.skuRepositoryMock = mocks.NewMockSkuRepository(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())

	s.sut = query.NewGetSkuByIDQuery(
		s.skuRepositoryMock,
		s.loggerMock,
	)

	now := time.Now().UTC()
	publishedStatus, _ := enum.NewListingStatusEnum(enum.EnumListingStatusPublished)
	s.mockSku, _ = model.RestoreSkuModel(
		10, 1, map[string]string{"颜色": "红"}, 19900, 5, publishedStatus, 1, now, now,
	)
}

func TestGetSkuByIDQuerySuite(t *testing.T) {
	suite.Run(t, new(GetSkuByIDQueryTestSuite))
}

func (s *GetSkuByIDQueryTestSuite) TestExecute_ValidID_ReturnsSku() {
	// Arrange
	ctx := context.Background()

	s.skuRepositoryMock.On("FindByID", mock.Anything, uint64(10)).Return(s.mockSku, nil)

	// Act
	output, err := s.sut.Execute(ctx, query.GetSkuByIDQueryInput{ID: 10})

	// Assert
	s.Require().NoError(err)
	s.Equal(uint64(10), output.Sku.ID)
	s.Equal(uint64(1), output.Sku.SpuID)
	s.Equal(uint64(19900), output.Sku.PriceInCents)
	s.Equal(enum.EnumListingStatusPublished, output.Sku.Status)
}

func (s *GetSkuByIDQueryTestSuite) TestExecute_NotFound_ReturnsError() {
	// Arrange
	ctx := context.Background()

	s.skuRepositoryMock.
		On("FindByID", mock.Anything, uint64(99)).
		Return(model.SkuModel{}, errs.ErrSkuNotFound)
	s.loggerMock.On("Error").Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, query.GetSkuByIDQueryInput{ID: 99})

	// Assert
	s.Require().ErrorIs(err, errs.ErrSkuNotFound)
	s.Equal(query.GetSkuByIDQueryOutput{}, output)
}
