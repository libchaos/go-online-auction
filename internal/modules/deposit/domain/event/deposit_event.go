package event

import (
	"time"

	"github.com/google/uuid"

	"auction/internal/modules/deposit/domain/model"
)

const (
	DepositHeldEventType      = "deposit_held"
	DepositReleasedEventType  = "deposit_released"
	DepositAppliedEventType   = "deposit_applied"
	DepositForfeitedEventType = "deposit_forfeited"
)

type DepositDomainEvent struct {
	eventID   string
	timestamp time.Time
}

func newDepositDomainEvent() DepositDomainEvent {
	return DepositDomainEvent{
		eventID:   uuid.New().String(),
		timestamp: time.Now().UTC(),
	}
}

func (domainEvent DepositDomainEvent) EventID() string {
	return domainEvent.eventID
}

func (domainEvent DepositDomainEvent) Timestamp() time.Time {
	return domainEvent.timestamp
}

type DepositHeldEvent struct {
	DepositDomainEvent
	depositID         uint64
	userID            uint64
	auctionID         uint64
	amount            model.MoneyModel
	currency          string
	externalReference string
}

func NewDepositHeldEvent(
	depositID uint64,
	userID uint64,
	auctionID uint64,
	amount model.MoneyModel,
	currency string,
	externalReference string,
) DepositHeldEvent {
	return DepositHeldEvent{
		DepositDomainEvent: newDepositDomainEvent(),
		depositID:          depositID,
		userID:             userID,
		auctionID:          auctionID,
		amount:             amount,
		currency:           currency,
		externalReference:  externalReference,
	}
}

func (domainEvent DepositHeldEvent) DepositID() uint64 {
	return domainEvent.depositID
}

func (domainEvent DepositHeldEvent) UserID() uint64 {
	return domainEvent.userID
}

func (domainEvent DepositHeldEvent) AuctionID() uint64 {
	return domainEvent.auctionID
}

func (domainEvent DepositHeldEvent) Amount() model.MoneyModel {
	return domainEvent.amount
}

func (domainEvent DepositHeldEvent) Currency() string {
	return domainEvent.currency
}

func (domainEvent DepositHeldEvent) ExternalReference() string {
	return domainEvent.externalReference
}

type DepositReleasedEvent struct {
	DepositDomainEvent
	depositID uint64
	userID    uint64
	auctionID uint64
	amount    model.MoneyModel
	currency  string
}

func NewDepositReleasedEvent(
	depositID uint64,
	userID uint64,
	auctionID uint64,
	amount model.MoneyModel,
	currency string,
) DepositReleasedEvent {
	return DepositReleasedEvent{
		DepositDomainEvent: newDepositDomainEvent(),
		depositID:          depositID,
		userID:             userID,
		auctionID:          auctionID,
		amount:             amount,
		currency:           currency,
	}
}

func (domainEvent DepositReleasedEvent) DepositID() uint64 {
	return domainEvent.depositID
}

func (domainEvent DepositReleasedEvent) UserID() uint64 {
	return domainEvent.userID
}

func (domainEvent DepositReleasedEvent) AuctionID() uint64 {
	return domainEvent.auctionID
}

func (domainEvent DepositReleasedEvent) Amount() model.MoneyModel {
	return domainEvent.amount
}

func (domainEvent DepositReleasedEvent) Currency() string {
	return domainEvent.currency
}

type DepositAppliedEvent struct {
	DepositDomainEvent
	depositID uint64
	userID    uint64
	auctionID uint64
	amount    model.MoneyModel
	currency  string
}

func NewDepositAppliedEvent(
	depositID uint64,
	userID uint64,
	auctionID uint64,
	amount model.MoneyModel,
	currency string,
) DepositAppliedEvent {
	return DepositAppliedEvent{
		DepositDomainEvent: newDepositDomainEvent(),
		depositID:          depositID,
		userID:             userID,
		auctionID:          auctionID,
		amount:             amount,
		currency:           currency,
	}
}

func (domainEvent DepositAppliedEvent) DepositID() uint64 {
	return domainEvent.depositID
}

func (domainEvent DepositAppliedEvent) UserID() uint64 {
	return domainEvent.userID
}

func (domainEvent DepositAppliedEvent) AuctionID() uint64 {
	return domainEvent.auctionID
}

func (domainEvent DepositAppliedEvent) Amount() model.MoneyModel {
	return domainEvent.amount
}

func (domainEvent DepositAppliedEvent) Currency() string {
	return domainEvent.currency
}

type DepositForfeitedEvent struct {
	DepositDomainEvent
	depositID uint64
	userID    uint64
	auctionID uint64
	amount    model.MoneyModel
	currency  string
}

func NewDepositForfeitedEvent(
	depositID uint64,
	userID uint64,
	auctionID uint64,
	amount model.MoneyModel,
	currency string,
) DepositForfeitedEvent {
	return DepositForfeitedEvent{
		DepositDomainEvent: newDepositDomainEvent(),
		depositID:          depositID,
		userID:             userID,
		auctionID:          auctionID,
		amount:             amount,
		currency:           currency,
	}
}

func (domainEvent DepositForfeitedEvent) DepositID() uint64 {
	return domainEvent.depositID
}

func (domainEvent DepositForfeitedEvent) UserID() uint64 {
	return domainEvent.userID
}

func (domainEvent DepositForfeitedEvent) AuctionID() uint64 {
	return domainEvent.auctionID
}

func (domainEvent DepositForfeitedEvent) Amount() model.MoneyModel {
	return domainEvent.amount
}

func (domainEvent DepositForfeitedEvent) Currency() string {
	return domainEvent.currency
}
