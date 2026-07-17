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

type ListCategoriesQueryTestSuite struct {
	suite.Suite
	sut                    *query.ListCategoriesQuery
	categoryRepositoryMock *mocks.MockCategoryRepository
	loggerMock             *mocks.MockLogger
}

func (s *ListCategoriesQueryTestSuite) SetupTest() {
	s.categoryRepositoryMock = mocks.NewMockCategoryRepository(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())

	s.sut = query.NewListCategoriesQuery(
		s.categoryRepositoryMock,
		s.loggerMock,
	)
}

func TestListCategoriesQuerySuite(t *testing.T) {
	suite.Run(t, new(ListCategoriesQueryTestSuite))
}

func (s *ListCategoriesQueryTestSuite) TestExecute_NoParent_ReturnsRootCategories() {
	// Arrange
	ctx := context.Background()
	now := time.Now().UTC()
	category, _ := model.RestoreCategoryModel(1, "数码", nil, 0, "", 0, 1, now, now)

	s.categoryRepositoryMock.
		On("List", mock.Anything, (*uint64)(nil)).
		Return([]model.CategoryModel{category}, nil)

	// Act
	output, err := s.sut.Execute(ctx, query.ListCategoriesQueryInput{})

	// Assert
	s.Require().NoError(err)
	s.Len(output.Categories, 1)
	s.Equal("数码", output.Categories[0].Name)
}

func (s *ListCategoriesQueryTestSuite) TestExecute_WithParent_ReturnsChildren() {
	// Arrange
	ctx := context.Background()
	parentID := uint64(1)
	now := time.Now().UTC()
	category, _ := model.RestoreCategoryModel(2, "手机", &parentID, 0, "", 0, 1, now, now)

	s.categoryRepositoryMock.
		On("List", mock.Anything, &parentID).
		Return([]model.CategoryModel{category}, nil)

	// Act
	output, err := s.sut.Execute(ctx, query.ListCategoriesQueryInput{ParentID: &parentID})

	// Assert
	s.Require().NoError(err)
	s.Len(output.Categories, 1)
	s.Equal(parentID, *output.Categories[0].ParentID)
}

func (s *ListCategoriesQueryTestSuite) TestExecute_RepositoryError_ReturnsError() {
	// Arrange
	ctx := context.Background()

	s.categoryRepositoryMock.
		On("List", mock.Anything, (*uint64)(nil)).
		Return(nil, errs.ErrCategoryNotFound)
	s.loggerMock.On("Error").Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, query.ListCategoriesQueryInput{})

	// Assert
	s.Require().ErrorIs(err, errs.ErrCategoryNotFound)
	s.Equal(query.ListCategoriesQueryOutput{}, output)
}
