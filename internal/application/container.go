package application

import (
	"context"
	"log/slog"
	"os"
	"subscriptions-service/internal"
	"subscriptions-service/internal/client/postgres"
	"subscriptions-service/internal/http/handler"
	"subscriptions-service/internal/lib/logger"
	"subscriptions-service/internal/repository"
	"subscriptions-service/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Container struct {
	Ctx    context.Context
	Config *internal.AppConfig
	Logger logger.Logger

	Pool *pgxpool.Pool

	SubscriptionRepo    *repository.SubscriptionRepository
	SubscriptionService *service.SubscriptionService
	SubscriptionHandler *handler.SubscriptionHandler
}

func NewContainer(ctx context.Context, config internal.AppConfig) *Container {
	return &Container{
		Ctx:    ctx,
		Config: &config,
	}
}

func (c *Container) Init(ctx context.Context) (*Container, error) {
	slogLogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	c.Logger = logger.NewStdLogger(slogLogger)

	pool, err := postgres.NewPool(ctx, c.Config.PostgresConfig)
	if err != nil {
		return c, err
	}
	c.Pool = pool

	c.SubscriptionRepo = repository.NewSubscriptionRepository(c.Pool)
	c.SubscriptionService = service.NewSubscriptionService(c.SubscriptionRepo, c.Logger)
	c.SubscriptionHandler = handler.NewSubscriptionHandler(c.SubscriptionService)

	return c, nil
}

func (c *Container) Close() {
	if c.Pool != nil {
		c.Pool.Close()
	}
}
