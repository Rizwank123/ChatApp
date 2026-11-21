//go:build wireinject

package dependency

import (
	"context"
	"fmt"
	"strings"
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
	"github.com/Rizwank123/chatApp/internal/storage"
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
	API            *api.ChatApi
	Hub            *websockets.Hub
	Broker         domain.MessageBroker
	Redis          *cache.RedisClient
	MediaProcessor *storage.MediaProcessor
}

func ProvideRedisClient(cfg config.ChatApiConfig) (*cache.RedisClient, error) {
	return cache.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
}

func ProvideStorageProvider(cfg config.ChatApiConfig) (domain.StorageProvider, error) {
	if cfg.AppEnv == "production" {
		if cfg.MinIOEndpoint == "" || strings.Contains(cfg.MinIOEndpoint, "://") || cfg.MinIOBucket == "" || !cfg.MinIOUseSSL {
			return nil, fmt.Errorf("production object storage requires an endpoint, bucket, and TLS")
		}
		if cfg.MinIOAccessKey == "" || cfg.MinIOSecretKey == "" || cfg.MinIOAccessKey == "minioadmin" || cfg.MinIOSecretKey == "minioadmin" || cfg.MinIOAccessKey == "localminio" || cfg.MinIOSecretKey == "local-minio-password-change-me" || cfg.MinIOAccessKey == "localdev" || cfg.MinIOSecretKey == "local-dev-secret-change-me" {
			return nil, fmt.Errorf("production object storage credentials must be explicitly configured")
		}
	}
	return storage.NewMinIOStorageProvider(storage.MinIOConfig{
		Provider: cfg.StorageProvider, Endpoint: cfg.MinIOEndpoint, AccessKey: cfg.MinIOAccessKey,
		SecretKey: cfg.MinIOSecretKey, Bucket: cfg.MinIOBucket, UseSSL: cfg.MinIOUseSSL,
		Region: cfg.MinIORegion, PublicURL: cfg.MinIOPublicURL, CreateBucket: cfg.AppEnv != "production",
	})
}

func ProvideMessageBroker(client *cache.RedisClient) domain.MessageBroker {
	return broker.NewRedisMessageBroker(client.Client())
}

func ProvideTicketManager(client *cache.RedisClient) *websockets.TicketManager {
	return websockets.NewTicketManager(client.Client())
}

func ProvideHybridFeedEngine(
	db *pgxpool.Pool,
	redisClient *cache.RedisClient,
	socialRepo domain.SocialGraphRepository,
	postRepo domain.PostRepository,
) *feed.HybridFeedEngine {
	return feed.NewHybridFeedEngine(db, redisClient, socialRepo, postRepo, nil)
}

func ProvideChatController(
	cs domain.ChatService,
	hub *websockets.Hub,
	tickets *websockets.TicketManager,
	cfg config.ChatApiConfig,
) controller.ChatController {
	origins := make([]string, 0)
	for _, origin := range strings.Split(cfg.AllowedOrigins, ",") {
		if origin = strings.TrimSpace(origin); origin != "" {
			origins = append(origins, origin)
		}
	}
	return controller.NewChatController(cs, hub, tickets, origins)
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
	mc controller.MediaController,
	search controller.SearchController,
) *api.ChatApi {
	chatApi := api.NewChatApi(cfg, cc, pr, psc, uc)
	chatApi.SetSocialController(sc)
	chatApi.SetMediaController(mc)
	chatApi.SetSearchController(search)
	return chatApi
}

func NewChatAppApi(cfg config.ChatApiConfig, db *pgxpool.Pool) (*ChatAppResult, error) {
	wire.Build(
		util.NewAppUtil,
		security.NewJwtSecurityManager,
		ProvideRedisClient,
		ProvideStorageProvider,
		ProvideMessageBroker,
		ProvideTicketManager,

		repository.NewTransactioner,

		repository.NewUserRepository,
		repository.NewPersonnelRepository,
		repository.NewChatRepository,
		repository.NewPostRepository,
		repository.NewSocialGraphRepository,
		repository.NewNotificationRepository,
		repository.NewMediaRepository,
		repository.NewSearchRepository,

		service.NewUserService,
		service.NewPersonnelService,
		ProvideUserExistenceChecker,
		service.NewChatService,
		service.NewMediaService,
		storage.NewMediaProcessor,
		service.NewEnhancedPostService,
		service.NewSearchService,
		ProvideHybridFeedEngine,
		service.NewSocialServiceWithPersonnel,

		websockets.NewHub,
		ProvideChatController,
		controller.NewUserController,
		controller.NewPersonnelController,
		controller.NewPostController,
		controller.NewSocialController,
		controller.NewMediaController,
		controller.NewSearchController,

		ProvideChatApi,
		wire.Bind(new(domain.FeedFanoutEngine), new(*feed.HybridFeedEngine)),
		wire.Bind(new(domain.MediaProcessorEnqueuer), new(*storage.MediaProcessor)),
		wire.Bind(new(service.UserEventBroadcaster), new(*websockets.Hub)),
		wire.Struct(new(ChatAppResult), "*"),
	)
	return &ChatAppResult{}, nil
}
