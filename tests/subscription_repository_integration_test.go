package tests

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"subscriptions-service/internal/domain"
	"subscriptions-service/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestSubscriptionRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test is skipped in short mode")
	}
	if os.Getenv("RUN_INTEGRATION_TESTS") != "1" {
		t.Skip("set RUN_INTEGRATION_TESTS=1 to run integration tests")
	}

	ctx := context.Background()
	repo := newIntegrationRepo(t, ctx)

	userA := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	userB := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	sub1 := &domain.Subscription{
		ID:          uuid.New(),
		ServiceName: "Netflix",
		Price:       100,
		UserID:      userA,
		StartDate:   testDate(2025, time.January),
	}
	sub2End := testDate(2025, time.May)
	sub2 := &domain.Subscription{
		ID:          uuid.New(),
		ServiceName: "Spotify",
		Price:       200,
		UserID:      userA,
		StartDate:   testDate(2025, time.February),
		EndDate:     &sub2End,
	}
	sub3 := &domain.Subscription{
		ID:          uuid.New(),
		ServiceName: "Netflix",
		Price:       300,
		UserID:      userB,
		StartDate:   testDate(2025, time.March),
	}

	require.NoError(t, repo.Create(ctx, sub1))
	require.NoError(t, repo.Create(ctx, sub2))
	require.NoError(t, repo.Create(ctx, sub3))

	t.Run("get by id", func(t *testing.T) {
		got, err := repo.GetByID(ctx, sub1.ID)
		require.NoError(t, err)
		require.Equal(t, sub1.ID, got.ID)
		require.Equal(t, "Netflix", got.ServiceName)
		require.Equal(t, 100, got.Price)
		require.Equal(t, userA, got.UserID)
	})

	t.Run("get by id not found", func(t *testing.T) {
		_, err := repo.GetByID(ctx, uuid.New())
		require.Error(t, err)
		require.ErrorIs(t, err, repository.ErrSubscriptionNotFound)
	})

	t.Run("list filters", func(t *testing.T) {
		byUser, err := repo.List(ctx, domain.SubscriptionFilter{UserID: &userA})
		require.NoError(t, err)
		require.Len(t, byUser, 2)

		service := "Netflix"
		byUserAndService, err := repo.List(ctx, domain.SubscriptionFilter{
			UserID:      &userA,
			ServiceName: &service,
		})
		require.NoError(t, err)
		require.Len(t, byUserAndService, 1)
		require.Equal(t, sub1.ID, byUserAndService[0].ID)

		from := testDate(2025, time.February)
		to := testDate(2025, time.February)
		byMonth, err := repo.List(ctx, domain.SubscriptionFilter{From: &from, To: &to})
		require.NoError(t, err)
		require.Len(t, byMonth, 1)
		require.Equal(t, sub2.ID, byMonth[0].ID)

		service = "NotExistingService"
		empty, err := repo.List(ctx, domain.SubscriptionFilter{ServiceName: &service})
		require.NoError(t, err)
		require.Empty(t, empty)
	})

	t.Run("sum filters", func(t *testing.T) {
		from := testDate(2025, time.March)
		to := testDate(2025, time.April)

		total, err := repo.Sum(ctx, domain.SubscriptionFilter{
			UserID: &userA,
			From:   &from,
			To:     &to,
		})
		require.NoError(t, err)
		require.Equal(t, 300, total)

		service := "Netflix"
		total, err = repo.Sum(ctx, domain.SubscriptionFilter{ServiceName: &service})
		require.NoError(t, err)
		require.Equal(t, 400, total)

		unknownUser := uuid.MustParse("33333333-3333-3333-3333-333333333333")
		total, err = repo.Sum(ctx, domain.SubscriptionFilter{UserID: &unknownUser})
		require.NoError(t, err)
		require.Zero(t, total)
	})

	t.Run("update", func(t *testing.T) {
		sub2.ServiceName = "Spotify Premium"
		sub2.Price = 250
		require.NoError(t, repo.Update(ctx, sub2))

		got, err := repo.GetByID(ctx, sub2.ID)
		require.NoError(t, err)
		require.Equal(t, "Spotify Premium", got.ServiceName)
		require.Equal(t, 250, got.Price)
	})

	t.Run("update not found", func(t *testing.T) {
		missing := &domain.Subscription{
			ID:          uuid.New(),
			ServiceName: "Missing",
			Price:       100,
			StartDate:   testDate(2025, time.January),
		}
		err := repo.Update(ctx, missing)
		require.Error(t, err)
		require.ErrorIs(t, err, repository.ErrSubscriptionNotFound)
	})

	t.Run("delete and not found", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, sub3.ID))

		_, err := repo.GetByID(ctx, sub3.ID)
		require.Error(t, err)
		require.ErrorIs(t, err, repository.ErrSubscriptionNotFound)

		err = repo.Delete(ctx, uuid.New())
		require.Error(t, err)
		require.ErrorIs(t, err, repository.ErrSubscriptionNotFound)
	})
}

func newIntegrationRepo(t *testing.T, ctx context.Context) *repository.SubscriptionRepository {
	t.Helper()

	pg, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("subscription_test"),
		postgres.WithUsername("subscription"),
		postgres.WithPassword("subscription"),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, pg.Terminate(ctx))
	})

	connString, err := pg.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connString)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	waitForDB(t, ctx, pool)

	_, err = pool.Exec(ctx, migrationSQL(t))
	require.NoError(t, err)

	return repository.NewSubscriptionRepository(pool)
}

func migrationSQL(t *testing.T) string {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	require.True(t, ok)

	sqlPath := filepath.Join(filepath.Dir(currentFile), "..", "migrations", "0001_create_initial_tables.up.sql")
	sqlBytes, err := os.ReadFile(sqlPath)
	require.NoError(t, err)

	return string(sqlBytes)
}

func testDate(year int, month time.Month) time.Time {
	return time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
}

func waitForDB(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()

	var err error
	for i := 0; i < 30; i++ {
		err = pool.Ping(ctx)
		if err == nil {
			return
		}
		time.Sleep(500 * time.Millisecond)
	}

	require.NoError(t, err)
}
