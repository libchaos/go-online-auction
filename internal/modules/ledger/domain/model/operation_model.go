package model

import (
	"time"

	"auction/internal/modules/ledger/domain/enum"
)

type OperationModel struct {
	id                    uint64
	accountID             uint64
	counterpartyAccountID *uint64
	operationType         enum.OperationTypeEnum
	amount                uint64
	idempotencyKey        string
	status                enum.OperationStatusEnum
	reference             string
	description           string
	createdAt             time.Time
	updatedAt             time.Time
}

func NewOperation(
	accountID uint64,
	counterpartyAccountID *uint64,
	operationType enum.OperationTypeEnum,
	amount uint64,
	idempotencyKey string,
	status enum.OperationStatusEnum,
	reference string,
	description string,
) OperationModel {
	now := time.Now().UTC()

	return OperationModel{
		accountID:             accountID,
		counterpartyAccountID: counterpartyAccountID,
		operationType:         operationType,
		amount:                amount,
		idempotencyKey:        idempotencyKey,
		status:                status,
		reference:             reference,
		description:           description,
		createdAt:             now,
		updatedAt:             now,
	}
}

func RestoreOperationModel(
	id uint64,
	accountID uint64,
	counterpartyAccountID *uint64,
	operationType enum.OperationTypeEnum,
	amount uint64,
	idempotencyKey string,
	status enum.OperationStatusEnum,
	reference string,
	description string,
	createdAt time.Time,
	updatedAt time.Time,
) OperationModel {
	return OperationModel{
		id:                    id,
		accountID:             accountID,
		counterpartyAccountID: counterpartyAccountID,
		operationType:         operationType,
		amount:                amount,
		idempotencyKey:        idempotencyKey,
		status:                status,
		reference:             reference,
		description:           description,
		createdAt:             createdAt,
		updatedAt:             updatedAt,
	}
}

func (operation *OperationModel) ID() uint64 {
	return operation.id
}

func (operation *OperationModel) AccountID() uint64 {
	return operation.accountID
}

func (operation *OperationModel) CounterpartyAccountID() *uint64 {
	return operation.counterpartyAccountID
}

func (operation *OperationModel) OperationType() enum.OperationTypeEnum {
	return operation.operationType
}

func (operation *OperationModel) Amount() uint64 {
	return operation.amount
}

func (operation *OperationModel) IdempotencyKey() string {
	return operation.idempotencyKey
}

func (operation *OperationModel) Status() enum.OperationStatusEnum {
	return operation.status
}

func (operation *OperationModel) Reference() string {
	return operation.reference
}

func (operation *OperationModel) Description() string {
	return operation.description
}

func (operation *OperationModel) CreatedAt() time.Time {
	return operation.createdAt
}

func (operation *OperationModel) UpdatedAt() time.Time {
	return operation.updatedAt
}
