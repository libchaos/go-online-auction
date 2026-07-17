package handler

import (
	"net/http"
	"strconv"

	"auction/internal/modules/users/application/command"
	"auction/internal/modules/users/application/query"
	"auction/internal/modules/users/infra/http/dto"
	httperrs "auction/internal/modules/users/infra/http/errs"
	"auction/internal/shared/modules/authn"
	"auction/internal/shared/sdk/http/request"
	"auction/internal/shared/sdk/http/response"
)

const bearerTokenType = "Bearer"

type UserHandler struct {
	registerUserCommand   *command.RegisterUserCommand
	loginCommand          *command.LoginCommand
	refreshTokenCommand   *command.RefreshTokenCommand
	logoutCommand         *command.LogoutCommand
	updateProfileCommand  *command.UpdateProfileCommand
	changePasswordCommand *command.ChangePasswordCommand
	updateUserRoleCommand *command.UpdateUserRoleCommand
	getUserByIDQuery      *query.GetUserByIDQuery
	listUsersQuery        *query.ListUsersQuery
}

func NewUserHandler(
	registerUserCommand *command.RegisterUserCommand,
	loginCommand *command.LoginCommand,
	refreshTokenCommand *command.RefreshTokenCommand,
	logoutCommand *command.LogoutCommand,
	updateProfileCommand *command.UpdateProfileCommand,
	changePasswordCommand *command.ChangePasswordCommand,
	updateUserRoleCommand *command.UpdateUserRoleCommand,
	getUserByIDQuery *query.GetUserByIDQuery,
	listUsersQuery *query.ListUsersQuery,
) *UserHandler {
	return &UserHandler{
		registerUserCommand:   registerUserCommand,
		loginCommand:          loginCommand,
		refreshTokenCommand:   refreshTokenCommand,
		logoutCommand:         logoutCommand,
		updateProfileCommand:  updateProfileCommand,
		changePasswordCommand: changePasswordCommand,
		updateUserRoleCommand: updateUserRoleCommand,
		getUserByIDQuery:      getUserByIDQuery,
		listUsersQuery:        listUsersQuery,
	}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := request.ReadJSON(w, r, &req); err != nil {
		response.Error(w, httperrs.ErrInvalidRequest)
		return
	}

	output, err := h.registerUserCommand.Execute(r.Context(), command.RegisterUserCommandInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Role:     req.Role,
	})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	_ = response.JSON(w, http.StatusCreated, dto.RegisterResponse{
		ID:        output.ID,
		Name:      output.Name,
		Email:     output.Email,
		Role:      output.Role,
		Status:    output.Status,
		CreatedAt: output.CreatedAt,
	}, nil)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := request.ReadJSON(w, r, &req); err != nil {
		response.Error(w, httperrs.ErrInvalidRequest)
		return
	}

	output, err := h.loginCommand.Execute(r.Context(), command.LoginCommandInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	_ = response.JSON(w, http.StatusOK, dto.LoginResponse{
		AccessToken:  output.AccessToken,
		TokenType:    bearerTokenType,
		ExpiresAt:    output.AccessTokenExpiresAt,
		RefreshToken: output.RefreshToken,
		UserID:       output.UserID,
		Role:         output.Role,
	}, nil)
}

