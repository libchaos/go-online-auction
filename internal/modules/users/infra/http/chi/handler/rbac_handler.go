package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"auction/internal/modules/users/application/command"
	"auction/internal/modules/users/application/query"
	"auction/internal/modules/users/infra/http/dto"
	httperrs "auction/internal/modules/users/infra/http/errs"
	"auction/internal/shared/sdk/http/request"
	"auction/internal/shared/sdk/http/response"
)

type RBACHandler struct {
	addPolicyCommand         *command.AddPolicyCommand
	removePolicyCommand      *command.RemovePolicyCommand
	listPoliciesQuery        *query.ListPoliciesQuery
	assignRoleCommand        *command.AssignRoleCommand
	revokeRoleCommand        *command.RevokeRoleCommand
	revokeAllRolesCommand    *command.RevokeAllRolesCommand
	listRoleAssignmentsQuery *query.ListRoleAssignmentsQuery
}

func NewRBACHandler(
	addPolicyCommand *command.AddPolicyCommand,
	removePolicyCommand *command.RemovePolicyCommand,
	listPoliciesQuery *query.ListPoliciesQuery,
	assignRoleCommand *command.AssignRoleCommand,
	revokeRoleCommand *command.RevokeRoleCommand,
	revokeAllRolesCommand *command.RevokeAllRolesCommand,
	listRoleAssignmentsQuery *query.ListRoleAssignmentsQuery,
) *RBACHandler {
	return &RBACHandler{
		addPolicyCommand:         addPolicyCommand,
		removePolicyCommand:      removePolicyCommand,
		listPoliciesQuery:        listPoliciesQuery,
		assignRoleCommand:        assignRoleCommand,
		revokeRoleCommand:        revokeRoleCommand,
		revokeAllRolesCommand:    revokeAllRolesCommand,
		listRoleAssignmentsQuery: listRoleAssignmentsQuery,
	}
}

func (h *RBACHandler) ListPolicies(w http.ResponseWriter, r *http.Request) {
	policies, err := h.listPoliciesQuery.Execute(r.Context())
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	_ = response.JSON(w, http.StatusOK, policies, nil)
}

func (h *RBACHandler) AddPolicy(w http.ResponseWriter, r *http.Request) {
	var req dto.PolicyRequest
	if readErr := request.ReadJSON(w, r, &req); readErr != nil {
		response.Error(w, httperrs.ErrInvalidRequest)
		return
	}

	if err := h.addPolicyCommand.Execute(r.Context(), command.AddPolicyCommandInput{
		Sub: req.Sub,
		Obj: req.Obj,
		Act: req.Act,
	}); err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	response.NoContent(w)
}

func (h *RBACHandler) RemovePolicy(w http.ResponseWriter, r *http.Request) {
	var req dto.PolicyRequest
	if readErr := request.ReadJSON(w, r, &req); readErr != nil {
		response.Error(w, httperrs.ErrInvalidRequest)
		return
	}

	if err := h.removePolicyCommand.Execute(r.Context(), command.RemovePolicyCommandInput{
		Sub: req.Sub,
		Obj: req.Obj,
		Act: req.Act,
	}); err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	response.NoContent(w)
}

func (h *RBACHandler) AssignRole(w http.ResponseWriter, r *http.Request) {
	var req dto.RoleAssignmentRequest
	if readErr := request.ReadJSON(w, r, &req); readErr != nil {
		response.Error(w, httperrs.ErrInvalidRequest)
		return
	}

	if err := h.assignRoleCommand.Execute(r.Context(), command.AssignRoleCommandInput{
		UserID: req.UserID,
		Role:   req.Role,
	}); err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	response.NoContent(w)
}

func (h *RBACHandler) ListRoleAssignments(w http.ResponseWriter, r *http.Request) {
	userID := parseUserIDQuery(r.URL.Query().Get("user_id"))

	views, err := h.listRoleAssignmentsQuery.Execute(r.Context(), userID)
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	_ = response.JSON(w, http.StatusOK, views, nil)
}

func (h *RBACHandler) RevokeRole(w http.ResponseWriter, r *http.Request) {
	var req dto.RoleAssignmentRequest
	if readErr := request.ReadJSON(w, r, &req); readErr != nil {
		response.Error(w, httperrs.ErrInvalidRequest)
		return
	}

	if err := h.revokeRoleCommand.Execute(r.Context(), command.RevokeRoleCommandInput{
		UserID: req.UserID,
		Role:   req.Role,
	}); err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	response.NoContent(w)
}

func (h *RBACHandler) RevokeAllRoles(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseUint(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		response.Error(w, httperrs.ErrInvalidRoleAssignment)
		return
	}

	if cmdErr := h.revokeAllRolesCommand.Execute(r.Context(), command.RevokeAllRolesCommandInput{
		UserID: userID,
	}); cmdErr != nil {
		response.Error(w, httperrs.MapDomainError(cmdErr))
		return
	}

	response.NoContent(w)
}

func parseUserIDQuery(value string) uint64 {
	if value == "" {
		return 0
	}
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0
	}
	return id
}
