package subscription

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(ctx context.Context, sub Subscription) (Subscription, error) {
	const query = `
		INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, service_name, price, user_id, start_date, end_date, created_at, updated_at
	`

	created, err := r.scanSubscription(r.pool.QueryRow(ctx, query,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		sub.StartDate,
		sub.EndDate,
	))
	if err != nil {
		return Subscription{}, fmt.Errorf("inserting subscription: %w", err)
	}

	return created, nil
}

func (r *PostgresRepository) Get(ctx context.Context, id uuid.UUID) (Subscription, error) {
	const query = `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
		WHERE id = $1
	`

	sub, err := r.scanSubscription(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Subscription{}, ErrNotFound
		}
		return Subscription{}, fmt.Errorf("selecting subscription: %w", err)
	}

	return sub, nil
}

func (r *PostgresRepository) List(ctx context.Context, filter ListFilter) ([]Subscription, error) {
	query := strings.Builder{}
	query.WriteString(`
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
		WHERE 1 = 1
	`)

	args := make([]any, 0, 4)
	argPos := 1

	if filter.UserID != nil {
		query.WriteString(fmt.Sprintf(" AND user_id = $%d", argPos))
		args = append(args, *filter.UserID)
		argPos++
	}
	if filter.ServiceName != nil {
		query.WriteString(fmt.Sprintf(" AND service_name = $%d", argPos))
		args = append(args, *filter.ServiceName)
		argPos++
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query.WriteString(fmt.Sprintf(" ORDER BY created_at DESC, id DESC LIMIT $%d OFFSET $%d", argPos, argPos+1))
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("querying subscriptions: %w", err)
	}
	defer rows.Close()

	subs := make([]Subscription, 0)
	for rows.Next() {
		sub, err := scanSubscriptionRows(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning subscription: %w", err)
		}
		subs = append(subs, sub)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating subscriptions: %w", err)
	}

	return subs, nil
}

func (r *PostgresRepository) Update(ctx context.Context, sub Subscription) (Subscription, error) {
	const query = `
		UPDATE subscriptions
		SET service_name = $2,
		    price = $3,
		    user_id = $4,
		    start_date = $5,
		    end_date = $6,
		    updated_at = now()
		WHERE id = $1
		RETURNING id, service_name, price, user_id, start_date, end_date, created_at, updated_at
	`

	updated, err := r.scanSubscription(r.pool.QueryRow(ctx, query,
		sub.ID,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		sub.StartDate,
		sub.EndDate,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Subscription{}, ErrNotFound
		}
		return Subscription{}, fmt.Errorf("updating subscription: %w", err)
	}

	return updated, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const query = `DELETE FROM subscriptions WHERE id = $1`

	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("deleting subscription: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *PostgresRepository) Total(ctx context.Context, filter TotalFilter) (int, error) {
	query := strings.Builder{}
	query.WriteString(`
		SELECT COALESCE(SUM(
			(
				(
					EXTRACT(YEAR FROM LEAST(COALESCE(end_date, $2), $2))::int
					- EXTRACT(YEAR FROM GREATEST(start_date, $1))::int
				) * 12
				+ (
					EXTRACT(MONTH FROM LEAST(COALESCE(end_date, $2), $2))::int
					- EXTRACT(MONTH FROM GREATEST(start_date, $1))::int
				)
				+ 1
			) * price
		), 0)::int
		FROM subscriptions
		WHERE start_date <= $2
		  AND COALESCE(end_date, $2) >= $1
	`)

	args := []any{filter.From, filter.To}
	argPos := 3

	if filter.UserID != nil {
		query.WriteString(fmt.Sprintf(" AND user_id = $%d", argPos))
		args = append(args, *filter.UserID)
		argPos++
	}
	if filter.ServiceName != nil {
		query.WriteString(fmt.Sprintf(" AND service_name = $%d", argPos))
		args = append(args, *filter.ServiceName)
	}

	var total int
	if err := r.pool.QueryRow(ctx, query.String(), args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("calculating total: %w", err)
	}

	return total, nil
}

func (r *PostgresRepository) scanSubscription(row pgx.Row) (Subscription, error) {
	return scanSubscription(row)
}

type subscriptionScanner interface {
	Scan(dest ...any) error
}

func scanSubscription(row subscriptionScanner) (Subscription, error) {
	var sub Subscription
	var endDate *time.Time

	err := row.Scan(
		&sub.ID,
		&sub.ServiceName,
		&sub.Price,
		&sub.UserID,
		&sub.StartDate,
		&endDate,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)
	if err != nil {
		return Subscription{}, err
	}
	sub.EndDate = endDate

	return sub, nil
}

func scanSubscriptionRows(rows pgx.Rows) (Subscription, error) {
	return scanSubscription(rows)
}
