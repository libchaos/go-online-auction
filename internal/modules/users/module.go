package users

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"

	"auction/internal/modules/users/application/command"
	"auction/internal/modules/users/application/query"
	"auction/internal/modules/users/infra/hasher"
	"auction/internal/modules/users/infra/http/chi/handler"
	"auction/internal/modules/users/infra/http/chi/router"
	"auction/internal/modules/users/infra/mapper"
	"auction/internal/modules/users/infra/repository"
	"auction/internal/modules/users/infra/sqlcgen"
	"auction/internal/modules/users/infra/token"
	"auction/internal/modules/users/ports"
	"auction/internal/shared/modules/authn"
	"auction/internal/shared/modules/authz"
	"auction/pkg/httpserver"
)

var Module = fx.Module(
	"users",

	fx.Provide(func(pool *pgxpool.Pool) sqlcgen.DBTX { return pool }),

	fx.Provide(mapper.NewUserMapper),
	fx.Provide(mapper.NewRefreshTokenMapper),

	fx.Provide(
		fx.Annotate(
			repository.NewPostgresUserRepository,
			fx.As(new(ports.UserRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			repository.NewPostgresRefreshTokenRepository,
			fx.As(new(ports.RefreshTokenRepository)),
		),
	),

	fx.Provide(
		fx.Annotate(
			hasher.NewBcryptPasswordHasher,
			fx.As(new(ports.PasswordHasher)),
		),
	),
	fx.Provide(
		fx.Annotate(
			token.NewJWTTokenService,
			fx.As(new(ports.TokenService)),
			fx.As(new(authn.TokenVerifier)),
		),
	),

	fx.Provide(command.NewRegisterUserCommand),
	fx.Provide(command.NewLoginCommand),
	fx.Provide(command.NewRefreshTokenCommand),
	fx.Provide(command.NewLogoutCommand),
	fx.Provide(command.NewUpdateProfileCommand),
	fx.Provide(command.NewChangePasswordCommand),
	fx.Provide(command.NewUpdateUserRoleCommand),

	fx.Provide(query.NewGetUserByIDQuery),
	fx.Provide(query.NewListUsersQuery),

	fx.Provide(command.NewAddPolicyCommand),
	fx.Provide(command.NewRemovePolicyCommand),
	fx.Provide(query.NewListPoliciesQuery),
	fx.Provide(command.NewAssignRoleCommand),
	fx.Provide(command.NewRevokeRoleCommand),
	fx.Provide(command.NewRevokeAllRolesCommand),
	fx.Provide(query.NewListRoleAssignmentsQuery),
	fx.Provide(handler.NewRBACHandler),

	fx.Provide(handler.NewUserHandler),
)

func RegisterUserRoutes(
	server *httpserver.Server,
	userHandler *handler.UserHandler,
	middleware *authn.Middleware,
	authzMiddleware *authz.Middleware,
) {
	router.RegisterUserRoutes(server, userHandler, middleware, authzMiddleware)
}

func RegisterRBACRoutes(
	server *httpserver.Server,
	rbacHandler *handler.RBACHandler,
	middleware *authn.Middleware,
) {
	router.RegisterRBACRoutes(server, rbacHandler, middleware)
}
