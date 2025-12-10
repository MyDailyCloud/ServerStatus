package domain

import "context"

type TaskRepository interface {
	Save(ctx context.Context, task *ExportTask) error
}
