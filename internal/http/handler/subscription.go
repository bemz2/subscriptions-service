package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"subscriptions-service/internal/domain"
	"subscriptions-service/internal/http/dto"
	"subscriptions-service/internal/service"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type SubscriptionService interface {
	Create(ctx context.Context, sub *domain.Subscription) (*domain.Subscription, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error)
	List(ctx context.Context, f domain.SubscriptionFilter) ([]domain.Subscription, error)
	Update(ctx context.Context, sub *domain.Subscription) (*domain.Subscription, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Sum(ctx context.Context, f domain.SubscriptionFilter) (int, error)
}

type SubscriptionHandler struct {
	service SubscriptionService
}

func NewSubscriptionHandler(svc SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{service: svc}
}

// Create godoc
// @Summary Create subscription
// @Description Creates a new subscription record.
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param request body dto.CreateSubscriptionRequest true "Subscription creation payload"
// @Success 201 {object} dto.SubscriptionResponse
// @Failure 400 {string} string "invalid request"
// @Router /subscriptions [post]
func (h *SubscriptionHandler) Create(c echo.Context) error {
	const contextKey = "SubscriptionHandler.Create"

	var req dto.CreateSubscriptionRequest

	if err := c.Bind(&req); err != nil {
		return httpError(http.StatusBadRequest, "invalid request", contextKey, err)
	}

	startDate, err := dto.ParseMonthYear(req.StartDate)
	if err != nil {
		return httpError(http.StatusBadRequest, err.Error(), contextKey, err)
	}

	sub := &domain.Subscription{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   startDate,
	}

	if req.EndDate != nil {
		endDate, err := dto.ParseMonthYear(*req.EndDate)
		if err != nil {
			return httpError(http.StatusBadRequest, err.Error(), contextKey, err)
		}
		sub.EndDate = &endDate
	}

	created, err := h.service.Create(c.Request().Context(), sub)
	if err != nil {
		return httpError(http.StatusInternalServerError, err.Error(), contextKey, err)
	}

	return c.JSON(http.StatusCreated, toResponse(created))
}

// GetByID godoc
// @Summary Get subscription by ID
// @Description Returns a single subscription by its UUID.
// @Tags subscriptions
// @Produce json
// @Param id path string true "Subscription ID" format(uuid)
// @Success 200 {object} dto.SubscriptionResponse
// @Failure 400 {string} string "invalid id"
// @Failure 404 {string} string "not found"
// @Router /subscriptions/{id} [get]
func (h *SubscriptionHandler) GetByID(c echo.Context) error {
	const contextKey = "SubscriptionHandler.GetByID"

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return httpError(http.StatusBadRequest, "invalid id", contextKey, err)
	}

	sub, err := h.service.GetByID(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return httpError(http.StatusNotFound, "subscription not found", contextKey, err)
		}
		return httpError(http.StatusInternalServerError, err.Error(), contextKey, err)
	}

	return c.JSON(http.StatusOK, toResponse(sub))
}

// List godoc
// @Summary List subscriptions
// @Description Returns a list of subscriptions with optional filters.
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "Filter by user ID" format(uuid)
// @Param service_name query string false "Filter by service name"
// @Param from query string false "Filter from date (MM-YYYY)"
// @Param to query string false "Filter to date (MM-YYYY)"
// @Success 200 {array} dto.SubscriptionResponse
// @Failure 400 {string} string "invalid filter params"
// @Router /subscriptions [get]
func (h *SubscriptionHandler) List(c echo.Context) error {
	const contextKey = "SubscriptionHandler.List"

	filter, err := parseFilter(c)
	if err != nil {
		return httpError(http.StatusBadRequest, err.Error(), contextKey, err)
	}

	subs, err := h.service.List(c.Request().Context(), filter)
	if err != nil {
		return httpError(http.StatusInternalServerError, err.Error(), contextKey, err)
	}

	resp := make([]dto.SubscriptionResponse, 0, len(subs))
	for _, s := range subs {
		resp = append(resp, toResponse(&s))
	}

	return c.JSON(http.StatusOK, resp)
}

