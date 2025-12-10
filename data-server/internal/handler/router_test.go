package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	servicehealth "github.com/kanshan/ServerStatus/data-server/internal/service/health"
	"github.com/kanshan/ServerStatus/data-server/pkg/logger"
)

type mockHealthService struct {
	resp *servicehealth.HealthResponse
	err  error
}

func (m *mockHealthService) CheckHealth(_ context.Context, _ *servicehealth.CheckRequest) (*servicehealth.HealthResponse, error) {
	return m.resp, m.err
}

func (m *mockHealthService) GetServiceInfo(_ context.Context) map[string]interface{} {
	return map[string]interface{}{"mock": true}
}

func TestHealthHandlerRespondsJSON(t *testing.T) {
	l := logger.GetDefaultLogger()
	h := &mockHealthService{resp: &servicehealth.HealthResponse{Status: servicehealth.StatusHealthy}}

	router, err := BuildHTTPHandler(&HandlerDependencies{
		HealthService: h,
		Logger:        l,
	}, DefaultHandlerConfig())
	if err != nil {
		t.Fatalf("build router: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json, got %s", ct)
	}
	if rid := rr.Header().Get("X-Request-ID"); rid == "" {
		t.Fatalf("expected X-Request-ID header")
	}
	var body map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}
	if body["status"] != string(servicehealth.StatusHealthy) {
		t.Fatalf("unexpected status: %v", body["status"])
	}
}

func TestNotFoundReturnsJSON(t *testing.T) {
	l := logger.GetDefaultLogger()
	h := &mockHealthService{resp: &servicehealth.HealthResponse{Status: servicehealth.StatusHealthy}}

	router, err := BuildHTTPHandler(&HandlerDependencies{
		HealthService: h,
		Logger:        l,
	}, DefaultHandlerConfig())
	if err != nil {
		t.Fatalf("build router: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/not-found", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json, got %s", ct)
	}
	var body map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}
	if body["error"] != "not_found" {
		t.Fatalf("unexpected error code: %v", body["error"])
	}
}

func TestRequireJSONEnforcedOnPost(t *testing.T) {
	l := logger.GetDefaultLogger()
	h := &mockHealthService{resp: &servicehealth.HealthResponse{Status: servicehealth.StatusHealthy}}

	router, err := BuildHTTPHandler(&HandlerDependencies{
		HealthService: h,
		Logger:        l,
	}, DefaultHandlerConfig())
	if err != nil {
		t.Fatalf("build router: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/health", nil) // POST without Content-Type
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected 415, got %d", rr.Code)
	}
	var body map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}
	if body["error"] != "unsupported_media_type" {
		t.Fatalf("unexpected error code: %v", body["error"])
	}
}
