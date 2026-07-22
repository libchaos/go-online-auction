package mapper

import (
	"auction/internal/modules/payment/domain/enum"
	"auction/internal/modules/payment/domain/model"
	"auction/internal/modules/payment/infra/sqlcgen"
)

// PaymentMapper converts between the payment domain models and the sqlc
// generated persistence structs.
type PaymentMapper struct{}

func NewPaymentMapper() *PaymentMapper {
	return &PaymentMapper{}
}

func (mapper *PaymentMapper) ToPaymentDomain(payment sqlcgen.Payment) (model.PaymentModel, error) {
	status, err := enum.ValidatePaymentStatus(payment.Status)
	if err != nil {
		return model.PaymentModel{}, err
	}

	return model.RestorePaymentModel(
		uint64(payment.ID),
		uint64(payment.UserID),
		uint64(payment.AmountCents),
		payment.Currency,
		status,
		payment.OutTradeNo,
		toString(payment.QrCodeUrl),
		toString(payment.AlipayTradeNo),
		uint64(payment.Version),
		payment.CreatedAt,
		payment.UpdatedAt,
	)
}

func (mapper *PaymentMapper) ToCreatePaymentParams(payment model.PaymentModel) sqlcgen.CreatePaymentParams {
	return sqlcgen.CreatePaymentParams{
		UserID:      int64(payment.UserID()),
		AmountCents: int64(payment.AmountInCents()),
		Currency:    payment.Currency(),
		Status:      string(payment.Status()),
		OutTradeNo:  payment.OutTradeNo(),
		QrCodeUrl:   toNullableString(payment.QRCodeURL()),
		Version:     int64(payment.Version()),
	}
}

func (mapper *PaymentMapper) ToUpdatePaymentParams(payment model.PaymentModel) sqlcgen.UpdatePaymentParams {
	return sqlcgen.UpdatePaymentParams{
		Status:        string(payment.Status()),
		AlipayTradeNo: toNullableString(payment.AlipayTradeNo()),
		Version:       int64(payment.Version()),
		ID:            int64(payment.ID()),
		Version_2:     int64(payment.Version()) - 1,
	}
}

func (mapper *PaymentMapper) ToWithdrawalDomain(withdrawal sqlcgen.Withdrawal) (model.WithdrawalModel, error) {
	status, err := enum.ValidateWithdrawalStatus(withdrawal.Status)
	if err != nil {
		return model.WithdrawalModel{}, err
	}

	return model.RestoreWithdrawalModel(
		uint64(withdrawal.ID),
		uint64(withdrawal.UserID),
		uint64(withdrawal.LedgerAccountID),
		withdrawal.AlipayAccount,
		withdrawal.AlipayRealName,
		uint64(withdrawal.AmountCents),
		withdrawal.Currency,
		status,
		withdrawal.OutBizNo,
		toString(withdrawal.FrozenOpID),
		toString(withdrawal.AlipayOrderID),
		toString(withdrawal.FailReason),
		uint64(withdrawal.Version),
		withdrawal.CreatedAt,
		withdrawal.UpdatedAt,
	)
}

func (mapper *PaymentMapper) ToCreateWithdrawalParams(withdrawal model.WithdrawalModel) sqlcgen.CreateWithdrawalParams {
	return sqlcgen.CreateWithdrawalParams{
		UserID:          int64(withdrawal.UserID()),
		LedgerAccountID: int64(withdrawal.LedgerAccountID()),
		AlipayAccount:   withdrawal.AlipayAccount(),
		AlipayRealName:  withdrawal.AlipayRealName(),
		AmountCents:     int64(withdrawal.AmountInCents()),
		Currency:        withdrawal.Currency(),
		Status:          string(withdrawal.Status()),
		OutBizNo:        withdrawal.OutBizNo(),
		FrozenOpID:      toNullableString(withdrawal.FrozenOpID()),
		Version:         int64(withdrawal.Version()),
	}
}

func (mapper *PaymentMapper) ToUpdateWithdrawalParams(withdrawal model.WithdrawalModel) sqlcgen.UpdateWithdrawalParams {
	return sqlcgen.UpdateWithdrawalParams{
		Status:        string(withdrawal.Status()),
		AlipayOrderID: toNullableString(withdrawal.AlipayOrderID()),
		FailReason:    toNullableString(withdrawal.FailReason()),
		Version:       int64(withdrawal.Version()),
		ID:            int64(withdrawal.ID()),
		Version_2:     int64(withdrawal.Version()) - 1,
	}
}

func toString(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}

func toNullableString(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}
