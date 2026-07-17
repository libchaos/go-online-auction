package mapper

import (
	"auction/internal/modules/users/domain/enum"
	"auction/internal/modules/users/domain/model"
	"auction/internal/modules/users/infra/sqlcgen"
)

type UserMapper struct{}

func NewUserMapper() *UserMapper {
	return &UserMapper{}
}

func (m *UserMapper) ToDomain(u sqlcgen.User) (model.UserModel, error) {
	role, err := enum.NewRoleEnum(string(u.Role))
	if err != nil {
		return model.UserModel{}, err
	}

	status, err := enum.NewUserStatusEnum(string(u.Status))
	if err != nil {
		return model.UserModel{}, err
	}

	return model.RestoreUserModel(
		uint64(u.ID),
		u.Name,
		u.Email,
		u.PasswordHash,
		role,
		status,
		u.OauthProvider,
		u.OauthProviderID,
		uint64(u.Version),
		u.CreatedAt,
		u.UpdatedAt,
	)
}

func (m *UserMapper) ToCreateParams(user model.UserModel) sqlcgen.CreateUserParams {
	role := user.Role()
	status := user.Status()

	return sqlcgen.CreateUserParams{
		Name:            user.Name(),
		Email:           user.Email(),
		PasswordHash:    user.PasswordHash(),
		Role:            sqlcgen.UserRole(role.String()),
		Status:          sqlcgen.UserStatus(status.String()),
		OauthProvider:   user.OauthProvider(),
		OauthProviderID: user.OauthProviderID(),
		Version:         int64(user.Version()),
		CreatedAt:       user.CreatedAt(),
		UpdatedAt:       user.UpdatedAt(),
	}
}

func (m *UserMapper) ToUpdateParams(user model.UserModel) sqlcgen.UpdateUserParams {
	role := user.Role()
	status := user.Status()

	return sqlcgen.UpdateUserParams{
		Name:            user.Name(),
		Email:           user.Email(),
		PasswordHash:    user.PasswordHash(),
		Role:            sqlcgen.UserRole(role.String()),
		Status:          sqlcgen.UserStatus(status.String()),
		OauthProvider:   user.OauthProvider(),
		OauthProviderID: user.OauthProviderID(),
		Version:         int64(user.Version()),
		UpdatedAt:       user.UpdatedAt(),
		ID:              int64(user.ID()),
	}
}
