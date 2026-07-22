package command

import (
	"context"

	"auction/internal/modules/notification/domain/model"
	"auction/internal/modules/notification/ports"
	"auction/internal/shared/modules/logger"
)

type CreateWatchCommandInput struct {
	UserID uint64
	SpuID  uint64
}

type CreateWatchCommandOutput struct {
	Watchlist model.Watchlist
}

type CreateWatchCommand struct {
	watchlists ports.WatchlistRepository
	logger     logger.Logger
}

func NewCreateWatchCommand(
	watchlists ports.WatchlistRepository,
	logger logger.Logger,
) *CreateWatchCommand {
	return &CreateWatchCommand{
		watchlists: watchlists,
		logger:     logger,
	}
}

func (command *CreateWatchCommand) Execute(
	ctx context.Context,
	input CreateWatchCommandInput,
) (CreateWatchCommandOutput, error) {
	watchlist, buildErr := model.NewWatchlist(input.UserID, input.SpuID)
	if buildErr != nil {
		return CreateWatchCommandOutput{}, buildErr
	}

	saved, saveErr := command.watchlists.Save(ctx, watchlist)
	if saveErr != nil {
		return CreateWatchCommandOutput{}, saveErr
	}

	command.logger.Info().
		Uint64("user_id", input.UserID).
		Uint64("spu_id", input.SpuID).
		Msg("watchlist entry created")

	return CreateWatchCommandOutput{Watchlist: saved}, nil
}
