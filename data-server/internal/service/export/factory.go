package export

import (
	"github.com/kanshan/ServerStatus/data-server/internal/repository"
	"github.com/kanshan/ServerStatus/data-server/pkg/logger"
)

// ExportServiceFactory 导出服务工厂
type ExportServiceFactory struct {
	repository repository.Repository
	logger     logger.Logger
}

// NewExportServiceFactory 创建导出服务工厂
func NewExportServiceFactory(repository repository.Repository, logger logger.Logger) *ExportServiceFactory {
	return &ExportServiceFactory{
		repository: repository,
		logger:     logger,
	}
}

// CreateExportService 创建导出服务
func (f *ExportServiceFactory) CreateExportService() *ExportService {
	// Try to access embedded repositories
	switch repo := f.repository.(type) {
	case interface{
		HistoryRepository() repository.HistoryRepository
		CacheRepository() repository.CacheRepository
	}:
		return NewExportService(
			repo.HistoryRepository(),
			repo.CacheRepository(),
			f.logger,
		)
	default:
		// Fallback: try direct casting if the repository implements the interfaces directly
		if historyRepo, ok := f.repository.(repository.HistoryRepository); ok {
			if cacheRepo, ok := f.repository.(repository.CacheRepository); ok {
				return NewExportService(historyRepo, cacheRepo, f.logger)
			}
		}
		// As a last resort, return nil or panic
		panic("repository does not implement required interfaces")
	}
}

// CreateExportServiceWithDependencies 创建带依赖的导出服务
func (f *ExportServiceFactory) CreateExportServiceWithDependencies(
	historyRepo repository.HistoryRepository,
	cacheRepo repository.CacheRepository,
) *ExportService {
	return NewExportService(historyRepo, cacheRepo, f.logger)
}