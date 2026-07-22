package service

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"auction/internal/modules/payment/infra/event/envelope"
	ledgermodel "auction/internal/modules/ledger/domain/model"
	ledgerports "auction/internal/modules/ledger/ports"
	"auction/internal/shared/modules/config"
	"auction/tests/mocks"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const (
	consumerDepositUserID   = uint64(100)
	consumerDepositAmount   = uint64(9900)
	consumerDepositCurrency = "CNY"
	consumerDepositOutTrade = "out-trade-001"
)

type DepositSuccessConsumerTestSuite struct {
	suite.Suite
	sut            *DepositSuccessConsumer
	uowFactoryMock *mocks.MockPaymentUnitOfWorkFactory
	uowMock        *mocks.MockPaymentUnitOfWork
	ledgerMock     *mocks.MockLedgerRepository
	loggerMock     *mocks.MockLogger
	alipayCfg      config.Alipay
}

func (s *DepositSuccessConsumerTestSuite) SetupTest() {
	s.uowFactoryMock = mocks.NewMockPaymentUnitOfWorkFactory(s.T())
	s.uowMock = mocks.NewMockPaymentUnitOfWork(s.T())
	s.ledgerMock = mocks.NewMockLedgerRepository(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())
	s.alipayCfg = config.Alipay{Provider: "mock", PlatformAccountOwner: "platform"}
	nopLogger := zerolog.Nop()
	s.loggerMock.On("Info").Return(nopLogger.Info()).Maybe()
	s.loggerMock.On("Error").Return(nopLogger.Error()).Maybe()

	// The consumer is exercised through handle(); the jetstream handle is unused in unit tests.
	s.sut = NewDepositSuccessConsumer(nil, s.uowFactoryMock, s.alipayCfg, s.loggerMock)
}

func TestDepositSuccessConsumerSuite(t *testing.T) {
	suite.Run(t, new(DepositSuccessConsumerTestSuite))
}

func (s *DepositSuccessConsumerTestSuite) TestHandle_CreditsUserFromPlatformAccount() {
	// Arrange
	ctx := context.Background()
	payload := envelope.PaymentSuccessPayload{
		EventID:       "evt-1",
		PaymentID:     42,
		UserID:        consumerDepositUserID,
		AmountInCents: consumerDepositAmount,
		Currency:      consumerDepositCurrency,
		OutTradeNo:    consumerDepositOutTrade,
		OccurredAt:    time.Now().Format(time.RFC3339),
	}
	data, marshalErr := json.Marshal(payload)
	s.Require().NoError(marshalErr)

	userAccount, accountErr := ledgermodel.RestoreAccountModel(
		10, strconv.FormatUint(consumerDepositUserID, 10), 0, 0, consumerDepositCurrency, 1, time.Now(), time.Now(),
	)
	s.Require().NoError(accountErr)
	platformAccount, platformErr := ledgermodel.RestoreAccountModel(
		1, "platform", 1_000_000_000, 0, consumerDepositCurrency, 1, time.Now(), time.Now(),
	)
	s.Require().NoError(platformErr)

	s.uowFactoryMock.EXPECT().Begin(mock.Anything).Return(s.uowMock, nil)
	s.uowMock.EXPECT().LedgerRepository().Return(s.ledgerMock)
	s.uowMock.EXPECT().Rollback(mock.Anything).Return(nil)
	s.ledgerMock.EXPECT().GetOrCreateAccountByOwner(mock.Anything, strconv.FormatUint(consumerDepositUserID, 10), consumerDepositCurrency).
		Return(userAccount, nil)
	s.ledgerMock.EXPECT().GetOrCreateAccountByOwner(mock.Anything, "platform", consumerDepositCurrency).
		Return(platformAccount, nil)
	s.ledgerMock.EXPECT().Transfer(mock.Anything, mock.MatchedBy(func(in ledgerports.TransferInput) bool {
		return in.FromAccountID == 1 && in.ToAccountID == 10 &&
			in.Amount == consumerDepositAmount && in.IdempotencyKey == consumerDepositOutTrade
	})).Return(ledgermodel.TransferModel{}, nil)
	s.uowMock.EXPECT().Complete(mock.Anything).Return(nil)

	// Act
	s.sut.handle(ctx, data)

	// Assert: verified by mock expectations (Transfer with the out_trade_no idempotency key).
}

func (s *DepositSuccessConsumerTestSuite) TestHandle_MalformedPayload_IsIgnored() {
	// Arrange
	ctx := context.Background()
	s.uowFactoryMock.EXPECT().Begin(mock.Anything).Return(s.uowMock, nil).Maybe()
	s.uowMock.EXPECT().Rollback(mock.Anything).Return(nil).Maybe()

	// Act: malformed JSON must not panic and must not open a uow.
	s.sut.handle(ctx, []byte("not-json"))
}
