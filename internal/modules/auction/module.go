package auction

import (
	"context"

	"go.uber.org/fx"

	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/application/command"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/application/query"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/infra/event/dispatcher"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/infra/http/chi/handler"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/infra/http/chi/router"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/infra/mapper"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/infra/repository"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/infra/uow"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/infra/websocket"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/ports"
	"github.com/cristiano-pacheco/go-online-auction/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/go-online-auction/pkg/httpserver"
)

var Module = fx.Module(
	"auction",

	fx.Provide(mapper.NewAuctionMapper),
	fx.Provide(mapper.NewBidMapper),

	fx.Provide(
		fx.Annotate(
			repository.NewPostgresAuctionRepository,
			fx.As(new(ports.AuctionRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			repository.NewPostgresBidRepository,
			fx.As(new(ports.BidRepository)),
		),
	),

	fx.Provide(
		fx.Annotate(
			uow.NewAuctionUnitOfWorkFactory,
			fx.As(new(ports.AuctionUnitOfWorkFactory)),
		),
	),

	fx.Provide(
		fx.Annotate(
			dispatcher.NewRedisAuctionStartedEventDispatcher,
			fx.As(new(ports.AuctionStartedEventDispatcher)),
		),
	),
	fx.Provide(
		fx.Annotate(
			dispatcher.NewRedisBidPlacedEventDispatcher,
			fx.As(new(ports.BidPlacedEventDispatcher)),
		),
	),
	fx.Provide(
		fx.Annotate(
			dispatcher.NewRedisAuctionEndedEventDispatcher,
			fx.As(new(ports.AuctionEndedEventDispatcher)),
		),
	),

	fx.Provide(websocket.NewAuctionSubscriberRegistry),
	fx.Provide(websocket.NewHub),

	fx.Provide(command.NewCreateAuctionCommand),
	fx.Provide(command.NewStartAuctionCommand),
	fx.Provide(command.NewPlaceBidCommand),
	fx.Provide(command.NewCloseAuctionCommand),
	fx.Provide(command.NewCancelAuctionCommand),

	fx.Provide(query.NewGetAuctionByIDQuery),
	fx.Provide(query.NewListAuctionsQuery),

	fx.Provide(handler.NewAuctionHandler),
	fx.Provide(handler.NewWebsocketHandler),
)

func RegisterAuctionRoutes(
	server *httpserver.Server,
	auctionHandler *handler.AuctionHandler,
) {
	router.RegisterAuctionRoutes(server, auctionHandler)
}

func RegisterWebsocketRoutes(
	lc fx.Lifecycle,
	hub *websocket.Hub,
	server *httpserver.Server,
	websocketHandler *handler.WebsocketHandler,
	logger logger.Logger,
) {
	router.RegisterWebsocketRoutes(server, websocketHandler)

	// Create a context for the hub that lives throughout the application lifecycle
	hubCtx, hubCancel := context.WithCancel(context.Background())

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			logger.Info().Msg("starting websocket hub")
			go hub.Run(hubCtx)
			return nil
		},
		OnStop: func(_ context.Context) error {
			logger.Info().Msg("stopping websocket hub")
			hubCancel() // Cancel the hub context
			return hub.Shutdown()
		},
	})
}
