package command_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"auction/internal/modules/users/application/command"
	"auction/internal/modules/users/domain/enum"
	"auction/internal/modules/users/domain/errs"
	"auction/internal/modules/users/domain/model"
	"auction/tests/mocks"
)

type UpdateProfileCommandTestSuite struct {
	suite.Suite
	sut                *command.UpdateProfileCommand
	userRepositoryMock *mocks.MockUserRepository
	loggerMock         *mocks.MockLogger
	mockUser           model.UserModel
}

func (s *UpdateProfileCommandTestSuite) SetupTest() {
	s.userRepositoryMock = mocks.NewMockUserRepository(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())

	s.sut = command.NewUpdateProfileCommand(
		s.userRepositoryMock,
		s.loggerMock,
	)

	role, _ := enum.NewRoleEnum(enum.EnumRoleBidder)
	status, _ := enum.NewUserStatusEnum(enum.EnumUserStatusActive)
	passwordHash := "hashed-password"
	now := time.Now().UTC()
	s.mockUser, _ = model.RestoreUserModel(
		1, "John Doe", "john@example.com", &passwordHash, role, status, nil, nil, 1, now, now,
	)
}

func TestUpdateProfileCommandSuite(t *testing.T) {
	suite.Run(t, new(UpdateProfileCommandTestSuite))
}

func (s *UpdateProfileCommandTestSuite) TestExecute_ValidInput_UpdatesProfile() {
	// Arrange
	ctx := context.Background()
	input := command.UpdateProfileCommandInput{UserID: 1, Name: "Jane Doe"}

	s.userRepositoryMock.On("FindByID", mock.Anything, uint64(1)).Return(s.mockUser, nil)
	s.userRepositoryMock.
		On("Update", mock.Anything, mock.AnythingOfType("model.UserModel")).
		Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal("Jane Doe", output.Name)
	s.Equal(uint64(1), output.ID)
}

func (s *UpdateProfileCommandTestSuite) TestExecute_ZeroUserID_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.UpdateProfileCommandInput{UserID: 0, Name: "Jane Doe"}

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrUserIDRequired)
	s.Equal(command.UpdateProfileCommandOutput{}, output)
}

func (s *UpdateProfileCommandTestSuite) TestExecute_ShortName_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.UpdateProfileCommandInput{UserID: 1, Name: " "}

	s.userRepositoryMock.On("FindByID", mock.Anything, uint64(1)).Return(s.mockUser, nil)
	s.loggerMock.On("Error").Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrNameRequired)
	s.Equal(command.UpdateProfileCommandOutput{}, output)
}

func (s *UpdateProfileCommandTestSuite) TestExecute_UserNotFound_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.UpdateProfileCommandInput{UserID: 99, Name: "Jane Doe"}

	emptyUser := model.UserModel{}
	s.userRepositoryMock.
		On("FindByID", mock.Anything, uint64(99)).
		Return(emptyUser, errs.ErrUserNotFound)
	s.loggerMock.On("Error").Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrUserNotFound)
	s.Equal(command.UpdateProfileCommandOutput{}, output)
}
