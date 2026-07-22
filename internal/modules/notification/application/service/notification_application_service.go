package service

import (
	"context"
	"encoding/json"
	"fmt"

	"auction/internal/modules/notification/application/command"
	"auction/internal/modules/notification/domain/enum"
	"auction/internal/modules/notification/ports"
	"auction/internal/shared/modules/logger"
)

// NotificationApplicationService turns a decoded source event into an in-app
// notification. It resolves the recipient, honours the recipient's channel
// preferences, and delegates the idempotent write to the create-notification
// command. Recipients whose in-app channel is disabled are skipped silently.
type NotificationApplicationService struct {
	createNotificationCommand *command.CreateNotificationCommand
	createEmailRequestCommand *command.CreateEmailRequestCommand
	preferences               ports.PreferenceRepository
	auctionRead               ports.AuctionReadPort
	watchlist                 ports.WatchlistRepository
	listingRead               ports.ListingReadPort
	userEmailResolver         ports.UserEmailResolver
	logger                    logger.Logger
}

func NewNotificationApplicationService(
	createNotificationCommand *command.CreateNotificationCommand,
	createEmailRequestCommand *command.CreateEmailRequestCommand,
	preferences ports.PreferenceRepository,
	auctionRead ports.AuctionReadPort,
	watchlist ports.WatchlistRepository,
	listingRead ports.ListingReadPort,
	userEmailResolver ports.UserEmailResolver,
	logger logger.Logger,
) *NotificationApplicationService {
	return &NotificationApplicationService{
		createNotificationCommand: createNotificationCommand,
		createEmailRequestCommand: createEmailRequestCommand,
		preferences:               preferences,
		auctionRead:               auctionRead,
		watchlist:                 watchlist,
		listingRead:               listingRead,
		userEmailResolver:         userEmailResolver,
		logger:                    logger,
	}
}

type RechargeSuccessInput struct {
	SourceEventID string
	UserID        uint64
	PaymentID     uint64
	AmountInCents uint64
	Currency      string
}

func (service *NotificationApplicationService) HandleRechargeSuccess(
	ctx context.Context,
	input RechargeSuccessInput,
) error {
	payload := map[string]any{
		"payment_id":      input.PaymentID,
		"amount_in_cents": input.AmountInCents,
		"currency":        input.Currency,
	}

	return service.dispatch(ctx, dispatchInput{
		SourceEventID: input.SourceEventID,
		UserID:        input.UserID,
		Type:          enum.EnumNotificationTypeRechargeSuccess,
		Title:         "Recharge successful",
		Body: fmt.Sprintf(
			"Your recharge of %s %s was successful.",
			formatMoney(input.AmountInCents),
			input.Currency,
		),
		Payload: payload,
	})
}

type WithdrawalCompletedInput struct {
	SourceEventID string
	UserID        uint64
	WithdrawalID  uint64
	AmountInCents uint64
	Currency      string
}

func (service *NotificationApplicationService) HandleWithdrawalCompleted(
	ctx context.Context,
	input WithdrawalCompletedInput,
) error {
	payload := map[string]any{
		"withdrawal_id":   input.WithdrawalID,
		"amount_in_cents": input.AmountInCents,
		"currency":        input.Currency,
	}

	return service.dispatch(ctx, dispatchInput{
		SourceEventID: input.SourceEventID,
		UserID:        input.UserID,
		Type:          enum.EnumNotificationTypeWithdrawalCompleted,
		Title:         "Withdrawal completed",
		Body: fmt.Sprintf(
			"Your withdrawal of %s %s has been paid out to your Alipay account.",
			formatMoney(input.AmountInCents),
			input.Currency,
		),
		Payload: payload,
	})
}

type WithdrawalFailedInput struct {
	SourceEventID string
	UserID        uint64
	WithdrawalID  uint64
	AmountInCents uint64
	Currency      string
	FailReason    string
}

func (service *NotificationApplicationService) HandleWithdrawalFailed(
	ctx context.Context,
	input WithdrawalFailedInput,
) error {
	payload := map[string]any{
		"withdrawal_id":   input.WithdrawalID,
		"amount_in_cents": input.AmountInCents,
		"currency":        input.Currency,
		"fail_reason":     input.FailReason,
	}

	return service.dispatch(ctx, dispatchInput{
		SourceEventID: input.SourceEventID,
		UserID:        input.UserID,
		Type:          enum.EnumNotificationTypeWithdrawalFailed,
		Title:         "Withdrawal failed",
		Body: fmt.Sprintf(
			"Your withdrawal of %s %s could not be completed and the amount was returned to your balance.",
			formatMoney(input.AmountInCents),
			input.Currency,
		),
		Payload: payload,
	})
}

type DepositEventInput struct {
	SourceEventID string
	EventType     string
	UserID        uint64
	AuctionID     uint64
	AmountInCents uint64
	Currency      string
}

