package export

import (
	"github.com/kanshan/ServerStatus/data-server/internal/repository"
	"github.com/kanshan/ServerStatus/data-server/pkg/logger"
)

type repositoryProvider interface {
	ServerRepository() repository.ServerRepository
	HistoryRepository() repository.HistoryRepository
	CacheRepository() repository.CacheRepository
}

// ExportServiceFactory 导出服务工厂
type ExportServiceFactory struct {
	repository repositoryProvider
	logger     logger.Logger
}

// NewExportServiceFactory 创建导出服务工厂
func NewExportServiceFactory(repository repositoryProvider, logger logger.Logger) *ExportServiceFactory {
	return &ExportServiceFactory{
		repository: repository,
		logger:     logger,
	}
}

// CreateExportService 创建导出服务
func (f *ExportServiceFactory) CreateExportService() *ExportService {
	return NewExportService(
		f.repository.ServerRepository(),
		f.repository.HistoryRepository(),
		f.repository.CacheRepository(),
		f.logger,
	)
}

// CreateExportServiceWithDependencies 创建带依赖的导出服务
func (f *ExportServiceFactory) CreateExportServiceWithDependencies(
	serverRepo repository.ServerRepository,
	historyRepo repository.HistoryRepository,
	cacheRepo repository.CacheRepository,
) *ExportService {
	return NewExportService(serverRepo, historyRepo, cacheRepo, f.logger)
}
