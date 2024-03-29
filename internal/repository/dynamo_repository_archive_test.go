package repository

import (
	"context"
	"testing"

	"github.com/hareku/habit-tracker-app/internal/apperrors"
	"github.com/hareku/habit-tracker-app/internal/auth"
	"github.com/stretchr/testify/require"
)

func TestDynamoRepository_AllArchivedHabits(t *testing.T) {
	repo := newDynamoRepositoryTest(t)
	ctx := context.Background()
	myUserID := auth.UserID("MyUserID")

	h1, err := repo.CreateHabit(ctx, myUserID, "Habit1")
	require.NoError(t, err)
	require.NoError(t, repo.ArchiveHabit(ctx, myUserID, h1.ID))
	h2, err := repo.CreateHabit(ctx, myUserID, "Habit2")
	require.NoError(t, err)
	require.NoError(t, repo.ArchiveHabit(ctx, myUserID, h2.ID))

	got, err := repo.AllArchivedHabits(ctx, myUserID)
	require.NoError(t, err)
	require.Len(t, got, 2)
}

func TestDynamoRepository_ArchiveHabit(t *testing.T) {
	repo := newDynamoRepositoryTest(t)
	ctx := context.Background()
	myUserID := auth.UserID("MyUserID")

	h1, err := repo.CreateHabit(ctx, myUserID, "Habit1")
	require.NoError(t, err)

	require.NoError(t, repo.ArchiveHabit(ctx, myUserID, h1.ID))

	_, err = repo.FindHabit(ctx, myUserID, h1.ID)
	require.ErrorIs(t, err, apperrors.ErrNotFound)

	archivedH1, err := repo.FindArchivedHabit(ctx, myUserID, h1.ID)
	require.NoError(t, err)
	require.Equal(t, h1.Title, archivedH1.Title)
}

func TestDynamoRepository_UnarchiveHabit(t *testing.T) {
	repo := newDynamoRepositoryTest(t)
	ctx := context.Background()
	myUserID := auth.UserID("MyUserID")

	h1, err := repo.CreateHabit(ctx, myUserID, "Habit1")
	require.NoError(t, err)

	require.NoError(t, repo.ArchiveHabit(ctx, myUserID, h1.ID))
	require.NoError(t, repo.UnarchiveHabit(ctx, myUserID, h1.ID))

	h2, err := repo.FindHabit(ctx, myUserID, h1.ID)
	require.NoError(t, err)

	require.Equal(t, h1, h2)
}
