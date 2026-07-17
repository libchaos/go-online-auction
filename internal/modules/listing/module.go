package listing

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"

	auctionports "auction/internal/modules/auction/ports"
	"auction/internal/modules/listing/application/command"
	"auction/internal/modules/listing/application/query"
	"auction/internal/modules/listing/infra/gateway"
	"auction/internal/modules/listing/infra/http/chi/handler"
	"auction/internal/modules/listing/infra/http/chi/router"
	"auction/internal/modules/listing/infra/mapper"
	"auction/internal/modules/listing/infra/repository"
	"auction/internal/modules/listing/infra/sqlcgen"
	"auction/internal/modules/listing/infra/uow"
	"auction/internal/modules/listing/ports"
	"auction/internal/shared/modules/authn"
	"auction/pkg/httpserver"
)

var Module = fx.Module(
	"listing",

	fx.Provide(func(pool *pgxpool.Pool) sqlcgen.DBTX { return pool }),

	fx.Provide(mapper.NewCategoryMapper),
	fx.Provide(mapper.NewSpuMapper),
	fx.Provide(mapper.NewSkuMapper),

	fx.Provide(
		fx.Annotate(
			repository.NewPostgresCategoryRepository,
			fx.As(new(ports.CategoryRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			repository.NewPostgresSpuRepository,
			fx.As(new(ports.SpuRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			repository.NewPostgresSkuRepository,
			fx.As(new(ports.SkuRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			repository.NewPostgresListingOutboxRepository,
			fx.As(new(ports.ListingOutboxRepository)),
		),
	),

	fx.Provide(
		fx.Annotate(
			uow.NewListingUnitOfWorkFactory,
			fx.As(new(ports.ListingUnitOfWorkFactory)),
		),
	),

	// Anti-corruption adapter: implements the auction module's ListingValidator
	// port so CreateAuctionCommand can verify a listing without coupling the
	// auction module to the listing catalog.
	fx.Provide(
		fx.Annotate(
			gateway.NewAuctionListingValidator,
			fx.As(new(auctionports.ListingValidator)),
		),
	),

	fx.Provide(command.NewCreateCategoryCommand),
	fx.Provide(command.NewUpdateCategoryCommand),
	fx.Provide(command.NewDeleteCategoryCommand),
	fx.Provide(command.NewCreateSpuCommand),
	fx.Provide(command.NewUpdateSpuCommand),
	fx.Provide(command.NewPublishSpuCommand),
	fx.Provide(command.NewOffShelfSpuCommand),
	fx.Provide(command.NewCreateSkuCommand),
	fx.Provide(command.NewUpdateSkuCommand),
	fx.Provide(command.NewPublishSkuCommand),
	fx.Provide(command.NewOffShelfSkuCommand),

	fx.Provide(query.NewListCategoriesQuery),
	fx.Provide(query.NewGetCategoryByIDQuery),
	fx.Provide(query.NewListSpusQuery),
	fx.Provide(query.NewGetSpuByIDQuery),
	fx.Provide(query.NewGetSkuByIDQuery),

	fx.Provide(handler.NewCategoryHandler),
	fx.Provide(handler.NewSpuHandler),
	fx.Provide(handler.NewSkuHandler),
)

func RegisterListingRoutes(
	server *httpserver.Server,
	categoryHandler *handler.CategoryHandler,
	spuHandler *handler.SpuHandler,
	skuHandler *handler.SkuHandler,
	middleware *authn.Middleware,
) {
	router.RegisterListingRoutes(server, categoryHandler, spuHandler, skuHandler, middleware)
}
