package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"subscriptions-service/internal/domain"
	"subscriptions-service/internal/lib/logger"
	"subscriptions-service/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSubscriptionService_Create_SetsIDAndPersists(t *testing.T) {
	t.Parallel()

	repo := mocks.NewMockSubscriptionRepository(t)
	svc := NewSubscriptionService(repo, newTestLogger())
	ctx := context.Background()

	sub := &domain.Subscription{
		ServiceName: "Netflix",
		Price:       999,
		UserID:      uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		StartDate:   time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC),
	}

	repo.EXPECT().
		Create(mock.Anything, mock.MatchedBy(func(s *domain.Subscription) bool {
			return s != nil &&
				s.ServiceName == sub.ServiceName &&
				s.Price == sub.Price &&
				s.UserID == sub.UserID &&
				s.StartDate.Equal(sub.StartDate) &&
				s.ID != uuid.Nil
		})).
		Return(nil).
		Once()

	got, err := svc.Create(ctx, sub)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.NotEqual(t, uuid.Nil, got.ID)
}

func TestSubscriptionService_Create_ReturnsWrappedError(t *testing.T) {
	t.Parallel()

	repo := mocks.NewMockSubscriptionRepository(t)
	svc := NewSubscriptionService(repo, newTestLogger())
	ctx := context.Background()

	sub := &domain.Subscription{
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba"),
		StartDate:   time.Date(2025, time.July, 1, 0, 0, 0, 0, time.UTC),
	}

	repoErr := errors.New("insert failed")
	repo.EXPECT().Create(mock.Anything, mock.Anything).Return(repoErr).Once()

	_, err := svc.Create(ctx, sub)
	require.Error(t, err)
	require.ErrorIs(t, err, repoErr)
	require.Contains(t, err.Error(), "SubscriptionService.Create")
}

func TestSubscriptionService_GetByID_Success(t *testing.T) {
	t.Parallel()

	repo := mocks.NewMockSubscriptionRepository(t)
	svc := NewSubscriptionService(repo, newTestLogger())
	ctx := context.Background()
	id := uuid.New()
	want := &domain.Subscription{ID: id, ServiceName: "Spotify", Price: 499}

	repo.EXPECT().GetByID(mock.Anything, id).Return(want, nil).Once()

	got, err := svc.GetByID(ctx, id)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestSubscriptionService_GetByID_NotFound(t *testing.T) {
	t.Parallel()

	repo := mocks.NewMockSubscriptionRepository(t)
	svc := NewSubscriptionService(repo, newTestLogger())
	ctx := context.Background()
	id := uuid.New()

	repo.EXPECT().GetByID(mock.Anything, id).Return((*domain.Subscription)(nil), ErrNotFound).Once()

	got, err := svc.GetByID(ctx, id)
	require.Nil(t, got)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrNotFound)
	require.Contains(t, err.Error(), "SubscriptionService.GetByID")
}

func TestSubscriptionService_GetByID_RepositoryError(t *testing.T) {
	t.Parallel()

	repo := mocks.NewMockSubscriptionRepository(t)
	svc := NewSubscriptionService(repo, newTestLogger())
	ctx := context.Background()
	id := uuid.New()
	repoErr := errors.New("select failed")

	repo.EXPECT().GetByID(mock.Anything, id).Return((*domain.Subscription)(nil), repoErr).Once()

	got, err := svc.GetByID(ctx, id)
	require.Nil(t, got)
	require.Error(t, err)
	require.ErrorIs(t, err, repoErr)
	require.Contains(t, err.Error(), "SubscriptionService.GetByID")
}

func TestSubscriptionService_List_Update_Delete_Sum(t *testing.T) {
	t.Parallel()

	repo := mocks.NewMockSubscriptionRepository(t)
	svc := NewSubscriptionService(repo, newTestLogger())
	ctx := context.Background()

	filter := domain.SubscriptionFilter{}
	list := []domain.Subscription{{ID: uuid.New(), ServiceName: "A", Price: 100}}
	repo.EXPECT().List(mock.Anything, filter).Return(list, nil).Once()

	gotList, err := svc.List(ctx, filter)
	require.NoError(t, err)
	require.Equal(t, list, gotList)

	sub := &domain.Subscription{ID: uuid.New(), ServiceName: "B", Price: 200}
	repo.EXPECT().Update(mock.Anything, sub).Return(nil).Once()

	gotSub, err := svc.Update(ctx, sub)
	require.NoError(t, err)
	require.Equal(t, sub, gotSub)

	repo.EXPECT().Delete(mock.Anything, sub.ID).Return(nil).Once()
	require.NoError(t, svc.Delete(ctx, sub.ID))

	repo.EXPECT().Sum(mock.Anything, filter).Return(300, nil).Once()
	total, err := svc.Sum(ctx, filter)
	require.NoError(t, err)
	require.Equal(t, 300, total)
}

func TestSubscriptionService_ErrorsAreWrapped(t *testing.T) {
	t.Parallel()

	repo := mocks.NewMockSubscriptionRepository(t)
	svc := NewSubscriptionService(repo, newTestLogger())
	ctx := context.Background()
	filter := domain.SubscriptionFilter{}
	sub := &domain.Subscription{ID: uuid.New()}
	baseErr := errors.New("boom")

	repo.EXPECT().List(mock.Anything, filter).Return(nil, baseErr).Once()
	_, err := svc.List(ctx, filter)
	require.Error(t, err)
	require.ErrorIs(t, err, baseErr)
	require.True(t, strings.Contains(err.Error(), "SubscriptionService.List"))

	repo.EXPECT().Update(mock.Anything, sub).Return(baseErr).Once()
	_, err = svc.Update(ctx, sub)
	require.Error(t, err)
	require.ErrorIs(t, err, baseErr)
	require.True(t, strings.Contains(err.Error(), "SubscriptionService.Update"))

	repo.EXPECT().Delete(mock.Anything, sub.ID).Return(baseErr).Once()
	err = svc.Delete(ctx, sub.ID)
	require.Error(t, err)
	require.ErrorIs(t, err, baseErr)
	require.True(t, strings.Contains(err.Error(), "SubscriptionService.Delete"))

	repo.EXPECT().Sum(mock.Anything, filter).Return(0, baseErr).Once()
	_, err = svc.Sum(ctx, filter)
	require.Error(t, err)
	require.ErrorIs(t, err, baseErr)
	require.True(t, strings.Contains(err.Error(), "SubscriptionService.Sum"))
}

func newTestLogger() logger.Logger {
	return logger.NewStdLogger(slog.New(slog.NewTextHandler(io.Discard, nil)))
}
