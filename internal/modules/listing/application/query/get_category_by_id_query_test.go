package query_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"auction/internal/modules/listing/application/query"
	"auction/internal/modules/listing/domain/errs"
	"auction/internal/modules/listing/domain/model"
	"auction/tests/mocks"
)

type GetCategoryByIDQueryTestSuite struct {
	suite.Suite
	sut                    *query.GetCategoryByIDQuery
	categoryRepositoryMock *mocks.MockCategoryRepository
	loggerMock             *mocks.MockLogger
}

func (s *GetCategoryByIDQueryTestSuite) SetupTest() {
	s.categoryRepositoryMock = mocks.NewMockCategoryRepository(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())

	s.sut = query.NewGetCategoryByIDQuery(
		s.categoryRepositoryMock,
		s.loggerMock,
	)
}

func TestGetCategoryByIDQuerySuite(t *testing.T) {
	suite.Run(t, new(GetCategoryByIDQueryTestSuite))
}

func (s *GetCategoryByIDQueryTestSuite) TestExecute_ValidID_ReturnsCategory() {
	// Arrange
	ctx := context.Background()
	now := time.Now().UTC()
	category, _ := model.RestoreCategoryModel(1, "数码", nil, 0, "", 0, 1, now, now)

	s.categoryRepositoryMock.On("FindByID", mock.Anything, uint64(1)).Return(category, nil)

	// Act
	output, err := s.sut.Execute(ctx, query.GetCategoryByIDQueryInput{ID: 1})

	// Assert
	s.Require().NoError(err)
	s.Equal(uint64(1), output.Category.ID)
	s.Equal("数码", output.Category.Name)
}

func (s *GetCategoryByIDQueryTestSuite) TestExecute_NotFound_ReturnsError() {
	// Arrange
	ctx := context.Background()

	s.categoryRepositoryMock.
		On("FindByID", mock.Anything, uint64(99)).
		Return(model.CategoryModel{}, errs.ErrCategoryNotFound)
	s.loggerMock.On("Error").Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, query.GetCategoryByIDQueryInput{ID: 99})

	// Assert
	s.Require().ErrorIs(err, errs.ErrCategoryNotFound)
	s.Equal(query.GetCategoryByIDQueryOutput{}, output)
}
