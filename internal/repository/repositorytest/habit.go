package repositorytest

import (
	"math/rand"

	"github.com/google/uuid"
	"github.com/hareku/habit-tracker-app/internal/auth"
	"github.com/hareku/habit-tracker-app/internal/repository"
)

// Seeder is a helper for seeding habit data.
// It always generates the same random data for testing.
type Seeder struct {
	src rand.Source
}

func NewSeeder() *Seeder {
	return &Seeder{
		src: rand.NewSource(1),
	}
}

func (s *Seeder) SeedHabit(userID auth.UserID, f func(h *repository.DynamoHabit)) *repository.DynamoHabit {
	habitUUID := uuid.Must(uuid.NewRandomFromReader(rand.New(s.src)))
	h := repository.NewDynamoHabit(userID, habitUUID)
	if f != nil {
		f(h)
	}
	return h
}

func (s *Seeder) SeedArchivedHabit(userID auth.UserID, f func(h *repository.DynamoHabit)) *repository.DynamoHabit {
	habitUUID := uuid.Must(uuid.NewRandomFromReader(rand.New(s.src)))
	h := repository.NewArchivedDynamoHabit(userID, habitUUID)
	if f != nil {
		f(h)
	}
	return h
}

func (s *Seeder) SeedCheck(userID auth.UserID, habitUUID uuid.UUID, date string, f func(c *repository.DynamoCheck)) *repository.DynamoCheck {
	c := repository.NewDynamoCheck(userID, habitUUID, date)
	if f != nil {
		f(c)
	}
	return c
}
