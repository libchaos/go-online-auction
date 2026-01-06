package database

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"

	"github.com/cristiano-pacheco/go-online-auction/internal/shared/modules/uow"
)

var Module = fx.Module(
	"shared/database",
	fx.Provide(New),
	fx.Provide(
		fx.Annotate(
			func(pool *pgxpool.Pool) uow.DBExecutor { return pool },
			fx.As(new(uow.DBExecutor)),
		),
	),
)
