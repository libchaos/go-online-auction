package alipay

import (
	"context"
	"fmt"
	"net/url"

	alipaysdk "github.com/smartwalle/alipay/v3"

	"auction/internal/modules/payment/domain/errs"
	"auction/internal/modules/payment/ports"
	"auction/internal/shared/modules/config"
	"auction/internal/shared/modules/logger"
)

const (
	transferProductCode = "TRANS_ACCOUNT_NO_PWD"
	transferBizScene    = "DIRECT_TRANSFER"
	identityTypeLogonID = "ALIPAY_LOGON_ID"
)

var _ ports.AlipayPort = (*SDKAlipayAdapter)(nil)

type SDKAlipayAdapter struct {
	client *alipaysdk.Client
	logger logger.Logger
}

// NewSDKAlipayAdapter constructs the real Alipay gateway adapter backed by the
// smartwalle SDK. It returns an error when the client cannot be initialised or
// the configured public key cannot be loaded.
func NewSDKAlipayAdapter(cfg config.Alipay, logger logger.Logger) (ports.AlipayPort, error) {
	client, err := alipaysdk.New(cfg.AppID, cfg.AppPrivateKey, cfg.IsProductionGateway())
	if err != nil {
		return nil, fmt.Errorf("payment: init alipay client: %w", err)
	}

	if cfg.PublicKey != "" {
		if loadErr := client.LoadAliPayPublicKey(cfg.PublicKey); loadErr != nil {
			return nil, fmt.Errorf("payment: load alipay public key: %w", loadErr)
		}
	}

	return &SDKAlipayAdapter{client: client, logger: logger}, nil
}

func (adapter *SDKAlipayAdapter) CreateFaceToFacePayment(
	ctx context.Context,
	input ports.FaceToFaceInput,
) (ports.FaceToFaceOutput, error) {
	param := alipaysdk.TradePreCreate{
		Trade: alipaysdk.Trade{
			Subject:     input.Subject,
			OutTradeNo:  input.OutTradeNo,
			TotalAmount: centsToYuan(input.AmountInCents),
			NotifyURL:   input.NotifyURL,
		},
	}

	result, err := adapter.client.TradePreCreate(ctx, param)
	if err != nil {
		return ports.FaceToFaceOutput{}, fmt.Errorf("payment: alipay trade precreate: %w", err)
	}
	if !result.IsSuccess() {
		return ports.FaceToFaceOutput{}, fmt.Errorf(
			"payment: alipay trade precreate failed: %s %s",
			result.Code,
			result.SubMsg,
		)
	}

	return ports.FaceToFaceOutput{
		QRCodeURL:  result.QRCode,
		OutTradeNo: result.OutTradeNo,
	}, nil
}

func (adapter *SDKAlipayAdapter) VerifyNotify(
	ctx context.Context,
	params map[string]string,
) (ports.NotifyResult, error) {
	values := url.Values{}
	for key, value := range params {
		values.Set(key, value)
	}

	notification, err := adapter.client.DecodeNotification(ctx, values)
	if err != nil {
		return ports.NotifyResult{}, fmt.Errorf("%w: %w", errs.ErrAlipayNotifyVerification, err)
	}

	return ports.NotifyResult{
		TradeNo:     notification.TradeNo,
		OutTradeNo:  notification.OutTradeNo,
		TradeStatus: string(notification.TradeStatus),
	}, nil
}

func (adapter *SDKAlipayAdapter) TransferToAlipayAccount(
	ctx context.Context,
	input ports.TransferInput,
) (ports.TransferOutput, error) {
	param := alipaysdk.FundTransUniTransfer{
		OutBizNo:    input.OutBizNo,
		TransAmount: centsToYuan(input.AmountInCents),
		ProductCode: transferProductCode,
		BizScene:    transferBizScene,
		PayeeInfo: &alipaysdk.PayeeInfo{
			Identity:     input.PayeeAccount,
			IdentityType: identityTypeLogonID,
			Name:         input.PayeeRealName,
		},
		Remark: input.Remark,
	}

	result, err := adapter.client.FundTransUniTransfer(ctx, param)
	if err != nil {
		return ports.TransferOutput{}, fmt.Errorf("payment: alipay fund trans uni transfer: %w", err)
	}
	if !result.IsSuccess() {
		return ports.TransferOutput{}, fmt.Errorf("payment: alipay transfer failed: %s %s", result.Code, result.SubMsg)
	}

	return ports.TransferOutput{AlipayOrderID: result.OrderId}, nil
}

// centsToYuan formats an integer amount in minor units (cents) as a yuan
// string with exactly two decimal places, as required by the Alipay API.
func centsToYuan(cents uint64) string {
	return fmt.Sprintf("%.2f", float64(cents)/centsPerYuan)
}

const centsPerYuan = 100