func (service *NotificationApplicationService) HandleDepositEvent(
	ctx context.Context,
	input DepositEventInput,
) error {
	notificationType, title, body, ok := depositNotificationContent(input)
	if !ok {
		return nil
	}

	payload := map[string]any{
		"auction_id":      input.AuctionID,
		"amount_in_cents": input.AmountInCents,
		"currency":        input.Currency,
	}

	return service.dispatch(ctx, dispatchInput{
		SourceEventID: input.SourceEventID,
		UserID:        input.UserID,
		Type:          notificationType,
		Title:         title,
		Body:          body,
		Payload:       payload,
	})
}

type BidPlacedInput struct {
	SourceEventID string
	AuctionID     uint64
	NewBidderID   uint64
	AmountInCents uint64
}

func (service *NotificationApplicationService) HandleBidPlaced(
	ctx context.Context,
	input BidPlacedInput,
) error {
	previousBidderID, found, err := service.auctionRead.FindPreviousHighestBidderID(
		ctx,
		input.AuctionID,
		input.NewBidderID,
	)
	if err != nil {
		return err
	}

	if !found {
		return nil
	}

	payload := map[string]any{
		"auction_id":          input.AuctionID,
		"new_amount_in_cents": input.AmountInCents,
	}

	return service.dispatch(ctx, dispatchInput{
		SourceEventID: input.SourceEventID,
		UserID:        previousBidderID,
		Type:          enum.EnumNotificationTypeOutbid,
		Title:         "You have been outbid",
		Body: fmt.Sprintf(
			"A higher bid of %s was placed on auction #%d.",
			formatMoney(input.AmountInCents),
			input.AuctionID,
		),
		Payload: payload,
	})
}

type AuctionEndedInput struct {
	SourceEventID      string
	AuctionID          uint64
	WinningBidID       *uint64
	FinalAmountInCents *uint64
}

func (service *NotificationApplicationService) HandleAuctionEnded(
	ctx context.Context,
	input AuctionEndedInput,
) error {
	if input.WinningBidID == nil {
		return nil
	}

	winnerID, found, err := service.auctionRead.FindBidderIDByBidID(ctx, *input.WinningBidID)
	if err != nil {
		return err
	}

	if !found {
		return nil
	}

	payload := map[string]any{
		"auction_id": input.AuctionID,
	}
	body := fmt.Sprintf("You won auction #%d.", input.AuctionID)
	if input.FinalAmountInCents != nil {
		payload["final_amount_in_cents"] = *input.FinalAmountInCents
		body = fmt.Sprintf(
			"You won auction #%d with a final bid of %s.",
			input.AuctionID,
			formatMoney(*input.FinalAmountInCents),
		)
	}

	return service.dispatch(ctx, dispatchInput{
		SourceEventID: input.SourceEventID,
		UserID:        winnerID,
		Type:          enum.EnumNotificationTypeAuctionEnded,
		Title:         "You won the auction",
		Body:          body,
		Payload:       payload,
	})
}

type ListingEventInput struct {
	SourceEventID string
	SpuID         uint64
	EventType     string
}

func (service *NotificationApplicationService) HandleListingEvent(
	ctx context.Context,
	input ListingEventInput,
) error {
	notificationType, ok := resolveListingNotificationType(input.EventType)
	if !ok {
		return nil
	}

	spuTitle := "the item"
	if title, found, titleErr := service.listingRead.FindSpuTitleByID(ctx, input.SpuID); titleErr != nil {
		return titleErr
	} else if found {
		spuTitle = title
	}

	watcherIDs, watcherErr := service.watchlist.FindWatcherIDsBySpuID(ctx, input.SpuID)
	if watcherErr != nil {
		return watcherErr
	}

	if len(watcherIDs) == 0 {
		return nil
	}

	title, body := listingNotificationContent(notificationType, spuTitle)
	for _, watcherID := range watcherIDs {
		if dispatchErr := service.dispatch(ctx, dispatchInput{
			SourceEventID: input.SourceEventID,
			UserID:        watcherID,
			Type:          notificationType,
			Title:         title,
			Body:          body,
			Payload:       map[string]any{"spu_id": input.SpuID},
		}); dispatchErr != nil {
			service.logger.Error().Err(dispatchErr).
				Uint64("user_id", watcherID).
				Str("type", notificationType).
				Msg("failed to dispatch listing notification")
		}
	}

	return nil
}

type dispatchInput struct {
	SourceEventID string
	UserID        uint64
	Type          string
	Title         string
	Body          string
	Payload       map[string]any
}

