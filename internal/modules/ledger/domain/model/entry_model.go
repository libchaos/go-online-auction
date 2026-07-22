package model

import (
	"time"

	"auction/internal/modules/ledger/domain/enum"
)

type EntryModel struct {
	id          int64
	accountID   uint64
	amount      int64
	entryType   enum.EntryTypeEnum
	operationID *uint64
	createdAt   time.Time
}

func RestoreEntryModel(
	id int64,
	accountID uint64,
	amount int64,
	entryType enum.EntryTypeEnum,
	operationID *uint64,
	createdAt time.Time,
) EntryModel {
	return EntryModel{
		id:          id,
		accountID:   accountID,
		amount:      amount,
		entryType:   entryType,
		operationID: operationID,
		createdAt:   createdAt,
	}
}

func (entry *EntryModel) ID() int64 {
	return entry.id
}

func (entry *EntryModel) AccountID() uint64 {
	return entry.accountID
}

func (entry *EntryModel) Amount() int64 {
	return entry.amount
}

func (entry *EntryModel) EntryType() enum.EntryTypeEnum {
	return entry.entryType
}

func (entry *EntryModel) OperationID() *uint64 {
	return entry.operationID
}

func (entry *EntryModel) CreatedAt() time.Time {
	return entry.createdAt
}
