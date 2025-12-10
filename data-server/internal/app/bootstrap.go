package app

import (
	"fmt"

	"github.com/kanshan/ServerStatus/data-server/internal/config"
	"github.com/kanshan/ServerStatus/data-server/internal/repository"
	redisrepo "github.com/kanshan/ServerStatus/data-server/internal/repository/redis"
	sqliterepo "github.com/kanshan/ServerStatus/data-server/internal/repository/sqlite"
	"github.com/kanshan/ServerStatus/data-server/internal/service/auth"
	"github.com/kanshan/ServerStatus/data-server/internal/service/export"
	"github.com/kanshan/ServerStatus/data-server/internal/service/health"
	"github.com/kanshan/ServerStatus/data-server/internal/service/server"
	"github.com/kanshan/ServerStatus/data-server/internal/service/websocket"
	"github.com/kanshan/ServerStatus/data-server/pkg/logger"
)

// Components 聚合应用的核心依赖，便于在入口处统一构建。
type Components struct {
	Config           *config.Config
	Logger           logger.Logger
	Repository       repository.Repository
	ServerService    *server.ServerService
	ExportService    *export.ExportService
	AuthService      *auth.AuthService
	HealthService    *health.HealthService
	WebSocketService *websocket.WebSocketService
}

// Build 从配置路径构建应用依赖，集中管理加载、日志、仓库和服务装配。
func Build(configPath string) (*Components, error) {
	bootLogger := logger.GetDefaultLogger()

	cfg, err := config.LoadConfig(configPath, bootLogger)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	appLogger, err := logger.NewLogger(&logger.Config{
		Level:        cfg.Logging.Level,
		Format:       cfg.Logging.Format,
		Output:       cfg.Logging.Output,
		Filename:     cfg.Logging.Filename,
		MaxSize:      cfg.Logging.MaxSize,
		MaxBackups:   cfg.Logging.MaxBackups,
		MaxAge:       cfg.Logging.MaxAge,
		Compress:     cfg.Logging.Compress,
		ReportCaller: cfg.Logging.ReportCaller,
	})
	if err != nil {
		return nil, fmt.Errorf("init logger: %w", err)
	}

	repo, err := buildRepository(cfg, appLogger)
	if err != nil {
		return nil, err
	}

	serverRepo, ok := repo.(repository.ServerRepository)
	if !ok {
		return nil, fmt.Errorf("repository does not implement ServerRepository")
	}
	historyRepo, ok := repo.(repository.HistoryRepository)
	if !ok {
		return nil, fmt.Errorf("repository does not implement HistoryRepository")
	}
	cacheRepo, ok := repo.(repository.CacheRepository)
	if !ok {
		return nil, fmt.Errorf("repository does not implement CacheRepository")
	}
	accessKeyRepo, ok := repo.(repository.AccessKeyRepository)
	if !ok {
		return nil, fmt.Errorf("repository does not implement AccessKeyRepository")
	}

	serverSvc := server.NewServerService(serverRepo, historyRepo, cacheRepo, appLogger)

	exportSvc := buildExportService(repo, serverRepo, historyRepo, cacheRepo, appLogger)

	authSvc := auth.NewAuthService(accessKeyRepo, cacheRepo, appLogger, nil)

	healthSvc := health.NewHealthService(serverRepo, cacheRepo, accessKeyRepo, appLogger, health.DefaultConfig())

	wsFactory := websocket.NewFactory(serverRepo, cacheRepo, accessKeyRepo, appLogger)
	wsSvc := wsFactory.CreateDefaultService()

	return &Components{
		Config:           cfg,
		Logger:           appLogger,
		Repository:       repo,
		ServerService:    serverSvc,
		ExportService:    exportSvc,
		AuthService:      authSvc,
		HealthService:    healthSvc,
		WebSocketService: wsSvc,
	}, nil
}

func buildRepository(cfg *config.Config, log logger.Logger) (repository.Repository, error) {
	switch cfg.Database.Type {
	case "sqlite", "sqlite3", "":
		factory, err := sqliterepo.NewSQLiteRepositoryFactory(&cfg.Database, log)
		if err != nil {
			return nil, fmt.Errorf("create sqlite repository: %w", err)
		}
		repo, err := factory.NewRepository(nil)
		if err != nil {
			return nil, fmt.Errorf("init sqlite repository: %w", err)
		}
		return repo, nil
	case "redis":
		factory, err := redisrepo.NewRedisRepositoryFactory(&cfg.Cache, log)
		if err != nil {
			return nil, fmt.Errorf("create redis repository: %w", err)
		}
		repo, err := factory.NewRepository(nil)
		if err != nil {
			return nil, fmt.Errorf("init redis repository: %w", err)
		}
		return nil, fmt.Errorf("redis repository provides cache/access-key only; configure sqlite for persistence")
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Database.Type)
	}
}

func buildExportService(
	repo repository.Repository,
	serverRepo repository.ServerRepository,
	historyRepo repository.HistoryRepository,
	cacheRepo repository.CacheRepository,
	log logger.Logger,
) *export.ExportService {
	if provider, ok := repo.(interface {
		ServerRepository() repository.ServerRepository
		HistoryRepository() repository.HistoryRepository
		CacheRepository() repository.CacheRepository
	}); ok {
		factory := export.NewExportServiceFactory(provider, log)
		return factory.CreateExportService()
	}

	// 回退使用底层仓库接口
	return export.NewExportService(serverRepo, historyRepo, cacheRepo, log)
}