// Update godoc
// @Summary Update subscription
// @Description Fully updates an existing subscription by its UUID.
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID" format(uuid)
// @Param request body dto.UpdateSubscriptionRequest true "Subscription update payload"
// @Success 200 {object} dto.SubscriptionResponse
// @Failure 400 {string} string "invalid request"
// @Failure 404 {string} string "not found"
// @Router /subscriptions/{id} [put]
func (h *SubscriptionHandler) Update(c echo.Context) error {
	const contextKey = "SubscriptionHandler.Update"

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return httpError(http.StatusBadRequest, "invalid id", contextKey, err)
	}

	var req dto.UpdateSubscriptionRequest
	if err := c.Bind(&req); err != nil {
		return httpError(http.StatusBadRequest, "invalid request", contextKey, err)
	}

	startDate, err := dto.ParseMonthYear(req.StartDate)
	if err != nil {
		return httpError(http.StatusBadRequest, err.Error(), contextKey, err)
	}

	sub := &domain.Subscription{
		ID:          id,
		ServiceName: req.ServiceName,
		Price:       req.Price,
		StartDate:   startDate,
	}

	if req.EndDate != nil {
		endDate, err := dto.ParseMonthYear(*req.EndDate)
		if err != nil {
			return httpError(http.StatusBadRequest, err.Error(), contextKey, err)
		}
		sub.EndDate = &endDate
	}

	updated, err := h.service.Update(c.Request().Context(), sub)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return httpError(http.StatusNotFound, "subscription not found", contextKey, err)
		}
		return httpError(http.StatusInternalServerError, err.Error(), contextKey, err)
	}

	return c.JSON(http.StatusOK, toResponse(updated))
}

// Delete godoc
// @Summary Delete subscription
// @Description Deletes a subscription by its UUID.
// @Tags subscriptions
// @Param id path string true "Subscription ID" format(uuid)
// @Success 204 "No Content"
// @Failure 400 {string} string "invalid id"
// @Failure 404 {string} string "not found"
// @Router /subscriptions/{id} [delete]
func (h *SubscriptionHandler) Delete(c echo.Context) error {
	const contextKey = "SubscriptionHandler.Delete"

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return httpError(http.StatusBadRequest, "invalid id", contextKey, err)
	}

	if err := h.service.Delete(c.Request().Context(), id); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return httpError(http.StatusNotFound, "subscription not found", contextKey, err)
		}
		return httpError(http.StatusInternalServerError, err.Error(), contextKey, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// Sum godoc
// @Summary Calculate total subscription cost
// @Description Returns the total cost of subscriptions active during the given period, with optional filters.
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "Filter by user ID" format(uuid)
// @Param service_name query string false "Filter by service name"
// @Param from query string false "Period start (MM-YYYY)"
// @Param to query string false "Period end (MM-YYYY)"
// @Success 200 {object} dto.SumResponse
// @Failure 400 {string} string "invalid filter params"
// @Router /subscriptions/sum [get]
func (h *SubscriptionHandler) Sum(c echo.Context) error {
	const contextKey = "SubscriptionHandler.Sum"

	filter, err := parseFilter(c)
	if err != nil {
		return httpError(http.StatusBadRequest, err.Error(), contextKey, err)
	}

	total, err := h.service.Sum(c.Request().Context(), filter)
	if err != nil {
		return httpError(http.StatusInternalServerError, err.Error(), contextKey, err)
	}

	return c.JSON(http.StatusOK, dto.SumResponse{Total: total})
}

// parseFilter extracts common query params (user_id, service_name, from, to) into a domain.SubscriptionFilter.
func parseFilter(c echo.Context) (domain.SubscriptionFilter, error) {
	const contextKey = "SubscriptionHandler.parseFilter"

	var filter domain.SubscriptionFilter

	if raw := c.QueryParam("user_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return filter, fmt.Errorf("%s: invalid user_id: %w", contextKey, err)
		}
		filter.UserID = &id
	}

	if raw := c.QueryParam("service_name"); raw != "" {
		filter.ServiceName = &raw
	}

	if raw := c.QueryParam("from"); raw != "" {
		t, err := dto.ParseMonthYear(raw)
		if err != nil {
			return filter, fmt.Errorf("%s: parse from: %w", contextKey, err)
		}
		filter.From = &t
	}

	if raw := c.QueryParam("to"); raw != "" {
		t, err := dto.ParseMonthYear(raw)
		if err != nil {
			return filter, fmt.Errorf("%s: parse to: %w", contextKey, err)
		}
		filter.To = &t
	}

	return filter, nil
}

func toResponse(s *domain.Subscription) dto.SubscriptionResponse {
	return dto.SubscriptionResponse{
		ID:          s.ID,
		ServiceName: s.ServiceName,
		Price:       s.Price,
		UserID:      s.UserID,
		StartDate:   dto.FormatMonthYear(s.StartDate),
		EndDate:     dto.FormatMonthYearPtr(s.EndDate),
	}
}

func httpError(status int, message string, contextKey string, err error) error {
	if err == nil {
		return echo.NewHTTPError(status, message)
	}

	return echo.NewHTTPError(status, message).SetInternal(fmt.Errorf("%s: %w", contextKey, err))
}
