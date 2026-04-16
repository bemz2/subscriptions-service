package service

import (
	"context"
	"errors"
	"fmt"
	"subscriptions-service/internal/domain"
	"subscriptions-service/internal/lib/logger"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("subscription not found")

type SubscriptionRepository interface {
	Create(ctx context.Context, s *domain.Subscription) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error)
	List(ctx context.Context, filter domain.SubscriptionFilter) ([]domain.Subscription, error)
	Update(ctx context.Context, s *domain.Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
	Sum(ctx context.Context, filter domain.SubscriptionFilter) (int, error)
}

type SubscriptionService struct {
	repo SubscriptionRepository
	log  logger.Logger
}

func NewSubscriptionService(repo SubscriptionRepository, log logger.Logger) *SubscriptionService {
	return &SubscriptionService{repo: repo, log: log}
}

func (s *SubscriptionService) Create(ctx context.Context, sub *domain.Subscription) (*domain.Subscription, error) {
	const contextKey = "SubscriptionService.Create"

	sub.ID = uuid.New()

	if err := s.repo.Create(ctx, sub); err != nil {
		s.log.ErrorContext(ctx, "create subscription", "error", err)
		return nil, fmt.Errorf("%s: %w", contextKey, err)
	}

	s.log.InfoContext(ctx, "subscription created", "id", sub.ID)
	return sub, nil
}

func (s *SubscriptionService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error) {
	const contextKey = "SubscriptionService.GetByID"

	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, fmt.Errorf("%s: %w", contextKey, ErrNotFound)
		}
		s.log.ErrorContext(ctx, "get subscription", "id", id, "error", err)
		return nil, fmt.Errorf("%s: %w", contextKey, err)
	}

	return sub, nil
}

func (s *SubscriptionService) List(ctx context.Context, filter domain.SubscriptionFilter) ([]domain.Subscription, error) {
	const contextKey = "SubscriptionService.List"

	subs, err := s.repo.List(ctx, filter)
	if err != nil {
		s.log.ErrorContext(ctx, "list subscriptions", "error", err)
		return nil, fmt.Errorf("%s: %w", contextKey, err)
	}

	return subs, nil
}

func (s *SubscriptionService) Update(ctx context.Context, sub *domain.Subscription) (*domain.Subscription, error) {
	const contextKey = "SubscriptionService.Update"

	if err := s.repo.Update(ctx, sub); err != nil {
		s.log.ErrorContext(ctx, "update subscription", "id", sub.ID, "error", err)
		return nil, fmt.Errorf("%s: %w", contextKey, err)
	}

	s.log.InfoContext(ctx, "subscription updated", "id", sub.ID)
	return sub, nil
}

func (s *SubscriptionService) Delete(ctx context.Context, id uuid.UUID) error {
	const contextKey = "SubscriptionService.Delete"

	if err := s.repo.Delete(ctx, id); err != nil {
		s.log.ErrorContext(ctx, "delete subscription", "id", id, "error", err)
		return fmt.Errorf("%s: %w", contextKey, err)
	}

	s.log.InfoContext(ctx, "subscription deleted", "id", id)
	return nil
}

func (s *SubscriptionService) Sum(ctx context.Context, filter domain.SubscriptionFilter) (int, error) {
	const contextKey = "SubscriptionService.Sum"

	total, err := s.repo.Sum(ctx, filter)
	if err != nil {
		s.log.ErrorContext(ctx, "sum subscriptions", "error", err)
		return 0, fmt.Errorf("%s: %w", contextKey, err)
	}

	return total, nil
}
