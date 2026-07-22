package command_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"auction/internal/modules/payment/application/command"
	"auction/internal/modules/payment/domain/enum"
	"auction/internal/modules/payment/domain/model"
	"auction/internal/modules/payment/ports"
	"auction/tests/mocks"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type AlipayNotifyCommandTestSuite struct {
	suite.Suite
	sut            *command.AlipayNotifyCommand
	alipayPortMock *mocks.MockAlipayPort
	uowFactoryMock *mocks.MockPaymentUnitOfWorkFactory
	uowMock        *mocks.MockPaymentUnitOfWork
	paymentsMock   *mocks.MockPaymentRepository
	outboxMock     *mocks.MockPaymentOutboxRepository
	loggerMock     *mocks.MockLogger
}

func (s *AlipayNotifyCommandTestSuite) SetupTest() {
	s.alipayPortMock = mocks.NewMockAlipayPort(s.T())
	s.uowFactoryMock = mocks.NewMockPaymentUnitOfWorkFactory(s.T())
	s.uowMock = mocks.NewMockPaymentUnitOfWork(s.T())
	s.paymentsMock = mocks.NewMockPaymentRepository(s.T())
	s.outboxMock = mocks.NewMockPaymentOutboxRepository(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())
	nopLogger := zerolog.Nop()
	s.loggerMock.On("Info").Return(nopLogger.Info()).Maybe()
	s.loggerMock.On("Error").Return(nopLogger.Error()).Maybe()

	s.sut = command.NewAlipayNotifyCommand(s.alipayPortMock, s.uowFactoryMock, s.loggerMock)
}

func TestAlipayNotifyCommandSuite(t *testing.T) {
	suite.Run(t, new(AlipayNotifyCommandTestSuite))
}

func (s *AlipayNotifyCommandTestSuite) TestExecute_TradeSuccess_MarksPaidAndWritesOutboxEvent() {
	// Arrange
	ctx := context.Background()
	params := map[string]string{
		"out_trade_no": depositOutTrade,
		"trade_no":     depositTradeNo,
		"trade_status": "TRADE_SUCCESS",
	}

	created, createdErr := model.RestorePaymentModel(
		42, depositUserID, depositAmount, depositCurrency, enum.PaymentStatusCreated,
		depositOutTrade, depositQRCode, "", 1, time.Now(), time.Now(),
	)
	s.Require().NoError(createdErr)
	successPayment, successErr := model.RestorePaymentModel(
		42, depositUserID, depositAmount, depositCurrency, enum.PaymentStatusSuccess,
		depositOutTrade, depositQRCode, depositTradeNo, 2, time.Now(), time.Now(),
	)
	s.Require().NoError(successErr)

	s.alipayPortMock.EXPECT().VerifyNotify(mock.Anything, mock.Anything).
		Return(ports.NotifyResult{TradeNo: depositTradeNo, OutTradeNo: depositOutTrade, TradeStatus: "TRADE_SUCCESS"}, nil)
	s.uowFactoryMock.EXPECT().Begin(mock.Anything).Return(s.uowMock, nil)
	s.uowMock.EXPECT().PaymentRepository().Return(s.paymentsMock)
	s.uowMock.EXPECT().OutboxRepository().Return(s.outboxMock)
	s.uowMock.EXPECT().Rollback(mock.Anything).Return(nil)
	s.paymentsMock.EXPECT().FindByOutTradeNo(mock.Anything, mock.Anything).Return(created, nil)
	s.paymentsMock.EXPECT().Update(mock.Anything, mock.Anything).Return(successPayment, nil)
	s.outboxMock.EXPECT().Save(mock.Anything, mock.Anything).Return(nil)
	s.uowMock.EXPECT().Complete(mock.Anything).Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, command.AlipayNotifyCommandInput{Params: params})

	// Assert
	s.Require().NoError(err)
	s.Equal(uint64(42), output.PaymentID)
	s.Equal(string(enum.PaymentStatusSuccess), output.Status)
}

func (s *AlipayNotifyCommandTestSuite) TestExecute_AlreadyTerminal_ReturnsWithoutUpdate() {
	// Arrange
	ctx := context.Background()
	params := map[string]string{
		"out_trade_no": depositOutTrade,
		"trade_no":     depositTradeNo,
		"trade_status": "TRADE_SUCCESS",
	}

	alreadySuccess, restoreErr := model.RestorePaymentModel(
		42, depositUserID, depositAmount, depositCurrency, enum.PaymentStatusSuccess,
		depositOutTrade, depositQRCode, depositTradeNo, 2, time.Now(), time.Now(),
	)
	s.Require().NoError(restoreErr)

	s.alipayPortMock.EXPECT().VerifyNotify(mock.Anything, mock.Anything).
		Return(ports.NotifyResult{TradeNo: depositTradeNo, OutTradeNo: depositOutTrade, TradeStatus: "TRADE_SUCCESS"}, nil)
	s.uowFactoryMock.EXPECT().Begin(mock.Anything).Return(s.uowMock, nil)
	s.uowMock.EXPECT().PaymentRepository().Return(s.paymentsMock)
	s.uowMock.EXPECT().Rollback(mock.Anything).Return(nil)
	s.paymentsMock.EXPECT().FindByOutTradeNo(mock.Anything, mock.Anything).Return(alreadySuccess, nil)

	// Act
	output, err := s.sut.Execute(ctx, command.AlipayNotifyCommandInput{Params: params})

	// Assert: no Update / outbox / Complete is expected for a terminal payment.
	s.Require().NoError(err)
	s.Equal(string(enum.PaymentStatusSuccess), output.Status)
}

func (s *AlipayNotifyCommandTestSuite) TestExecute_VerifyFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	params := map[string]string{"out_trade_no": depositOutTrade}
	verifyErr := errors.New("signature mismatch")

	s.alipayPortMock.EXPECT().VerifyNotify(mock.Anything, mock.Anything).
		Return(ports.NotifyResult{}, verifyErr)

	// Act
	output, err := s.sut.Execute(ctx, command.AlipayNotifyCommandInput{Params: params})

	// Assert
	s.Require().ErrorIs(err, verifyErr)
	s.Equal(uint64(0), output.PaymentID)
}
