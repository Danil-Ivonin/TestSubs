package httpserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func TestRouterHealthReturnsOK(t *testing.T) {
	t.Setenv("GIN_MODE", gin.TestMode)

	logger := logrus.New()
	router := NewRouter(logger, RouterOptions{})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("status body = %q, want %q", body["status"], "ok")
	}
}

func TestRequestLoggerWritesRequestMetadata(t *testing.T) {
	t.Setenv("GIN_MODE", gin.TestMode)

	var logs bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&logs)
	logger.SetFormatter(&logrus.JSONFormatter{})

	router := NewRouter(logger, RouterOptions{})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	line := logs.String()
	if !strings.Contains(line, `"method":"GET"`) {
		t.Fatalf("log line %q does not contain method", line)
	}
	if !strings.Contains(line, `"path":"/health"`) {
		t.Fatalf("log line %q does not contain path", line)
	}
	if !strings.Contains(line, `"status":200`) {
		t.Fatalf("log line %q does not contain status", line)
	}
}
