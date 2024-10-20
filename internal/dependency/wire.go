//go:build wireinject

package dependency

import (
	"github.com/google/wire"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chatApp/internal/database"
	"github.com/chatApp/internal/http/api"
	"github.com/chatApp/internal/http/controller"
	"github.com/chatApp/internal/pkg/config"
	"github.com/chatApp/internal/pkg/security"
	"github.com/chatApp/internal/pkg/util"
	"github.com/chatApp/internal/repository"
	"github.com/chatApp/internal/service"
)

func NewConfig(opt config.Options) (cfg config.ChatApiConfig, err error) {
	wire.Build(
		config.NewConfig,
	)
	return config.ChatApiConfig{}, err
}

func NewDatabaseConfig(cfg config.ChatApiConfig) (*pgxpool.Pool, error) {
	wire.Build(
		database.NewDB,
	)

	return &pgxpool.Pool{}, nil
}
func NewChatAppApi(cfg config.ChatApiConfig, db *pgxpool.Pool) (*api.ChatApi, error) {
	wire.Build(
		util.NewAppUtil,
		security.NewJwtSecurityManager,

		repository.NewTransactioner,

		repository.NewUserRepository,
		repository.NewPersonnelRepository,

		service.NewUserService,
		service.NewPersonnelService,
		service.NewProcpectaService,

		controller.NewUserController,
		controller.NewPersonnelController,
		controller.NewProspectaController,

		api.NewChatApi,
	)
	return &api.ChatApi{}, nil
}
