package uow

import "errors"

var (
	// ErrTransactionFailed is returned when transaction begin or commit fails
	ErrTransactionFailed = errors.New("transaction failed")
)
