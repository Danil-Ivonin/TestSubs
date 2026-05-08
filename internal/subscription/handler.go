package subscription

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type HandlerService interface {
	Create(ctx context.Context, req CreateRequest) (Response, error)
	Get(ctx context.Context, id uuid.UUID) (Response, error)
	List(ctx context.Context, filter ListFilter) ([]Response, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (Response, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Total(ctx context.Context, req TotalRequest) (TotalResponse, error)
}

type Handler struct {
	service HandlerService
	logger  *logrus.Logger
}

func NewHandler(service HandlerService, logger *logrus.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) RegisterRoutes(router gin.IRouter) {
	router.POST("/subscriptions", h.Create)
	router.GET("/subscriptions", h.List)
	router.GET("/subscriptions/total", h.Total)
	router.GET("/subscriptions/:id", h.Get)
	router.PUT("/subscriptions/:id", h.Update)
	router.DELETE("/subscriptions/:id", h.Delete)
}

// Create godoc
// @Summary Create subscription
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param request body CreateRequest true "Subscription payload"
// @Success 201 {object} Response
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions [post]
func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	resp, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// Get godoc
// @Summary Get subscription by ID
// @Tags subscriptions
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 200 {object} Response
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	id, ok := h.parsePathUUID(c, "id")
	if !ok {
		return
	}

	resp, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// List godoc
// @Summary List subscriptions
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "User ID"
// @Param service_name query string false "Service name"
// @Param limit query int false "Limit" minimum(1) maximum(100)
// @Param offset query int false "Offset" minimum(0)
// @Success 200 {array} Response
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions [get]
func (h *Handler) List(c *gin.Context) {
	filter, ok := h.listFilterFromQuery(c)
	if !ok {
		return
	}

	resp, err := h.service.List(c.Request.Context(), filter)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Update godoc
// @Summary Update subscription
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID"
// @Param request body UpdateRequest true "Subscription payload"
// @Success 200 {object} Response
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id, ok := h.parsePathUUID(c, "id")
	if !ok {
		return
	}

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	resp, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Delete godoc
// @Summary Delete subscription
// @Tags subscriptions
// @Param id path string true "Subscription ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id, ok := h.parsePathUUID(c, "id")
	if !ok {
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		h.writeServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// Total godoc
// @Summary Calculate subscriptions total
// @Description Calculates total subscription cost for an inclusive month period. Date format is MM-YYYY.
// @Tags subscriptions
// @Produce json
// @Param from query string true "Period start in MM-YYYY format"
// @Param to query string true "Period end in MM-YYYY format"
// @Param user_id query string false "User ID"
// @Param service_name query string false "Service name"
// @Success 200 {object} TotalResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/total [get]
func (h *Handler) Total(c *gin.Context) {
	req := TotalRequest{
		From:        c.Query("from"),
		To:          c.Query("to"),
		UserID:      c.Query("user_id"),
		ServiceName: c.Query("service_name"),
	}

	resp, err := h.service.Total(c.Request.Context(), req)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) parsePathUUID(c *gin.Context, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		h.writeError(c, http.StatusBadRequest, "invalid "+name, err)
		return uuid.Nil, false
	}

	return id, true
}

func (h *Handler) listFilterFromQuery(c *gin.Context) (ListFilter, bool) {
	var filter ListFilter

	if userIDValue := c.Query("user_id"); userIDValue != "" {
		userID, err := uuid.Parse(userIDValue)
		if err != nil {
			h.writeError(c, http.StatusBadRequest, "invalid user_id", err)
			return ListFilter{}, false
		}
		filter.UserID = &userID
	}

	if serviceName := c.Query("service_name"); serviceName != "" {
		filter.ServiceName = &serviceName
	}

	limit, ok := h.parseOptionalIntQuery(c, "limit")
	if !ok {
		return ListFilter{}, false
	}
	filter.Limit = limit

	offset, ok := h.parseOptionalIntQuery(c, "offset")
	if !ok {
		return ListFilter{}, false
	}
	filter.Offset = offset

	return filter, true
}

func (h *Handler) parseOptionalIntQuery(c *gin.Context, name string) (int, bool) {
	value := c.Query(name)
	if value == "" {
		return 0, true
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		h.writeError(c, http.StatusBadRequest, "invalid "+name, err)
		return 0, false
	}

	return parsed, true
}

func (h *Handler) writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		h.writeError(c, http.StatusNotFound, "subscription not found", err)
	case errors.Is(err, ErrInvalidDate),
		errors.Is(err, ErrInvalidPeriod),
		errors.Is(err, ErrInvalidPrice),
		errors.Is(err, ErrInvalidUUID):
		h.writeError(c, http.StatusBadRequest, "invalid request", err)
	default:
		h.writeError(c, http.StatusInternalServerError, "internal server error", err)
	}
}

func (h *Handler) writeError(c *gin.Context, status int, message string, err error) {
	if h.logger != nil {
		entry := h.logger.WithFields(logrus.Fields{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
			"status": status,
		})
		if err != nil {
			entry = entry.WithError(err)
		}
		if status >= http.StatusInternalServerError {
			entry.Error("http error response")
		} else {
			entry.Warn("http error response")
		}
	}

	c.AbortWithStatusJSON(status, gin.H{"error": message})
}
