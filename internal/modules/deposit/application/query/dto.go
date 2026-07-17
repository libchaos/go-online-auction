package query

import (
	"time"

	"auction/internal/modules/deposit/domain/model"
)

type DepositView struct {
	DepositID         uint64
	UserID            uint64
	AuctionID         uint64
	AmountInCents     uint64
	Currency          string
	Status            string
	ExternalReference string
	Reference         string
	Version           uint64
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func toDepositView(deposit model.DepositModel) DepositView {
	status := deposit.Status()

	return DepositView{
		DepositID:         deposit.ID(),
		UserID:            deposit.UserID(),
		AuctionID:         deposit.AuctionID(),
		AmountInCents:     deposit.Amount().AmountInCents(),
		Currency:          deposit.Currency(),
		Status:            status.String(),
		ExternalReference: deposit.ExternalReference(),
		Reference:         deposit.Reference(),
		Version:           deposit.Version(),
		CreatedAt:         deposit.CreatedAt(),
		UpdatedAt:         deposit.UpdatedAt(),
	}
}
