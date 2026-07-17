package dto

type CreateDepositRequest struct {
	AuctionID     uint64 `json:"auction_id"`
	AmountInCents uint64 `json:"amount_in_cents"`
	Currency      string `json:"currency"`
}

type CreateDepositResponse struct {
	DepositID uint64 `json:"deposit_id"`
	Status    string `json:"status"`
}

type DepositResponse struct {
	DepositID         uint64 `json:"deposit_id"`
	UserID            uint64 `json:"user_id"`
	AuctionID         uint64 `json:"auction_id"`
	AmountInCents     uint64 `json:"amount_in_cents"`
	Currency          string `json:"currency"`
	Status            string `json:"status"`
	ExternalReference string `json:"external_reference,omitempty"`
	Reference         string `json:"reference,omitempty"`
	Version           uint64 `json:"version"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

type ListDepositsResponse struct {
	Deposits []DepositResponse `json:"deposits"`
}

type EligibilityResponse struct {
	Eligible bool `json:"eligible"`
}

type ApplyDepositRequest struct {
	CaptureAmountInCents uint64 `json:"capture_amount_in_cents"`
}

type ActionDepositResponse struct {
	DepositID uint64 `json:"deposit_id"`
	Status    string `json:"status"`
}
