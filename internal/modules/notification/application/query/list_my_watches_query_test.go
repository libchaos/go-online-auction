package query_test

import (
	"context"
	"testing"

	"auction/internal/modules/notification/application/query"
	"auction/internal/modules/notification/domain/model"
	"auction/tests/mocks"
	"github.com/stretchr/testify/suite"
)

type ListMyWatchesQueryTestSuite struct {
	suite.Suite
	sut            *query.ListMyWatchesQuery
	watchlistsMock *mocks.MockWatchlistRepository
}

func (s *ListMyWatchesQueryTestSuite) SetupTest() {
	s.watchlistsMock = mocks.NewMockWatchlistRepository(s.T())
	s.sut = query.NewListMyWatchesQuery(s.watchlistsMock)
}

func TestListMyWatchesQuerySuite(t *testing.T) {
	suite.Run(t, new(ListMyWatchesQueryTestSuite))
}

func (s *ListMyWatchesQueryTestSuite) TestExecute_DefaultLimit_IsClampedToTwenty() {
	ctx := context.Background()
	s.watchlistsMock.EXPECT().ListByUser(ctx, uint64(100), 20, 0).
		Return([]model.Watchlist{}, nil)

	output, err := s.sut.Execute(ctx, query.ListMyWatchesQueryInput{UserID: 100, Limit: 0, Offset: 0})

	s.Require().NoError(err)
	s.Equal(20, output.Limit)
	s.Equal(0, output.Offset)
}

func (s *ListMyWatchesQueryTestSuite) TestExecute_OverMaxLimit_IsClampedToHundred() {
	ctx := context.Background()
	s.watchlistsMock.EXPECT().ListByUser(ctx, uint64(100), 100, 0).
		Return([]model.Watchlist{}, nil)

	output, err := s.sut.Execute(ctx, query.ListMyWatchesQueryInput{UserID: 100, Limit: 999, Offset: 0})

	s.Require().NoError(err)
	s.Equal(100, output.Limit)
}

func (s *ListMyWatchesQueryTestSuite) TestExecute_NegativeOffset_IsClampedToZero() {
	ctx := context.Background()
	s.watchlistsMock.EXPECT().ListByUser(ctx, uint64(100), 20, 0).
		Return([]model.Watchlist{}, nil)

	output, err := s.sut.Execute(ctx, query.ListMyWatchesQueryInput{UserID: 100, Limit: 0, Offset: -5})

	s.Require().NoError(err)
	s.Equal(0, output.Offset)
}
