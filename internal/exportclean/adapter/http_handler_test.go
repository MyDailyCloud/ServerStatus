package adapter

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"serverstatus-monitor/internal/exportclean/application"
	"serverstatus-monitor/internal/exportclean/domain"
	"serverstatus-monitor/internal/exportclean/infrastructure"
)

func TestHTTPHandler_Success(t *testing.T) {
	s := loadSamplesAdapter(t)
	repo := infrastructure.NewInMemoryTaskRepo()
	svc := application.NewSubmitTaskService(repo)
	ctrl := NewController(svc)
	h := NewHTTPHandler(ctrl)
	server := httptest.NewServer(h)
	defer server.Close()

	body, _ := json.Marshal(s.Perfect)
	req, _ := http.NewRequest(http.MethodPost, server.URL+"/api/export/tasks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if len(repo.All()) != 1 {
		t.Fatalf("expected 1 task saved")
	}
}

func TestHTTPHandler_InvalidFormat(t *testing.T) {
	s := loadSamplesAdapter(t)
	raw := s.Invalid
	raw.ProjectKey = "proj-prod-insight"
	raw.Limit = 10

	repo := infrastructure.NewInMemoryTaskRepo()
	svc := application.NewSubmitTaskService(repo)
	ctrl := NewController(svc)
	h := NewHTTPHandler(ctrl)
	server := httptest.NewServer(h)
	defer server.Close()

	body, _ := json.Marshal(raw)
	req, _ := http.NewRequest(http.MethodPost, server.URL+"/api/export/tasks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHTTPHandler_BadCharsProject(t *testing.T) {
	s := loadSamplesAdapter(t)
	raw := s.BadCharsP

	repo := infrastructure.NewInMemoryTaskRepo()
	svc := application.NewSubmitTaskService(repo)
	ctrl := NewController(svc)
	h := NewHTTPHandler(ctrl)
	server := httptest.NewServer(h)
	defer server.Close()

	body, _ := json.Marshal(raw)
	req, _ := http.NewRequest(http.MethodPost, server.URL+"/api/export/tasks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	bodyBytes, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(bodyBytes), "invalid characters") {
		t.Fatalf("expected invalid characters error")
	}
}

func TestHTTPHandler_UnsupportedMediaType(t *testing.T) {
	repo := infrastructure.NewInMemoryTaskRepo()
	svc := application.NewSubmitTaskService(repo)
	ctrl := NewController(svc)
	h := NewHTTPHandler(ctrl)
	server := httptest.NewServer(h)
	defer server.Close()

	req, _ := http.NewRequest(http.MethodPost, server.URL+"/api/export/tasks", bytes.NewReader([]byte("{}")))
	// missing Content-Type
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnsupportedMediaType {
		t.Fatalf("expected 415, got %d", resp.StatusCode)
	}
}

// compile-time check to avoid import pruning
var _ = domain.ExportTask{}
var _ = application.ErrInvalidFormat
