package api

import (
	"context"

	firebase "firebase.google.com/go/auth"
	"github.com/google/uuid"
	"github.com/hareku/habit-tracker-app/internal/auth"
	"github.com/hareku/habit-tracker-app/internal/repository"
)

type Authenticator interface {
	Authenticate(ctx context.Context, session string) (context.Context, error)
	DeleteUser(ctx context.Context, uid auth.UserID) error
	GetUser(ctx context.Context, uid auth.UserID) (*firebase.UserRecord, error)
	SessionCookie(ctx context.Context, idToken string) (string, error)
}

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
