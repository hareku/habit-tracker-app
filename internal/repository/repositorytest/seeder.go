package repositorytest

import (
	"math/rand"
	"time"

	"github.com/brianvoe/gofakeit/v7"
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
	faker := gofakeit.NewFaker(rand.New(s.src), true)

	h := repository.NewDynamoHabit(userID, habitUUID)
	h.Title = faker.ProductCategory()
	h.CreatedAt = faker.Date()
	h.UpdatedAt = faker.DateRange(h.CreatedAt, h.CreatedAt.Add(time.Hour*24*30))

	if f != nil {
		f(h)
	}
	return h
}

func (s *Seeder) SeedArchivedHabit(userID auth.UserID, f func(h *repository.DynamoHabit)) *repository.DynamoHabit {
	habitUUID := uuid.Must(uuid.NewRandomFromReader(rand.New(s.src)))
	faker := gofakeit.NewFaker(rand.New(s.src), true)

	h := repository.NewArchivedDynamoHabit(userID, habitUUID)
	h.Title = faker.ProductCategory()
	h.CreatedAt = faker.Date()
	h.UpdatedAt = faker.DateRange(h.CreatedAt, h.CreatedAt.Add(time.Hour*24*30))

	if f != nil {
		f(h)
	}
	return h
}

func (s *Seeder) SeedCheck(userID auth.UserID, habitUUID uuid.UUID, date string, f func(c *repository.DynamoCheck)) *repository.DynamoCheck {
	faker := gofakeit.NewFaker(rand.New(s.src), true)

	c := repository.NewDynamoCheck(userID, habitUUID, date)
	c.Date = faker.Date().Format("2006-01-02")
	c.CreatedAt = faker.Date()
	c.UpdatedAt = c.CreatedAt

	if f != nil {
		f(c)
	}
	return c
}