func (h *UserHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshRequest
	if err := request.ReadJSON(w, r, &req); err != nil {
		response.Error(w, httperrs.ErrInvalidRequest)
		return
	}

	output, err := h.refreshTokenCommand.Execute(r.Context(), command.RefreshTokenCommandInput{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	_ = response.JSON(w, http.StatusOK, dto.LoginResponse{
		AccessToken:  output.AccessToken,
		TokenType:    bearerTokenType,
		ExpiresAt:    output.AccessTokenExpiresAt,
		RefreshToken: output.RefreshToken,
		UserID:       output.UserID,
		Role:         output.Role,
	}, nil)
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req dto.LogoutRequest
	if err := request.ReadJSON(w, r, &req); err != nil {
		response.Error(w, httperrs.ErrInvalidRequest)
		return
	}

	if err := h.logoutCommand.Execute(r.Context(), command.LogoutCommandInput{
		RefreshToken: req.RefreshToken,
	}); err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	response.NoContent(w)
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	claims, ok := authn.ClaimsFromContext(r.Context())
	if !ok {
		response.Error(w, authn.ErrUnauthorized)
		return
	}

	output, err := h.getUserByIDQuery.Execute(r.Context(), query.GetUserByIDQueryInput{
		UserID: claims.UserID,
	})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	_ = response.JSON(w, http.StatusOK, toUserResponse(output), nil)
}

func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	claims, ok := authn.ClaimsFromContext(r.Context())
	if !ok {
		response.Error(w, authn.ErrUnauthorized)
		return
	}

	var req dto.UpdateProfileRequest
	if err := request.ReadJSON(w, r, &req); err != nil {
		response.Error(w, httperrs.ErrInvalidRequest)
		return
	}

	output, err := h.updateProfileCommand.Execute(r.Context(), command.UpdateProfileCommandInput{
		UserID: claims.UserID,
		Name:   req.Name,
	})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	_ = response.JSON(w, http.StatusOK, dto.UserResponse{
		ID:        output.ID,
		Name:      output.Name,
		Email:     output.Email,
		Role:      output.Role,
		Status:    output.Status,
		UpdatedAt: output.UpdatedAt,
	}, nil)
}

func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	claims, ok := authn.ClaimsFromContext(r.Context())
	if !ok {
		response.Error(w, authn.ErrUnauthorized)
		return
	}

	var req dto.ChangePasswordRequest
	if err := request.ReadJSON(w, r, &req); err != nil {
		response.Error(w, httperrs.ErrInvalidRequest)
		return
	}

	if err := h.changePasswordCommand.Execute(r.Context(), command.ChangePasswordCommandInput{
		UserID:          claims.UserID,
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	}); err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	response.NoContent(w)
}

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUserID(r)
	if err != nil {
		response.Error(w, httperrs.ErrInvalidUserID)
		return
	}

	output, err := h.getUserByIDQuery.Execute(r.Context(), query.GetUserByIDQueryInput{
		UserID: userID,
	})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	_ = response.JSON(w, http.StatusOK, toUserResponse(output), nil)
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	output, err := h.listUsersQuery.Execute(r.Context(), query.ListUsersQueryInput{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	users := make([]dto.UserResponse, 0, len(output.Users))
	for _, user := range output.Users {
		users = append(users, dto.UserResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			Role:      user.Role,
			Status:    user.Status,
			CreatedAt: user.CreatedAt,
		})
	}

	_ = response.JSON(w, http.StatusOK, dto.UserListResponse{
		Users:      users,
		TotalCount: output.TotalCount,
		Limit:      output.Limit,
		Offset:     output.Offset,
	}, nil)
}

func (h *UserHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUserID(r)
	if err != nil {
		response.Error(w, httperrs.ErrInvalidUserID)
		return
	}

	var req dto.UpdateUserRoleRequest
	if readErr := request.ReadJSON(w, r, &req); readErr != nil {
		response.Error(w, httperrs.ErrInvalidRequest)
		return
	}

	output, err := h.updateUserRoleCommand.Execute(r.Context(), command.UpdateUserRoleCommandInput{
		UserID: userID,
		Role:   req.Role,
	})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	_ = response.JSON(w, http.StatusOK, dto.UserResponse{
		ID:        output.ID,
		Name:      output.Name,
		Email:     output.Email,
		Role:      output.Role,
		Status:    output.Status,
		UpdatedAt: output.UpdatedAt,
	}, nil)
}

func parseUserID(r *http.Request) (uint64, error) {
	idString := request.Param(r, "id")
	return strconv.ParseUint(idString, 10, 64)
}

func toUserResponse(output query.GetUserByIDQueryOutput) dto.UserResponse {
	return dto.UserResponse{
		ID:        output.ID,
		Name:      output.Name,
		Email:     output.Email,
		Role:      output.Role,
		Status:    output.Status,
		CreatedAt: output.CreatedAt,
		UpdatedAt: output.UpdatedAt,
	}
}
