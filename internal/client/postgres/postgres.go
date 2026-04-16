package postgres

import (
	"context"
	"fmt"
	"subscriptions-service/internal"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool creates a pgx connection pool using the provided DSN.
func NewPool(ctx context.Context, cfg internal.PostgresConfig) (*pgxpool.Pool, error) {
	cfgPool, err := pgxpool.ParseConfig(createDSN(cfg))
	if err != nil {
		return nil, fmt.Errorf("parse pool config: %w", err)
	}
	cfgPool.MaxConns = 10
	cfgPool.MinConns = 2
	cfgPool.MaxConnLifetime = time.Hour
	cfgPool.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfgPool)
	if err != nil {
		return nil, fmt.Errorf("create pgx pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping pgx pool: %w", err)
	}
	return pool, nil
}

func createDSN(config internal.PostgresConfig) string {
	dsn := fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=%s",
		config.DBAdapter,
		config.DBUser,
		config.DBPassword,
		config.DBHost,
		config.DBPort,
		config.DBName,
		config.DBSSLMode,
	)
	return dsn
}
