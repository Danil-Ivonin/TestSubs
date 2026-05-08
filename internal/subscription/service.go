package subscription

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (Response, error) {
	sub, err := subscriptionFromCreateRequest(req)
	if err != nil {
		return Response{}, err
	}

	created, err := s.repo.Create(ctx, sub)
	if err != nil {
		return Response{}, fmt.Errorf("creating subscription: %w", err)
	}

	return responseFromSubscription(created), nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (Response, error) {
	sub, err := s.repo.Get(ctx, id)
	if err != nil {
		return Response{}, fmt.Errorf("getting subscription: %w", err)
	}

	return responseFromSubscription(sub), nil
}

func (s *Service) List(ctx context.Context, filter ListFilter) ([]Response, error) {
	subs, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("listing subscriptions: %w", err)
	}

	responses := make([]Response, 0, len(subs))
	for _, sub := range subs {
		responses = append(responses, responseFromSubscription(sub))
	}

	return responses, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (Response, error) {
	sub, err := subscriptionFromUpdateRequest(id, req)
	if err != nil {
		return Response{}, err
	}

	updated, err := s.repo.Update(ctx, sub)
	if err != nil {
		return Response{}, fmt.Errorf("updating subscription: %w", err)
	}

	return responseFromSubscription(updated), nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("deleting subscription: %w", err)
	}

	return nil
}

func (s *Service) Total(ctx context.Context, req TotalRequest) (TotalResponse, error) {
	filter, err := totalFilterFromRequest(req)
	if err != nil {
		return TotalResponse{}, err
	}

	total, err := s.repo.Total(ctx, filter)
	if err != nil {
		return TotalResponse{}, fmt.Errorf("calculating subscriptions total: %w", err)
	}

	return TotalResponse{Total: total}, nil
}

func subscriptionFromCreateRequest(req CreateRequest) (Subscription, error) {
	return subscriptionFromFields(uuid.Nil, req.ServiceName, req.Price, req.UserID, req.StartDate, req.EndDate)
}

func subscriptionFromUpdateRequest(id uuid.UUID, req UpdateRequest) (Subscription, error) {
	return subscriptionFromFields(id, req.ServiceName, req.Price, req.UserID, req.StartDate, req.EndDate)
}

func subscriptionFromFields(id uuid.UUID, serviceName string, price int, userIDValue string, startDateValue string, endDateValue *string) (Subscription, error) {
	if price <= 0 {
		return Subscription{}, ErrInvalidPrice
	}

	userID, err := uuid.Parse(userIDValue)
	if err != nil {
		return Subscription{}, fmt.Errorf("%w: user_id", ErrInvalidUUID)
	}

	startDate, err := ParseMonthYear(startDateValue)
	if err != nil {
		return Subscription{}, err
	}

	parsedEndDate, err := parseOptionalEndDate(endDateValue)
	if err != nil {
		return Subscription{}, err
	}
	if parsedEndDate != nil && parsedEndDate.Before(startDate) {
		return Subscription{}, ErrInvalidPeriod
	}

	return Subscription{
		ID:          id,
		ServiceName: strings.TrimSpace(serviceName),
		Price:       price,
		UserID:      userID,
		StartDate:   startDate,
		EndDate:     parsedEndDate,
	}, nil
}

func totalFilterFromRequest(req TotalRequest) (TotalFilter, error) {
	from, err := ParseMonthYear(req.From)
	if err != nil {
		return TotalFilter{}, err
	}

	to, err := ParseMonthYear(req.To)
	if err != nil {
		return TotalFilter{}, err
	}
	if to.Before(from) {
		return TotalFilter{}, ErrInvalidPeriod
	}

	var userID *uuid.UUID
	if strings.TrimSpace(req.UserID) != "" {
		parsed, err := uuid.Parse(req.UserID)
		if err != nil {
			return TotalFilter{}, fmt.Errorf("%w: user_id", ErrInvalidUUID)
		}
		userID = &parsed
	}

	var serviceName *string
	if strings.TrimSpace(req.ServiceName) != "" {
		value := strings.TrimSpace(req.ServiceName)
		serviceName = &value
	}

	return TotalFilter{
		From:        from,
		To:          to,
		UserID:      userID,
		ServiceName: serviceName,
	}, nil
}

func parseOptionalEndDate(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}

	parsed, err := ParseMonthYear(*value)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}

func responseFromSubscription(sub Subscription) Response {
	var endDate *string
	if sub.EndDate != nil {
		formatted := FormatMonthYear(*sub.EndDate)
		endDate = &formatted
	}

	return Response{
		ID:          sub.ID.String(),
		ServiceName: sub.ServiceName,
		Price:       sub.Price,
		UserID:      sub.UserID.String(),
		StartDate:   FormatMonthYear(sub.StartDate),
		EndDate:     endDate,
	}
}
