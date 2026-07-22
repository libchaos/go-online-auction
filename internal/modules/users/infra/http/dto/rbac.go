package dto

// PolicyRequest is the body for adding or removing an RBAC policy rule.
type PolicyRequest struct {
	Sub string `json:"sub"`
	Obj string `json:"obj"`
	Act string `json:"act"`
}
