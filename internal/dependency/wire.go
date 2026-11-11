//go:build wireinject

package dependency

import (
	"context"
	"time"

	"github.com/google/wire"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Rizwank123/chatApp/internal/broker"
	"github.com/Rizwank123/chatApp/internal/cache"
	"github.com/Rizwank123/chatApp/internal/database"
	"github.com/Rizwank123/chatApp/internal/domain"
	"github.com/Rizwank123/chatApp/internal/feed"
	"github.com/Rizwank123/chatApp/internal/http/api"
	"github.com/Rizwank123/chatApp/internal/http/controller"
	"github.com/Rizwank123/chatApp/internal/pkg/config"
	"github.com/Rizwank123/chatApp/internal/pkg/security"
	"github.com/Rizwank123/chatApp/internal/pkg/util"
	"github.com/Rizwank123/chatApp/internal/repository"
	"github.com/Rizwank123/chatApp/internal/service"
	"github.com/Rizwank123/chatApp/internal/websockets"
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

type ChatAppResult struct {
	API    *api.ChatApi
	Hub    *websockets.Hub
	Broker domain.MessageBroker
}

func ProvideHybridFeedEngine(
	db *pgxpool.Pool,
	socialRepo domain.SocialGraphRepository,
	postRepo domain.PostRepository,
) *feed.HybridFeedEngine {
	return feed.NewHybridFeedEngine(db, nil, socialRepo, postRepo, nil)
}

func ProvideUserExistenceChecker(ur domain.UserRepository) *cache.UserExistenceChecker {
	checker := cache.NewUserExistenceChecker(
		cache.NewMemoryBloomFilter(100000, 0.01),
		ur.ExistsByUserName,
	)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	usernames, err := ur.GetAllUsernames(ctx)
	if err == nil {
		checker.PopulateUsernames(usernames)
	}
	return checker
}

func ProvideChatApi(
	cfg config.ChatApiConfig,
	cc controller.ChatController,
	pr controller.PersonnelController,
	psc controller.PostController,
	uc controller.UserController,
	sc controller.SocialController,
) *api.ChatApi {
	chatApi := api.NewChatApi(cfg, cc, pr, psc, uc)
	chatApi.SetSocialController(sc)
	return chatApi
}

func NewChatAppApi(cfg config.ChatApiConfig, db *pgxpool.Pool) (*ChatAppResult, error) {
	wire.Build(
		util.NewAppUtil,
		security.NewJwtSecurityManager,

		repository.NewTransactioner,

		repository.NewUserRepository,
		repository.NewPersonnelRepository,
		repository.NewChatRepository,
		repository.NewPostRepository,
		repository.NewSocialGraphRepository,
		repository.NewNotificationRepository,

		service.NewUserService,
		service.NewPersonnelService,
		ProvideUserExistenceChecker,
		broker.NewInMemoryBroker,
		service.NewChatService,
		service.NewPostService,
		ProvideHybridFeedEngine,
		wire.Value((*websockets.DistributedHub)(nil)),
		service.NewSocialService,

		websockets.NewHub,
		controller.NewUserController,
		controller.NewPersonnelController,
		controller.NewChatController,
		controller.NewPostController,
		controller.NewSocialController,

		ProvideChatApi,
		wire.Struct(new(ChatAppResult), "*"),
	)
	return &ChatAppResult{}, nil
}
