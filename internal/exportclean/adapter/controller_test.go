package adapter

import (
	"context"
	"errors"
	"testing"

	"serverstatus-monitor/internal/exportclean/application"
	"serverstatus-monitor/internal/exportclean/domain"
	"serverstatus-monitor/internal/exportclean/infrastructure"
)

type failingRepo struct{}

func (f *failingRepo) Save(ctx context.Context, task *domain.ExportTask) error {
	return errors.New("persist failed")
}

func TestController_Success_Perfect(t *testing.T) {
	s := loadSamplesAdapter(t)
	repo := infrastructure.NewInMemoryTaskRepo()
	uc := application.NewSubmitTaskService(repo)
	ctrl := NewController(uc)

	if err := ctrl.HandleSubmit(context.Background(), toDTO(s.Perfect)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.All()) != 1 {
		t.Fatalf("expected 1 task saved")
	}
}

func TestController_InvalidFormat(t *testing.T) {
	s := loadSamplesAdapter(t)
	raw := s.Invalid
	raw.ProjectKey = "proj-prod-insight"
	raw.Limit = 10

	repo := infrastructure.NewInMemoryTaskRepo()
	uc := application.NewSubmitTaskService(repo)
	ctrl := NewController(uc)

	err := ctrl.HandleSubmit(context.Background(), toDTO(raw))
	if err == nil || !errors.Is(err, application.ErrInvalidFormat) {
		t.Fatalf("expected ErrInvalidFormat, got %v", err)
	}
	if len(repo.All()) != 0 {
		t.Fatalf("expected 0 task saved")
	}
}

func TestController_MissingDTO(t *testing.T) {
	repo := infrastructure.NewInMemoryTaskRepo()
	uc := application.NewSubmitTaskService(repo)
	ctrl := NewController(uc)

	err := ctrl.HandleSubmit(context.Background(), nil)
	if err == nil {
		t.Fatalf("expected error for nil dto")
	}
}

func TestController_RepoError(t *testing.T) {
	s := loadSamplesAdapter(t)
	repo := &failingRepo{}
	uc := application.NewSubmitTaskService(repo)
	ctrl := NewController(uc)

	err := ctrl.HandleSubmit(context.Background(), toDTO(s.Perfect))
	if err == nil {
		t.Fatalf("expected repo error")
	}
}
