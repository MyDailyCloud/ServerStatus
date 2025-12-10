package tests

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"serverstatus-monitor/internal/exportclean/application"
	"serverstatus-monitor/internal/exportclean/domain"
	"serverstatus-monitor/internal/exportclean/infrastructure"
)

type failingRepo struct{}

func (f *failingRepo) Save(ctx context.Context, task *domain.ExportTask) error {
	return errors.New("persist failed")
}

type taskRaw struct {
	ID         string            `json:"id"`
	ProjectKey string            `json:"project_key"`
	Format     string            `json:"format"`
	Type       string            `json:"type"`
	Limit      int               `json:"limit"`
	Filters    map[string]string `json:"filters"`
}

type samples struct {
	Perfect   taskRaw `json:"perfect_sample"`
	Edge      taskRaw `json:"edge_sample"`
	Invalid   taskRaw `json:"invalid_sample"`
	TooLarge  taskRaw `json:"too_large_limit"`
	MissingID taskRaw `json:"missing_id"`
	TooLongPK taskRaw `json:"too_long_project"`
	TooLongFV taskRaw `json:"too_long_filter_value"`
	BadCharsP taskRaw `json:"invalid_char_project"`
	BadCharsF taskRaw `json:"invalid_char_filter"`
}

func (r taskRaw) toDomain() domain.ExportTask {
	return domain.ExportTask{
		ID:         r.ID,
		ProjectKey: r.ProjectKey,
		Format:     domain.TaskFormat(r.Format),
		Type:       domain.TaskType(r.Type),
		Limit:      r.Limit,
		Filters:    r.Filters,
	}
}

func loadSamples(t *testing.T) samples {
	t.Helper()
	path := filepath.Join("..", "fixtures", "api_samples.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixtures: %v", err)
	}
	var s samples
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatalf("unmarshal fixtures: %v", err)
	}
	return s
}

func TestSubmitTask_Success_Perfect(t *testing.T) {
	s := loadSamples(t)
	repo := infrastructure.NewInMemoryTaskRepo()
	svc := application.NewSubmitTaskService(repo)

	task := s.Perfect.toDomain()
	if err := svc.Submit(context.Background(), &task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.All()) != 1 {
		t.Fatalf("expected 1 task saved")
	}
}

func TestSubmitTask_Success_Edge(t *testing.T) {
	s := loadSamples(t)
	repo := infrastructure.NewInMemoryTaskRepo()
	svc := application.NewSubmitTaskService(repo)

	task := s.Edge.toDomain()
	if err := svc.Submit(context.Background(), &task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.All()) != 1 {
		t.Fatalf("expected 1 task saved")
	}
}

func TestSubmitTask_InvalidFormat(t *testing.T) {
	s := loadSamples(t)
	task := s.Invalid
	task.ProjectKey = "proj-prod-insight"
	task.Limit = 10
	d := task.toDomain()

	repo := infrastructure.NewInMemoryTaskRepo()
	svc := application.NewSubmitTaskService(repo)

	err := svc.Submit(context.Background(), &d)
	if err == nil || !errors.Is(err, application.ErrInvalidFormat) {
		t.Fatalf("expected ErrInvalidFormat, got %v", err)
	}
	if len(repo.All()) != 0 {
		t.Fatalf("expected 0 task saved")
	}
}

func TestSubmitTask_InvalidType(t *testing.T) {
	s := loadSamples(t)
	task := s.Invalid
	task.ProjectKey = "proj-prod-insight"
	task.Limit = 10
	task.Format = "csv"
	d := task.toDomain()

	repo := infrastructure.NewInMemoryTaskRepo()
	svc := application.NewSubmitTaskService(repo)

	err := svc.Submit(context.Background(), &d)
	if err == nil || !errors.Is(err, application.ErrInvalidType) {
		t.Fatalf("expected ErrInvalidType, got %v", err)
	}
	if len(repo.All()) != 0 {
		t.Fatalf("expected 0 task saved")
	}
}

func TestSubmitTask_InvalidLimitZero(t *testing.T) {
	s := loadSamples(t)
	task := s.Invalid
	task.ProjectKey = "proj-prod-insight"
	task.Format = "csv"
	task.Type = "servers"
	task.Limit = 0
	d := task.toDomain()

	repo := infrastructure.NewInMemoryTaskRepo()
	svc := application.NewSubmitTaskService(repo)

	err := svc.Submit(context.Background(), &d)
	if err == nil || !errors.Is(err, application.ErrInvalidLimit) {
		t.Fatalf("expected ErrInvalidLimit, got %v", err)
	}
	if len(repo.All()) != 0 {
		t.Fatalf("expected 0 task saved")
	}
}

