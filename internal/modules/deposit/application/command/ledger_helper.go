package command

import (
	"context"
	"strconv"

	ledgerports "auction/internal/modules/ledger/ports"
)

const platformAccountOwner = "platform"

func buyerLedgerAccountID(
	ctx context.Context,
	ledger ledgerports.LedgerRepository,
	userID uint64,
	currency string,
) (uint64, error) {
	owner := strconv.FormatUint(userID, 10)
	account, err := ledger.GetOrCreateAccountByOwner(ctx, owner, currency)
	if err != nil {
		return 0, err
	}

	return account.ID(), nil
}

func platformLedgerAccountID(
	ctx context.Context,
	ledger ledgerports.LedgerRepository,
	currency string,
) (uint64, error) {
	account, err := ledger.GetOrCreateAccountByOwner(ctx, platformAccountOwner, currency)
	if err != nil {
		return 0, err
	}

	return account.ID(), nil
}

func depositLedgerIdempotencyKey(action string, depositID uint64) string {
	return "deposit:" + strconv.FormatUint(depositID, 10) + ":" + action
}
