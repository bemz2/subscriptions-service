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
	query, args, err := sq.
		Insert("subscriptions").
		Columns("id", "service_name", "price", "user_id", "start_date", "end_date").
		Values(s.ID, s.ServiceName, s.Price, s.UserID, s.StartDate, s.EndDate).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	return nil
}

func (r *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error) {
	query, args, err := sq.
		Select("id", "service_name", "price", "user_id", "start_date", "end_date").
		From("subscriptions").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
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
		return nil, ErrSubscriptionNotFound
	}
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (r *SubscriptionRepository) List(ctx context.Context, filter domain.SubscriptionFilter) ([]domain.Subscription, error) {
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
		return nil, fmt.Errorf("build query: %w", err)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
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
			return nil, fmt.Errorf("scan: %w", err)
		}
		result = append(result, s)
	}

	return result, rows.Err()
}

func (r *SubscriptionRepository) Update(ctx context.Context, s *domain.Subscription) error {
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
		return fmt.Errorf("build query: %w", err)
	}

	res, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	if res.RowsAffected() == 0 {
		return ErrSubscriptionNotFound
	}

	return nil
}

func (r *SubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query, args, err := sq.
		Delete("subscriptions").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}

	res, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	if res.RowsAffected() == 0 {
		return ErrSubscriptionNotFound
	}

	return nil
}

func (r *SubscriptionRepository) Sum(ctx context.Context, filter domain.SubscriptionFilter) (int, error) {
	builder := sq.
		Select("COALESCE(SUM(price), 0)").
		From("subscriptions").
		PlaceholderFormat(sq.Dollar)

	if filter.UserID != nil {
		builder = builder.Where(sq.Eq{"user_id": *filter.UserID})
	}
	if filter.ServiceName != nil {
		builder = builder.Where(sq.Eq{"service_name": *filter.ServiceName})
	}
	if filter.To != nil {
		builder = builder.Where(sq.LtOrEq{"start_date": *filter.To})
	}
	if filter.From != nil {
		builder = builder.Where(
			sq.Or{
				sq.Eq{"end_date": nil},
				sq.GtOrEq{"end_date": *filter.From},
			},
		)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return 0, fmt.Errorf("build query: %w", err)
	}

	var total int
	err = r.pool.QueryRow(ctx, query, args...).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("query: %w", err)
	}

	return total, nil
}
