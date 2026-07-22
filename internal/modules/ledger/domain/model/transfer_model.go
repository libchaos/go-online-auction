package model

import (
	"time"
)

type TransferModel struct {
	id             uint64
	fromAccountID  uint64
	toAccountID    uint64
	amount         uint64
	idempotencyKey string
	createdAt      time.Time
}

func RestoreTransferModel(
	id uint64,
	fromAccountID uint64,
	toAccountID uint64,
	amount uint64,
	idempotencyKey string,
	createdAt time.Time,
) TransferModel {
	return TransferModel{
		id:             id,
		fromAccountID:  fromAccountID,
		toAccountID:    toAccountID,
		amount:         amount,
		idempotencyKey: idempotencyKey,
		createdAt:      createdAt,
	}
}

func (transfer *TransferModel) ID() uint64 {
	return transfer.id
}

func (transfer *TransferModel) FromAccountID() uint64 {
	return transfer.fromAccountID
}

func (transfer *TransferModel) ToAccountID() uint64 {
	return transfer.toAccountID
}

func (transfer *TransferModel) Amount() uint64 {
	return transfer.amount
}

func (transfer *TransferModel) IdempotencyKey() string {
	return transfer.idempotencyKey
}

func (transfer *TransferModel) CreatedAt() time.Time {
	return transfer.createdAt
}
