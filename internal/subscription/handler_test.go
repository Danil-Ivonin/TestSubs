package subscription

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type stubService struct {
	createFn func(ctx context.Context, req CreateRequest) (Response, error)
	getFn    func(ctx context.Context, id uuid.UUID) (Response, error)
	listFn   func(ctx context.Context, filter ListFilter) ([]Response, error)
	updateFn func(ctx context.Context, id uuid.UUID, req UpdateRequest) (Response, error)
	deleteFn func(ctx context.Context, id uuid.UUID) error
	totalFn  func(ctx context.Context, req TotalRequest) (TotalResponse, error)
}

func (s stubService) Create(ctx context.Context, req CreateRequest) (Response, error) {
	return s.createFn(ctx, req)
}

func (s stubService) Get(ctx context.Context, id uuid.UUID) (Response, error) {
	return s.getFn(ctx, id)
}

func (s stubService) List(ctx context.Context, filter ListFilter) ([]Response, error) {
	return s.listFn(ctx, filter)
}

func (s stubService) Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (Response, error) {
	return s.updateFn(ctx, id, req)
}

func (s stubService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.deleteFn(ctx, id)
}

func (s stubService) Total(ctx context.Context, req TotalRequest) (TotalResponse, error) {
	return s.totalFn(ctx, req)
}

func TestHandlerCreateSubscription(t *testing.T) {
	t.Setenv("GIN_MODE", gin.TestMode)

	userID := uuid.New()
	subID := uuid.New()
	handler := NewHandler(stubService{
		createFn: func(ctx context.Context, req CreateRequest) (Response, error) {
			if ctx == nil {
				t.Fatal("ctx is nil")
			}
			if req.ServiceName != "Yandex Plus" {
				t.Fatalf("ServiceName = %q", req.ServiceName)
			}
			if req.Price != 400 {
				t.Fatalf("Price = %d", req.Price)
			}
			if req.UserID != userID.String() {
				t.Fatalf("UserID = %q", req.UserID)
			}
			if req.StartDate != "07-2025" {
				t.Fatalf("StartDate = %q", req.StartDate)
			}
			return Response{
				ID:          subID.String(),
				ServiceName: req.ServiceName,
				Price:       req.Price,
				UserID:      req.UserID,
				StartDate:   req.StartDate,
			}, nil
		},
	}, nil)
	router := newTestRouter(handler)

	body := []byte(`{"service_name":"Yandex Plus","price":400,"user_id":"` + userID.String() + `","start_date":"07-2025"}`)
	rec := performRequest(router, http.MethodPost, "/api/v1/subscriptions", body)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var got Response
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.ID != subID.String() {
		t.Fatalf("ID = %q, want %q", got.ID, subID.String())
	}
}

