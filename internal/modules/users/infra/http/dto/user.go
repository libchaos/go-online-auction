package dto

import "time"

type UpdateProfileRequest struct {
	Name string `json:"name"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type UpdateUserRoleRequest struct {
	Role string `json:"role"`
}

type UserResponse struct {
	ID        uint64    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitzero"`
}

type UserListResponse struct {
	Users      []UserResponse `json:"users"`
	TotalCount uint64         `json:"total_count"`
	Limit      int            `json:"limit"`
	Offset     int            `json:"offset"`
}
