package api

import (
	"context"

	"github.com/google/uuid"
	"github.com/hareku/habit-tracker-app/internal/auth"
	"github.com/hareku/habit-tracker-app/internal/repository"
)

type DynamoRepository interface {
	AllArchivedHabits(ctx context.Context, uid auth.UserID) ([]*repository.DynamoHabit, error)
	AllHabits(ctx context.Context, uid auth.UserID) ([]*repository.DynamoHabit, error)
	ArchiveHabit(ctx context.Context, uid auth.UserID, hid uuid.UUID) error
	CreateCheck(ctx context.Context, uid auth.UserID, hid uuid.UUID, date string) (*repository.DynamoCheck, error)
	CreateHabit(ctx context.Context, uid auth.UserID, title string) (*repository.DynamoHabit, error)
	DeleteCheck(ctx context.Context, uid auth.UserID, hid uuid.UUID, date string) error
	DeleteHabit(ctx context.Context, uid auth.UserID, hid uuid.UUID) error
	FindArchivedHabit(ctx context.Context, uid auth.UserID, hid uuid.UUID) (*repository.DynamoHabit, error)
	FindHabit(ctx context.Context, uid auth.UserID, hid uuid.UUID) (*repository.DynamoHabit, error)
	ListLastWeekChecksInAllHabits(ctx context.Context, uid auth.UserID) ([]*repository.DynamoCheck, error)
	ListLatestChecksWithLimit(ctx context.Context, uid auth.UserID, hid uuid.UUID, limit int32) ([]*repository.DynamoCheck, error)
	UnarchiveHabit(ctx context.Context, uid auth.UserID, hid uuid.UUID) error
	UpdateHabit(ctx context.Context, in *repository.DynamoRepositoryUpdateHabitInput) error
}
