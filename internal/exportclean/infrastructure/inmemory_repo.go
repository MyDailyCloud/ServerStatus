package infrastructure

import (
	"context"
	"sync"

	"serverstatus-monitor/internal/exportclean/domain"
)

type InMemoryTaskRepo struct {
	mu    sync.Mutex
	tasks []*domain.ExportTask
}

func NewInMemoryTaskRepo() *InMemoryTaskRepo {
	return &InMemoryTaskRepo{tasks: make([]*domain.ExportTask, 0)}
}

func (r *InMemoryTaskRepo) Save(_ context.Context, task *domain.ExportTask) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	copyTask := *task
	r.tasks = append(r.tasks, &copyTask)
	return nil
}

// All returns a copy of tasks for testing
func (r *InMemoryTaskRepo) All() []*domain.ExportTask {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]*domain.ExportTask, len(r.tasks))
	copy(out, r.tasks)
	return out
}
