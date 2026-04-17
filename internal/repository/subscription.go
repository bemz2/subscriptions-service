package repository

import (
	"context"
	"errors"
	"fmt"
	"subscriptions-service/internal/domain"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrSubscriptionNotFound = errors.New("subscription not found")

type SubscriptionRepository struct {
	pool *pgxpool.Pool
}

func NewSubscriptionRepository(pool *pgxpool.Pool) *SubscriptionRepository {
	return &SubscriptionRepository{pool: pool}
}

func (r *SubscriptionRepository) Create(ctx context.Context, s *domain.Subscription) error {
	const contextKey = "SubscriptionRepository.Create"

	query, args, err := sq.
		Insert("subscriptions").
		Columns("id", "service_name", "price", "user_id", "start_date", "end_date").
		Values(s.ID, s.ServiceName, s.Price, s.UserID, s.StartDate, s.EndDate).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("%s: build query: %w", contextKey, err)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: exec: %w", contextKey, err)
	}

	return nil
}

func (r *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error) {
	const contextKey = "SubscriptionRepository.GetByID"

	query, args, err := sq.
		Select("id", "service_name", "price", "user_id", "start_date", "end_date").
		From("subscriptions").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: build query: %w", contextKey, err)
	}

	var s domain.Subscription

	err = r.pool.QueryRow(ctx, query, args...).Scan(
		&s.ID,
		&s.ServiceName,
		&s.Price,
		&s.UserID,
		&s.StartDate,
		&s.EndDate,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("%s: %w", contextKey, ErrSubscriptionNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: query row: %w", contextKey, err)
	}

	return &s, nil
}

func (r *SubscriptionRepository) List(ctx context.Context, filter domain.SubscriptionFilter) ([]domain.Subscription, error) {
	const contextKey = "SubscriptionRepository.List"

	builder := sq.
		Select("id", "service_name", "price", "user_id", "start_date", "end_date").
		From("subscriptions").
		PlaceholderFormat(sq.Dollar)

	if filter.UserID != nil {
		builder = builder.Where(sq.Eq{"user_id": *filter.UserID})
	}
	if filter.ServiceName != nil {
		builder = builder.Where(sq.Eq{"service_name": *filter.ServiceName})
	}
	if filter.From != nil {
		builder = builder.Where(sq.GtOrEq{"start_date": *filter.From})
	}
	if filter.To != nil {
		builder = builder.Where(sq.LtOrEq{"start_date": *filter.To})
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: build query: %w", contextKey, err)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: query: %w", contextKey, err)
	}
	defer rows.Close()

	var result []domain.Subscription

	for rows.Next() {
		var s domain.Subscription
		if err := rows.Scan(
			&s.ID,
			&s.ServiceName,
			&s.Price,
			&s.UserID,
			&s.StartDate,
			&s.EndDate,
		); err != nil {
			return nil, fmt.Errorf("%s: scan: %w", contextKey, err)
		}
		result = append(result, s)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows: %w", contextKey, err)
	}

	return result, nil
}

func (r *SubscriptionRepository) Update(ctx context.Context, s *domain.Subscription) error {
	const contextKey = "SubscriptionRepository.Update"

	query, args, err := sq.
		Update("subscriptions").
		Set("service_name", s.ServiceName).
		Set("price", s.Price).
		Set("start_date", s.StartDate).
		Set("end_date", s.EndDate).
		Where(sq.Eq{"id": s.ID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("%s: build query: %w", contextKey, err)
	}

	res, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: exec: %w", contextKey, err)
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", contextKey, ErrSubscriptionNotFound)
	}

	return nil
}

func (r *SubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const contextKey = "SubscriptionRepository.Delete"

	query, args, err := sq.
		Delete("subscriptions").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("%s: build query: %w", contextKey, err)
	}

	res, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: exec: %w", contextKey, err)
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", contextKey, ErrSubscriptionNotFound)
	}

	return nil
}

func (r *SubscriptionRepository) ListForSum(ctx context.Context, filter domain.SubscriptionFilter) ([]domain.Subscription, error) {
	const contextKey = "SubscriptionRepository.ListForSum"
	if filter.From == nil || filter.To == nil {
		return nil, fmt.Errorf("%s: from and to are required", contextKey)
	}
	if filter.From.After(*filter.To) {
		return nil, fmt.Errorf("%s: from must be before or equal to to", contextKey)
	}

	builder := sq.
		Select("id", "service_name", "price", "user_id", "start_date", "end_date").
		From("subscriptions").
		Where(sq.LtOrEq{"start_date": *filter.To}).
		Where(
			sq.Or{
				sq.Eq{"end_date": nil},
				sq.GtOrEq{"end_date": *filter.From},
			},
		).
		PlaceholderFormat(sq.Dollar)

	if filter.UserID != nil {
		builder = builder.Where(sq.Eq{"user_id": *filter.UserID})
	}
	if filter.ServiceName != nil {
		builder = builder.Where(sq.Eq{"service_name": *filter.ServiceName})
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: build query: %w", contextKey, err)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: query: %w", contextKey, err)
	}
	defer rows.Close()

	result := make([]domain.Subscription, 0)
	for rows.Next() {
		var sub domain.Subscription

		if err = rows.Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &sub.StartDate, &sub.EndDate); err != nil {
			return nil, fmt.Errorf("%s: scan: %w", contextKey, err)
		}
		result = append(result, sub)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows: %w", contextKey, err)
	}

	return result, nil
}
