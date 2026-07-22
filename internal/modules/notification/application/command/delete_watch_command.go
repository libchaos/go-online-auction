package command

import (
	"context"

	"auction/internal/modules/notification/ports"
	"auction/internal/shared/modules/logger"
)

type DeleteWatchCommandInput struct {
	UserID uint64
	SpuID  uint64
}

type DeleteWatchCommand struct {
	watchlists ports.WatchlistRepository
	logger     logger.Logger
}

func NewDeleteWatchCommand(
	watchlists ports.WatchlistRepository,
	logger logger.Logger,
) *DeleteWatchCommand {
	return &DeleteWatchCommand{
		watchlists: watchlists,
		logger:     logger,
	}
}

func (command *DeleteWatchCommand) Execute(
	ctx context.Context,
	input DeleteWatchCommandInput,
) error {
	removeErr := command.watchlists.Remove(ctx, input.UserID, input.SpuID)
	if removeErr != nil {
		return removeErr
	}

	command.logger.Info().
		Uint64("user_id", input.UserID).
		Uint64("spu_id", input.SpuID).
		Msg("watchlist entry removed")

	return nil
}
