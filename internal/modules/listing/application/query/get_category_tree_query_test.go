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

type GetCategoryTreeQueryTestSuite struct {
	suite.Suite
	sut                    *query.GetCategoryTreeQuery
	categoryRepositoryMock *mocks.MockCategoryRepository
	loggerMock             *mocks.MockLogger
	now                    time.Time
}

func (s *GetCategoryTreeQueryTestSuite) SetupTest() {
	s.categoryRepositoryMock = mocks.NewMockCategoryRepository(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())
	s.sut = query.NewGetCategoryTreeQuery(s.categoryRepositoryMock, s.loggerMock)
	s.now = time.Now().UTC()
}

func TestGetCategoryTreeQuerySuite(t *testing.T) {
	suite.Run(t, new(GetCategoryTreeQueryTestSuite))
}

func (s *GetCategoryTreeQueryTestSuite) rootCategory(id uint64, name string) model.CategoryModel {
	category, _ := model.RestoreCategoryModel(id, name, nil, 0, "/"+itoa(id), 0, 1, s.now, s.now)
	return category
}

func (s *GetCategoryTreeQueryTestSuite) childCategory(id, parentID uint64, name, path string, depth int32) model.CategoryModel {
	category, _ := model.RestoreCategoryModel(id, name, &parentID, depth, path, 0, 1, s.now, s.now)
	return category
}

func (s *GetCategoryTreeQueryTestSuite) TestExecute_FullForest_BuildsNestedTree() {
	// Arrange
	ctx := context.Background()
	root := s.rootCategory(1, "数码")
	child := s.childCategory(2, 1, "手机", "/1/2", 1)
	grandchild := s.childCategory(3, 2, "智能手机", "/1/2/3", 2)
	otherRoot := s.rootCategory(4, "服装")

	s.categoryRepositoryMock.
		On("ListAll", mock.Anything).
		Return([]model.CategoryModel{root, child, grandchild, otherRoot}, nil)

	// Act
	output, err := s.sut.Execute(ctx, query.GetCategoryTreeQueryInput{RootID: nil})

	// Assert
	s.Require().NoError(err)
	s.Len(output.Roots, 2)

	var digitalRoot *query.CategoryTreeNode
	for _, node := range output.Roots {
		if node.ID == 1 {
			digitalRoot = node
		}
	}
	s.Require().NotNil(digitalRoot)
	s.Require().Len(digitalRoot.Children, 1)
	s.Equal(uint64(2), digitalRoot.Children[0].ID)
	s.Require().Len(digitalRoot.Children[0].Children, 1)
	s.Equal(uint64(3), digitalRoot.Children[0].Children[0].ID)
}

func (s *GetCategoryTreeQueryTestSuite) TestExecute_Subtree_ReturnsSingleRootWithDescendants() {
	// Arrange
	ctx := context.Background()
	rootID := uint64(1)
	root := s.rootCategory(1, "数码")
	child := s.childCategory(2, 1, "手机", "/1/2", 1)

	s.categoryRepositoryMock.
		On("FindByID", mock.Anything, rootID).
		Return(root, nil)
	s.categoryRepositoryMock.
		On("ListDescendants", mock.Anything, rootID).
		Return([]model.CategoryModel{child}, nil)

	// Act
	output, err := s.sut.Execute(ctx, query.GetCategoryTreeQueryInput{RootID: &rootID})

	// Assert
	s.Require().NoError(err)
	s.Len(output.Roots, 1)
	s.Equal(uint64(1), output.Roots[0].ID)
	s.Require().Len(output.Roots[0].Children, 1)
	s.Equal(uint64(2), output.Roots[0].Children[0].ID)
}

func (s *GetCategoryTreeQueryTestSuite) TestExecute_RootNotFound_ReturnsError() {
	// Arrange
	ctx := context.Background()
	rootID := uint64(99)

	s.categoryRepositoryMock.
		On("FindByID", mock.Anything, rootID).
		Return(model.CategoryModel{}, errs.ErrCategoryNotFound)

	// Act
	output, err := s.sut.Execute(ctx, query.GetCategoryTreeQueryInput{RootID: &rootID})

	// Assert
	s.Require().ErrorIs(err, errs.ErrCategoryNotFound)
	s.Equal(query.GetCategoryTreeQueryOutput{}, output)
}

func itoa(v uint64) string {
	if v == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	return string(buf[i:])
}
