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

type GetSpuByIDQueryTestSuite struct {
	suite.Suite
	sut               *query.GetSpuByIDQuery
	spuRepositoryMock *mocks.MockSpuRepository
	skuRepositoryMock *mocks.MockSkuRepository
	loggerMock        *mocks.MockLogger
	mockSpu           model.SpuModel
	mockSku           model.SkuModel
}

func (s *GetSpuByIDQueryTestSuite) SetupTest() {
	s.spuRepositoryMock = mocks.NewMockSpuRepository(s.T())
	s.skuRepositoryMock = mocks.NewMockSkuRepository(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())

	s.sut = query.NewGetSpuByIDQuery(
		s.spuRepositoryMock,
		s.skuRepositoryMock,
		s.loggerMock,
	)

	now := time.Now().UTC()
	draftStatus, _ := enum.NewListingStatusEnum(enum.EnumListingStatusDraft)
	s.mockSpu, _ = model.RestoreSpuModel(1, "iPhone 15", "", 1, nil, nil, draftStatus, 1, now, now)
	s.mockSku, _ = model.RestoreSkuModel(
		10, 1, map[string]string{"颜色": "红"}, 19900, 5, draftStatus, 1, now, now,
	)
}

func TestGetSpuByIDQuerySuite(t *testing.T) {
	suite.Run(t, new(GetSpuByIDQueryTestSuite))
}

func (s *GetSpuByIDQueryTestSuite) TestExecute_ValidID_ReturnsSpuWithSkus() {
	// Arrange
	ctx := context.Background()

	s.spuRepositoryMock.On("FindByID", mock.Anything, uint64(1)).Return(s.mockSpu, nil)
	s.skuRepositoryMock.
		On("FindBySpuID", mock.Anything, uint64(1)).
		Return([]model.SkuModel{s.mockSku}, nil)

	// Act
	output, err := s.sut.Execute(ctx, query.GetSpuByIDQueryInput{ID: 1})

	// Assert
	s.Require().NoError(err)
	s.Equal(uint64(1), output.ID)
	s.Equal("iPhone 15", output.Title)
	s.Require().Len(output.Skus, 1)
	s.Equal(uint64(10), output.Skus[0].ID)
	s.Equal(map[string]string{"颜色": "红"}, output.Skus[0].SpecValues)
}

func (s *GetSpuByIDQueryTestSuite) TestExecute_SpuNotFound_ReturnsError() {
	// Arrange
	ctx := context.Background()

	s.spuRepositoryMock.
		On("FindByID", mock.Anything, uint64(99)).
		Return(model.SpuModel{}, errs.ErrSpuNotFound)
	s.loggerMock.On("Error").Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, query.GetSpuByIDQueryInput{ID: 99})

	// Assert
	s.Require().ErrorIs(err, errs.ErrSpuNotFound)
	s.Equal(query.GetSpuByIDQueryOutput{}, output)
	s.skuRepositoryMock.AssertNotCalled(s.T(), "FindBySpuID")
}
