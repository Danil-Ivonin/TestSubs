package subscription

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type stubRepository struct {
	createFn func(ctx context.Context, sub Subscription) (Subscription, error)
	getFn    func(ctx context.Context, id uuid.UUID) (Subscription, error)
	listFn   func(ctx context.Context, filter ListFilter) ([]Subscription, error)
	updateFn func(ctx context.Context, sub Subscription) (Subscription, error)
	deleteFn func(ctx context.Context, id uuid.UUID) error
	totalFn  func(ctx context.Context, filter TotalFilter) (int, error)
}

func (s stubRepository) Create(ctx context.Context, sub Subscription) (Subscription, error) {
	return s.createFn(ctx, sub)
}

func (s stubRepository) Get(ctx context.Context, id uuid.UUID) (Subscription, error) {
	return s.getFn(ctx, id)
}

func (s stubRepository) List(ctx context.Context, filter ListFilter) ([]Subscription, error) {
	return s.listFn(ctx, filter)
}

func (s stubRepository) Update(ctx context.Context, sub Subscription) (Subscription, error) {
	return s.updateFn(ctx, sub)
}

func (s stubRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return s.deleteFn(ctx, id)
}

func (s stubRepository) Total(ctx context.Context, filter TotalFilter) (int, error) {
	return s.totalFn(ctx, filter)
}

func TestService_CreateValidatesPrice(t *testing.T) {
	svc := NewService(stubRepository{})

	_, err := svc.Create(context.Background(), CreateRequest{
		ServiceName: "Yandex Plus",
		Price:       0,
		UserID:      uuid.New().String(),
		StartDate:   "07-2025",
	})

	if !errors.Is(err, ErrInvalidPrice) {
		t.Fatalf("Create() error = %v, want %v", err, ErrInvalidPrice)
	}
}

func TestService_CreateValidatesEndDateAfterStartDate(t *testing.T) {
	svc := NewService(stubRepository{})
	endDate := "06-2025"

	_, err := svc.Create(context.Background(), CreateRequest{
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      uuid.New().String(),
		StartDate:   "07-2025",
		EndDate:     &endDate,
	})

	if !errors.Is(err, ErrInvalidPeriod) {
		t.Fatalf("Create() error = %v, want %v", err, ErrInvalidPeriod)
	}
}

func TestService_CreateValidatesUserID(t *testing.T) {
	svc := NewService(stubRepository{})

	_, err := svc.Create(context.Background(), CreateRequest{
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      "not-a-uuid",
		StartDate:   "07-2025",
	})

	if !errors.Is(err, ErrInvalidUUID) {
		t.Fatalf("Create() error = %v, want %v", err, ErrInvalidUUID)
	}
}

func TestService_CreatePassesSubscriptionToRepository(t *testing.T) {
	userID := uuid.New()
	endDate := "09-2025"
	wantStart := time.Date(2025, time.July, 1, 0, 0, 0, 0, time.UTC)
	wantEnd := time.Date(2025, time.September, 1, 0, 0, 0, 0, time.UTC)

	repo := stubRepository{
		createFn: func(ctx context.Context, sub Subscription) (Subscription, error) {
			if ctx == nil {
				t.Fatal("ctx is nil")
			}
			if sub.ServiceName != "Yandex Plus" {
				t.Fatalf("ServiceName = %q", sub.ServiceName)
			}
			if sub.Price != 400 {
				t.Fatalf("Price = %d", sub.Price)
			}
			if sub.UserID != userID {
				t.Fatalf("UserID = %s, want %s", sub.UserID, userID)
			}
			if !sub.StartDate.Equal(wantStart) {
				t.Fatalf("StartDate = %s, want %s", sub.StartDate, wantStart)
			}
			if sub.EndDate == nil || !sub.EndDate.Equal(wantEnd) {
				t.Fatalf("EndDate = %v, want %s", sub.EndDate, wantEnd)
			}

			sub.ID = uuid.New()
			return sub, nil
		},
	}
	svc := NewService(repo)

	got, err := svc.Create(context.Background(), CreateRequest{
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      userID.String(),
		StartDate:   "07-2025",
		EndDate:     &endDate,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if got.StartDate != "07-2025" {
		t.Fatalf("response StartDate = %q, want %q", got.StartDate, "07-2025")
	}
	if got.EndDate == nil || *got.EndDate != "09-2025" {
		t.Fatalf("response EndDate = %v, want %q", got.EndDate, "09-2025")
	}
}

func TestService_TotalValidatesRequestedPeriod(t *testing.T) {
	svc := NewService(stubRepository{})

	_, err := svc.Total(context.Background(), TotalRequest{
		From: "12-2025",
		To:   "07-2025",
	})

	if !errors.Is(err, ErrInvalidPeriod) {
		t.Fatalf("Total() error = %v, want %v", err, ErrInvalidPeriod)
	}
}

func TestService_TotalPassesFilterToRepository(t *testing.T) {
	userID := uuid.New()
	serviceName := "Yandex Plus"
	wantFrom := time.Date(2025, time.July, 1, 0, 0, 0, 0, time.UTC)
	wantTo := time.Date(2025, time.December, 1, 0, 0, 0, 0, time.UTC)

	repo := stubRepository{
		totalFn: func(ctx context.Context, filter TotalFilter) (int, error) {
			if ctx == nil {
				t.Fatal("ctx is nil")
			}
			if !filter.From.Equal(wantFrom) {
				t.Fatalf("From = %s, want %s", filter.From, wantFrom)
			}
			if !filter.To.Equal(wantTo) {
				t.Fatalf("To = %s, want %s", filter.To, wantTo)
			}
			if filter.UserID == nil || *filter.UserID != userID {
				t.Fatalf("UserID = %v, want %s", filter.UserID, userID)
			}
			if filter.ServiceName == nil || *filter.ServiceName != serviceName {
				t.Fatalf("ServiceName = %v, want %s", filter.ServiceName, serviceName)
			}
			return 1200, nil
		},
	}
	svc := NewService(repo)

	got, err := svc.Total(context.Background(), TotalRequest{
		From:        "07-2025",
		To:          "12-2025",
		UserID:      userID.String(),
		ServiceName: serviceName,
	})
	if err != nil {
		t.Fatalf("Total() error = %v", err)
	}
	if got.Total != 1200 {
		t.Fatalf("Total = %d, want %d", got.Total, 1200)
	}
}

func TestService_TotalValidatesUserIDFilter(t *testing.T) {
	svc := NewService(stubRepository{})

	_, err := svc.Total(context.Background(), TotalRequest{
		From:   "07-2025",
		To:     "12-2025",
		UserID: "not-a-uuid",
	})

	if !errors.Is(err, ErrInvalidUUID) {
		t.Fatalf("Total() error = %v, want %v", err, ErrInvalidUUID)
	}
}