func TestSubmitTask_InvalidLimitTooLarge(t *testing.T) {
	s := loadSamples(t)
	task := s.TooLarge
	d := task.toDomain()

	repo := infrastructure.NewInMemoryTaskRepo()
	svc := application.NewSubmitTaskService(repo)

	err := svc.Submit(context.Background(), &d)
	if err == nil || !errors.Is(err, application.ErrInvalidLimit) {
		t.Fatalf("expected ErrInvalidLimit for too large limit, got %v", err)
	}
	if len(repo.All()) != 0 {
		t.Fatalf("expected 0 task saved")
	}
}

func TestSubmitTask_MissingID(t *testing.T) {
	s := loadSamples(t)
	task := s.MissingID
	d := task.toDomain()

	repo := infrastructure.NewInMemoryTaskRepo()
	svc := application.NewSubmitTaskService(repo)

	err := svc.Submit(context.Background(), &d)
	if err == nil || !errors.Is(err, application.ErrMissingID) {
		t.Fatalf("expected ErrMissingID, got %v", err)
	}
	if len(repo.All()) != 0 {
		t.Fatalf("expected 0 task saved")
	}
}

func TestSubmitTask_ProjectKeyTrimmed(t *testing.T) {
	s := loadSamples(t)
	task := s.Perfect
	task.ProjectKey = "  proj-prod-insight  "
	d := task.toDomain()

	repo := infrastructure.NewInMemoryTaskRepo()
	svc := application.NewSubmitTaskService(repo)

	if err := svc.Submit(context.Background(), &d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.All()) != 1 {
		t.Fatalf("expected 1 task saved")
	}
	if repo.All()[0].ProjectKey != "proj-prod-insight" {
		t.Fatalf("expected trimmed project key, got %q", repo.All()[0].ProjectKey)
	}
}

func TestSubmitTask_RepoError(t *testing.T) {
	s := loadSamples(t)
	repo := &failingRepo{}
	svc := application.NewSubmitTaskService(repo)

	task := s.Perfect.toDomain()
	err := svc.Submit(context.Background(), &task)
	if err == nil {
		t.Fatalf("expected repo error")
	}
}

func TestSubmitTask_ProjectKeyTooLong(t *testing.T) {
	s := loadSamples(t)
	task := s.TooLongPK
	d := task.toDomain()

	repo := infrastructure.NewInMemoryTaskRepo()
	svc := application.NewSubmitTaskService(repo)

	err := svc.Submit(context.Background(), &d)
	if err == nil {
		t.Fatalf("expected error for long project_key")
	}
	if len(repo.All()) != 0 {
		t.Fatalf("expected 0 task saved")
	}
}

func TestSubmitTask_FilterValueTooLong(t *testing.T) {
	s := loadSamples(t)
	task := s.TooLongFV
	d := task.toDomain()

	repo := infrastructure.NewInMemoryTaskRepo()
	svc := application.NewSubmitTaskService(repo)

	err := svc.Submit(context.Background(), &d)
	if err == nil {
		t.Fatalf("expected error for long filter value")
	}
	if len(repo.All()) != 0 {
		t.Fatalf("expected 0 task saved")
	}
}

func TestSubmitTask_ProjectKeyBadChars(t *testing.T) {
	s := loadSamples(t)
	task := s.BadCharsP
	d := task.toDomain()

	repo := infrastructure.NewInMemoryTaskRepo()
	svc := application.NewSubmitTaskService(repo)

	err := svc.Submit(context.Background(), &d)
	if err == nil || !strings.Contains(err.Error(), "invalid characters") {
		t.Fatalf("expected invalid characters error, got %v", err)
	}
	if len(repo.All()) != 0 {
		t.Fatalf("expected 0 task saved")
	}
}

func TestSubmitTask_FilterValueBadChars(t *testing.T) {
	s := loadSamples(t)
	task := s.BadCharsF
	d := task.toDomain()

	repo := infrastructure.NewInMemoryTaskRepo()
	svc := application.NewSubmitTaskService(repo)

	err := svc.Submit(context.Background(), &d)
	if err == nil || !strings.Contains(err.Error(), "invalid characters") {
		t.Fatalf("expected invalid characters error, got %v", err)
	}
	if len(repo.All()) != 0 {
		t.Fatalf("expected 0 task saved")
	}
}
