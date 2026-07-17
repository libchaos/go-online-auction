package authn

import (
	"net/http"

	"auction/pkg/errs"
)

var (
	ErrUnauthorized = errs.New("AUTH_01", "Missing or invalid access token", http.StatusUnauthorized, nil)
	ErrForbidden    = errs.New("AUTH_02", "Insufficient permissions", http.StatusForbidden, nil)
)
