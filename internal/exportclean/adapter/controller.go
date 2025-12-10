package adapter

import (
	"context"
	"fmt"

	"serverstatus-monitor/internal/exportclean/application"
	"serverstatus-monitor/internal/exportclean/domain"
)

// DTO 输入数据传输对象（例如来自 HTTP 层的解析结果）
type SubmitTaskDTO struct {
	ID         string            `json:"id"`
	ProjectKey string            `json:"project_key"`
	Format     string            `json:"format"`
	Type       string            `json:"type"`
	Limit      int               `json:"limit"`
	Filters    map[string]string `json:"filters"`
}

// Controller 负责 DTO → UseCase 的适配，不依赖具体框架
type Controller struct {
	service *application.SubmitTaskService
}

func NewController(service *application.SubmitTaskService) *Controller {
	return &Controller{service: service}
}

// HandleSubmit 适配输入并调用用例
func (c *Controller) HandleSubmit(ctx context.Context, dto *SubmitTaskDTO) error {
	if dto == nil {
		return fmt.Errorf("dto is nil")
	}

	task := &domain.ExportTask{
		ID:         dto.ID,
		ProjectKey: dto.ProjectKey,
		Format:     domain.TaskFormat(dto.Format),
		Type:       domain.TaskType(dto.Type),
		Limit:      dto.Limit,
		Filters:    dto.Filters,
	}
	return c.service.Submit(ctx, task)
}
