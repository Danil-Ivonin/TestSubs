package subscription

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, sub Subscription) (Subscription, error)
	Get(ctx context.Context, id uuid.UUID) (Subscription, error)
	List(ctx context.Context, filter ListFilter) ([]Subscription, error)
	Update(ctx context.Context, sub Subscription) (Subscription, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Total(ctx context.Context, filter TotalFilter) (int, error)
}
