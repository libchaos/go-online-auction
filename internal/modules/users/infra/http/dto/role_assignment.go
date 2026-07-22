package dto

// RoleAssignmentRequest assigns a role to a user through the g (grouping)
// relation. It is the body for adding or removing a single user->role binding.
type RoleAssignmentRequest struct {
	UserID uint64 `json:"user_id"`
	Role   string `json:"role"`
}