func TestHandlerListSubscriptionsParsesFilters(t *testing.T) {
	t.Setenv("GIN_MODE", gin.TestMode)

	userID := uuid.New()
	serviceName := "Yandex Plus"
	handler := NewHandler(stubService{
		listFn: func(ctx context.Context, filter ListFilter) ([]Response, error) {
			if filter.UserID == nil || *filter.UserID != userID {
				t.Fatalf("UserID = %v, want %s", filter.UserID, userID)
			}
			if filter.ServiceName == nil || *filter.ServiceName != serviceName {
				t.Fatalf("ServiceName = %v, want %s", filter.ServiceName, serviceName)
			}
			if filter.Limit != 10 {
				t.Fatalf("Limit = %d, want %d", filter.Limit, 10)
			}
			if filter.Offset != 5 {
				t.Fatalf("Offset = %d, want %d", filter.Offset, 5)
			}
			return []Response{}, nil
		},
	}, nil)
	router := newTestRouter(handler)

	rec := performRequest(router, http.MethodGet, "/api/v1/subscriptions?user_id="+userID.String()+"&service_name=Yandex%20Plus&limit=10&offset=5", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestHandlerGetSubscriptionInvalidID(t *testing.T) {
	t.Setenv("GIN_MODE", gin.TestMode)

	handler := NewHandler(stubService{}, nil)
	router := newTestRouter(handler)

	rec := performRequest(router, http.MethodGet, "/api/v1/subscriptions/not-a-uuid", nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandlerGetSubscriptionNotFound(t *testing.T) {
	t.Setenv("GIN_MODE", gin.TestMode)

	handler := NewHandler(stubService{
		getFn: func(ctx context.Context, id uuid.UUID) (Response, error) {
			return Response{}, ErrNotFound
		},
	}, nil)
	router := newTestRouter(handler)

	rec := performRequest(router, http.MethodGet, "/api/v1/subscriptions/"+uuid.New().String(), nil)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestHandlerTotalSubscriptions(t *testing.T) {
	t.Setenv("GIN_MODE", gin.TestMode)

	userID := uuid.New()
	handler := NewHandler(stubService{
		totalFn: func(ctx context.Context, req TotalRequest) (TotalResponse, error) {
			if req.From != "07-2025" {
				t.Fatalf("From = %q", req.From)
			}
			if req.To != "09-2025" {
				t.Fatalf("To = %q", req.To)
			}
			if req.UserID != userID.String() {
				t.Fatalf("UserID = %q", req.UserID)
			}
			if req.ServiceName != "Yandex Plus" {
				t.Fatalf("ServiceName = %q", req.ServiceName)
			}
			return TotalResponse{Total: 1200}, nil
		},
	}, nil)
	router := newTestRouter(handler)

	rec := performRequest(router, http.MethodGet, "/api/v1/subscriptions/total?from=07-2025&to=09-2025&user_id="+userID.String()+"&service_name=Yandex%20Plus", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"total":1200`)) {
		t.Fatalf("body = %s, want total", rec.Body.String())
	}
}

func TestHandlerMapsValidationError(t *testing.T) {
	t.Setenv("GIN_MODE", gin.TestMode)

	handler := NewHandler(stubService{
		createFn: func(ctx context.Context, req CreateRequest) (Response, error) {
			return Response{}, ErrInvalidPrice
		},
	}, nil)
	router := newTestRouter(handler)

	body := []byte(`{"service_name":"Yandex Plus","price":0,"user_id":"` + uuid.New().String() + `","start_date":"07-2025"}`)
	rec := performRequest(router, http.MethodPost, "/api/v1/subscriptions", body)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandlerMapsUnexpectedError(t *testing.T) {
	t.Setenv("GIN_MODE", gin.TestMode)

	handler := NewHandler(stubService{
		getFn: func(ctx context.Context, id uuid.UUID) (Response, error) {
			return Response{}, errors.New("database unavailable")
		},
	}, nil)
	router := newTestRouter(handler)

	rec := performRequest(router, http.MethodGet, "/api/v1/subscriptions/"+uuid.New().String(), nil)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestHandlerLogsReturnedError(t *testing.T) {
	t.Setenv("GIN_MODE", gin.TestMode)

	var logs bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&logs)
	logger.SetFormatter(&logrus.JSONFormatter{})

	handler := NewHandler(stubService{
		getFn: func(ctx context.Context, id uuid.UUID) (Response, error) {
			return Response{}, errors.New("database unavailable")
		},
	}, logger)
	router := newTestRouter(handler)

	rec := performRequest(router, http.MethodGet, "/api/v1/subscriptions/"+uuid.New().String(), nil)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	if !bytes.Contains(logs.Bytes(), []byte(`"error":"database unavailable"`)) {
		t.Fatalf("logs = %s, want error field", logs.String())
	}
	if !bytes.Contains(logs.Bytes(), []byte(`"status":500`)) {
		t.Fatalf("logs = %s, want status field", logs.String())
	}
}

func newTestRouter(handler *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	api := router.Group("/api/v1")
	handler.RegisterRoutes(api)
	return router
}

func performRequest(router http.Handler, method string, path string, body []byte) *httptest.ResponseRecorder {
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		reader = bytes.NewReader(body)
	}

	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}
