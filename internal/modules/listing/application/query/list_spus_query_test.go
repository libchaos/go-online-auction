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
	"auction/internal/modules/listing/ports"
	"auction/tests/mocks"
)

type ListSpusQueryTestSuite struct {
	suite.Suite
	sut               *query.ListSpusQuery
	spuRepositoryMock *mocks.MockSpuRepository
	loggerMock        *mocks.MockLogger
	mockSpu           model.SpuModel
}

func (s *ListSpusQueryTestSuite) SetupTest() {
	s.spuRepositoryMock = mocks.NewMockSpuRepository(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())

	s.sut = query.NewListSpusQuery(
		s.spuRepositoryMock,
		s.loggerMock,
	)

	now := time.Now().UTC()
	draftStatus, _ := enum.NewListingStatusEnum(enum.EnumListingStatusDraft)
	s.mockSpu, _ = model.RestoreSpuModel(1, "iPhone 15", "", 1, nil, nil, draftStatus, 1, now, now)
}

func TestListSpusQuerySuite(t *testing.T) {
	suite.Run(t, new(ListSpusQueryTestSuite))
}

func (s *ListSpusQueryTestSuite) TestExecute_DefaultPagination_ReturnsSpus() {
	// Arrange
	ctx := context.Background()

	s.spuRepositoryMock.
		On("List", mock.Anything, mock.MatchedBy(func(f ports.ListSpusFilter) bool {
			return f.Limit == 20 && f.Offset == 0
		})).
		Return([]model.SpuModel{s.mockSpu}, nil)
	s.spuRepositoryMock.
		On("Count", mock.Anything, mock.AnythingOfType("ports.ListSpusFilter")).
		Return(uint64(1), nil)

	// Act
	output, err := s.sut.Execute(ctx, query.ListSpusQueryInput{})

	// Assert
	s.Require().NoError(err)
	s.Len(output.Spus, 1)
	s.Equal(uint64(1), output.TotalCount)
	s.Equal(20, output.Limit)
}

func (s *ListSpusQueryTestSuite) TestExecute_LimitAboveMax_ClampsToMax() {
	// Arrange
	ctx := context.Background()

	s.spuRepositoryMock.
		On("List", mock.Anything, mock.MatchedBy(func(f ports.ListSpusFilter) bool {
			return f.Limit == 100
		})).
		Return([]model.SpuModel{}, nil)
	s.spuRepositoryMock.
		On("Count", mock.Anything, mock.AnythingOfType("ports.ListSpusFilter")).
		Return(uint64(0), nil)

	// Act
	output, err := s.sut.Execute(ctx, query.ListSpusQueryInput{Limit: 500})

	// Assert
	s.Require().NoError(err)
	s.Equal(100, output.Limit)
}

func (s *ListSpusQueryTestSuite) TestExecute_InvalidStatus_ReturnsError() {
	// Arrange
	ctx := context.Background()
	invalidStatus := "archived"

	s.loggerMock.On("Error").Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, query.ListSpusQueryInput{Status: &invalidStatus})

	// Assert
	s.Require().ErrorIs(err, errs.ErrInvalidListingStatus)
	s.Equal(query.ListSpusQueryOutput{}, output)
}