func (service *NotificationApplicationService) dispatch(ctx context.Context, input dispatchInput) error {
	notificationType, typeErr := enum.NewNotificationTypeEnum(input.Type)
	if typeErr != nil {
		return typeErr
	}

	category := notificationType.Category()
	categoryValue := category.String()

	preference, prefErr := service.preferences.Get(ctx, input.UserID)
	if prefErr != nil {
		return prefErr
	}

	if !preference.IsChannelEnabled(categoryValue, enum.EnumNotificationChannelInApp) {
		service.logger.Info().
			Uint64("user_id", input.UserID).
			Str("type", input.Type).
			Msg("in-app notification skipped by user preference")

		return nil
	}

	if preference.IsChannelEnabled(categoryValue, enum.EnumNotificationChannelEmail) {
		service.dispatchEmail(ctx, input)
	}

	payloadBytes, marshalErr := json.Marshal(input.Payload)
	if marshalErr != nil {
		return marshalErr
	}

	idempotencyKey := fmt.Sprintf(
		"%s:%d:%s",
		input.SourceEventID,
		input.UserID,
		enum.EnumNotificationChannelInApp,
	)

	_, execErr := service.createNotificationCommand.Execute(ctx, command.CreateNotificationCommandInput{
		UserID:         input.UserID,
		Type:           input.Type,
		Title:          input.Title,
		Body:           input.Body,
		Payload:        payloadBytes,
		Channels:       []string{enum.EnumNotificationChannelInApp},
		IdempotencyKey: idempotencyKey,
	})

	return execErr
}

// dispatchEmail enqueues an email notification for the same source event, using
// the recipient's preference-resolved email address. A missing email or a send
// failure must not block the in-app notification already written above, so
// errors are logged and swallowed here.
func (service *NotificationApplicationService) dispatchEmail(ctx context.Context, input dispatchInput) {
	email, found, resolveErr := service.userEmailResolver.ResolveEmail(ctx, input.UserID)
	if resolveErr != nil {
		service.logger.Error().Err(resolveErr).
			Uint64("user_id", input.UserID).
			Msg("failed to resolve user email; skipping email notification")

		return
	}

	if !found || email == "" {
		return
	}

	requestErr := service.createEmailRequestCommand.Execute(ctx, command.CreateEmailRequestCommandInput{
		SourceEventID: input.SourceEventID,
		UserID:        input.UserID,
		To:            email,
		Subject:       input.Title,
		Title:         input.Title,
		Body:          input.Body,
	})
	if requestErr != nil {
		service.logger.Error().Err(requestErr).
			Uint64("user_id", input.UserID).
			Str("to", email).
			Msg("failed to enqueue email notification")
	}
}

func depositNotificationContent(input DepositEventInput) (string, string, string, bool) {
	switch input.EventType {
	case enum.EnumNotificationTypeDepositHeld:
		return enum.EnumNotificationTypeDepositHeld,
			"Deposit held",
			fmt.Sprintf(
				"A deposit of %s %s was held for auction #%d.",
				formatMoney(input.AmountInCents),
				input.Currency,
				input.AuctionID,
			),
			true
	case enum.EnumNotificationTypeDepositReleased:
		return enum.EnumNotificationTypeDepositReleased,
			"Deposit released",
			fmt.Sprintf(
				"Your deposit of %s %s for auction #%d was released.",
				formatMoney(input.AmountInCents),
				input.Currency,
				input.AuctionID,
			),
			true
	case enum.EnumNotificationTypeDepositApplied:
		return enum.EnumNotificationTypeDepositApplied,
			"Deposit applied",
			fmt.Sprintf(
				"Your deposit of %s %s was applied to auction #%d.",
				formatMoney(input.AmountInCents),
				input.Currency,
				input.AuctionID,
			),
			true
	case enum.EnumNotificationTypeDepositForfeited:
		return enum.EnumNotificationTypeDepositForfeited,
			"Deposit forfeited",
			fmt.Sprintf(
				"Your deposit of %s %s for auction #%d was forfeited.",
				formatMoney(input.AmountInCents),
				input.Currency,
				input.AuctionID,
			),
			true
	default:
		return "", "", "", false
	}
}

const centsPerUnit = 100

func formatMoney(amountInCents uint64) string {
	return fmt.Sprintf("%d.%02d", amountInCents/centsPerUnit, amountInCents%centsPerUnit)
}

const (
	listingEventTypeSpuPublished = "listing.spu.published"
	listingEventTypeSkuPublished = "listing.sku.published"
	listingEventTypeSpuOffShelf  = "listing.spu.off_shelf"
	listingEventTypeSkuOffShelf  = "listing.sku.off_shelf"
)

func resolveListingNotificationType(eventType string) (string, bool) {
	switch eventType {
	case listingEventTypeSpuPublished, listingEventTypeSkuPublished:
		return enum.EnumNotificationTypeListingPublished, true
	case listingEventTypeSpuOffShelf, listingEventTypeSkuOffShelf:
		return enum.EnumNotificationTypeListingOffShelf, true
	default:
		return "", false
	}
}

func listingNotificationContent(notificationType, spuTitle string) (string, string) {
	title := "New item published"
	body := fmt.Sprintf("The item you watch \"%s\" is now available.", spuTitle)
	if notificationType == enum.EnumNotificationTypeListingOffShelf {
		title = "Item off the shelf"
		body = fmt.Sprintf("The item you watch \"%s\" has been taken down.", spuTitle)
	}

	return title, body
}
