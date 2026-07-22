package command_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"auction/internal/modules/payment/application/command"
	"auction/internal/modules/payment/domain/enum"
	"auction/internal/modules/payment/domain/errs"
	"auction/internal/modules/payment/domain/model"
	"auction/internal/modules/payment/ports"
	"auction/internal/shared/modules/config"
	"auction/tests/mocks"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const (
	depositUserID   = uint64(100)
	depositAmount   = uint64(9900)
	depositCurrency = "CNY"
	depositOutTrade = "out-trade-001"
	depositQRCode   = "https://qr.alipay.com/abc123"
	depositTradeNo  = "alipay-trade-001"
)

type CreateDepositCommandTestSuite struct {
	suite.Suite
	sut            *command.CreateDepositCommand
	alipayPortMock *mocks.MockAlipayPort
	paymentsMock   *mocks.MockPaymentRepository
	loggerMock     *mocks.MockLogger
	alipayCfg      config.Alipay
}

func (s *CreateDepositCommandTestSuite) SetupTest() {
	s.alipayPortMock = mocks.NewMockAlipayPort(s.T())
	s.paymentsMock = mocks.NewMockPaymentRepository(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())
	s.alipayCfg = config.Alipay{
		Provider:             "mock",
		NotifyBaseURL:        "https://pay.example.com",
		PlatformAccountOwner: "platform",
	}
	nopLogger := zerolog.Nop()
	s.loggerMock.On("Info").Return(nopLogger.Info()).Maybe()
	s.loggerMock.On("Error").Return(nopLogger.Error()).Maybe()

	s.sut = command.NewCreateDepositCommand(s.alipayPortMock, s.paymentsMock, s.alipayCfg, s.loggerMock)
}

func TestCreateDepositCommandSuite(t *testing.T) {
	suite.Run(t, new(CreateDepositCommandTestSuite))
}

func (s *CreateDepositCommandTestSuite) TestExecute_Success_CreatesOrderAndReturnsQR() {
	// Arrange
	ctx := context.Background()
	input := command.CreateDepositCommandInput{
		UserID:         depositUserID,
		AmountInCents:  depositAmount,
		Currency:       depositCurrency,
		IdempotencyKey: depositOutTrade,
	}

	persisted, buildErr := model.RestorePaymentModel(
		42, depositUserID, depositAmount, depositCurrency, enum.PaymentStatusCreated,
		depositOutTrade, depositQRCode, "", 1, time.Now(), time.Now(),
	)
	s.Require().NoError(buildErr)

	s.alipayPortMock.EXPECT().CreateFaceToFacePayment(mock.Anything, mock.Anything).
		Return(ports.FaceToFaceOutput{QRCodeURL: depositQRCode, OutTradeNo: depositOutTrade}, nil)
	s.paymentsMock.EXPECT().Save(mock.Anything, mock.Anything).Return(persisted, nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(uint64(42), output.PaymentID)
	s.Equal(depositOutTrade, output.OutTradeNo)
	s.Equal(depositQRCode, output.QRCodeURL)
	s.Equal(string(enum.PaymentStatusCreated), output.Status)
}

func (s *CreateDepositCommandTestSuite) TestExecute_AlipayFailure_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.CreateDepositCommandInput{
		UserID:         depositUserID,
		AmountInCents:  depositAmount,
		Currency:       depositCurrency,
		IdempotencyKey: depositOutTrade,
	}
	alipayErr := errors.New("alipay gateway unreachable")

	s.alipayPortMock.EXPECT().CreateFaceToFacePayment(mock.Anything, mock.Anything).
		Return(ports.FaceToFaceOutput{}, alipayErr)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, alipayErr)
	s.Equal(uint64(0), output.PaymentID)
}

func (s *CreateDepositCommandTestSuite) TestExecute_ZeroAmount_ReturnsValidationError() {
	// Arrange
	ctx := context.Background()
	input := command.CreateDepositCommandInput{
		UserID:         depositUserID,
		AmountInCents:  0,
		Currency:       depositCurrency,
		IdempotencyKey: depositOutTrade,
	}

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrPaymentAmountRequired)
	s.Equal(uint64(0), output.PaymentID)
}

func (s *CreateDepositCommandTestSuite) TestExecute_GeneratesOutTradeNo_WhenNotProvided() {
	// Arrange
	ctx := context.Background()
	input := command.CreateDepositCommandInput{
		UserID:        depositUserID,
		AmountInCents: depositAmount,
		Currency:      depositCurrency,
	}

	persisted, buildErr := model.RestorePaymentModel(
		42, depositUserID, depositAmount, depositCurrency, enum.PaymentStatusCreated,
		"generated-out-trade", depositQRCode, "", 1, time.Now(), time.Now(),
	)
	s.Require().NoError(buildErr)

	s.alipayPortMock.EXPECT().CreateFaceToFacePayment(mock.Anything, mock.Anything).
		Return(ports.FaceToFaceOutput{QRCodeURL: depositQRCode, OutTradeNo: "generated-out-trade"}, nil)
	s.paymentsMock.EXPECT().Save(mock.Anything, mock.Anything).Return(persisted, nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(uint64(42), output.PaymentID)
}
