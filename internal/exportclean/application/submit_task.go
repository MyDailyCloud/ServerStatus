package application

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"serverstatus-monitor/internal/exportclean/domain"
)

var (
	ErrInvalidFormat = errors.New("invalid format")
	ErrInvalidType   = errors.New("invalid type")
	ErrInvalidLimit  = errors.New("invalid limit")
	ErrMissingID     = errors.New("id is required")
)

const maxLimit = 10000
const maxProjectKeyLen = 128
const maxFilterValueLen = 256
const maxFilterKeyLen = 64

var allowedPattern = regexp.MustCompile(`^[\p{L}\p{N}._:-]+$`)

type SubmitTaskService struct {
	repo domain.TaskRepository
}

func NewSubmitTaskService(repo domain.TaskRepository) *SubmitTaskService {
	return &SubmitTaskService{repo: repo}
}

func (s *SubmitTaskService) Submit(ctx context.Context, task *domain.ExportTask) error {
	if err := validate(task); err != nil {
		return err
	}
	return s.repo.Save(ctx, task)
}

func validate(task *domain.ExportTask) error {
	if task == nil {
		return errors.New("task is nil")
	}
	if task.ID == "" {
		return ErrMissingID
	}
	task.ProjectKey = strings.TrimSpace(task.ProjectKey)
	if task.ProjectKey == "" {
		return errors.New("project_key is required")
	}
	if len(task.ProjectKey) > maxProjectKeyLen {
		return fmt.Errorf("project_key too long, max %d", maxProjectKeyLen)
	}
	if !allowedPattern.MatchString(task.ProjectKey) {
		return errors.New("project_key contains invalid characters")
	}
	if task.Limit <= 0 || task.Limit > maxLimit {
		return ErrInvalidLimit
	}
	for k, v := range task.Filters {
		if len(k) > maxFilterKeyLen {
			return fmt.Errorf("filter key too long: %s", k)
		}
		if len(v) > maxFilterValueLen {
			return fmt.Errorf("filter value too long for key %s", k)
		}
		if k != "" && !allowedPattern.MatchString(k) {
			return fmt.Errorf("filter key contains invalid characters: %s", k)
		}
		if v != "" && !allowedPattern.MatchString(v) {
			return fmt.Errorf("filter value contains invalid characters: %s", k)
		}
	}
	switch task.Format {
	case domain.FormatCSV, domain.FormatJSON:
	default:
		return ErrInvalidFormat
	}
	switch task.Type {
	case domain.TypeServers, domain.TypeHistory, domain.TypeUserResources:
	default:
		return ErrInvalidType
	}
	return nil
}
