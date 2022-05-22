package habit

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/guregu/dynamo"
)

type DynamoRepository struct {
	DB    *dynamo.DB
	Table dynamo.Table
}

type DynamoHabit struct {
	PK          string
	SK          string
	UUID        uuid.UUID
	UserID      UserID
	Title       string
	ChecksCount int
	CreatedAt   time.Time
	UpdatedAt   time.Time

	LatestCheck *DynamoCheck
}

type DynamoCheck struct {
	PK             string
	SK             string
	CheckDateLSISK string
	HabitUUID      uuid.UUID
	Date           string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (r *DynamoRepository) AllHabits(ctx context.Context, uid UserID) ([]*DynamoHabit, error) {
	var habits []*DynamoHabit
	err := r.Table.Get("PK", fmt.Sprintf("USER#%s", uid)).
		Range("SK", dynamo.BeginsWith, "HABITS#").
		AllWithContext(ctx, &habits)
	if err != nil {
		return nil, fmt.Errorf("dynamo: %w", err)
	}

	return habits, nil
}

func (r *DynamoRepository) AllArchivedHabits(ctx context.Context, uid UserID) ([]*DynamoHabit, error) {
	var habits []*DynamoHabit
	err := r.Table.Get("PK", fmt.Sprintf("USER#%s", uid)).
		Range("SK", dynamo.BeginsWith, "ARCHIVED_HABITS#").
		AllWithContext(ctx, &habits)
	if err != nil {
		return nil, fmt.Errorf("dynamo: %w", err)
	}

	return habits, nil
}

func (r *DynamoRepository) FindHabit(ctx context.Context, uid UserID, hid uuid.UUID) (*DynamoHabit, error) {
	var habit *DynamoHabit
	err := r.Table.Get("PK", fmt.Sprintf("USER#%s", uid)).
		Range("SK", dynamo.Equal, fmt.Sprintf("HABITS#%s", hid)).
		OneWithContext(ctx, &habit)
	if err != nil {
		return nil, fmt.Errorf("dynamo: %w", err)
	}

	return habit, nil
}

func (r *DynamoRepository) FindArchivedHabit(ctx context.Context, uid UserID, hid uuid.UUID) (*DynamoHabit, error) {
	var habit *DynamoHabit
	err := r.Table.Get("PK", fmt.Sprintf("USER#%s", uid)).
		Range("SK", dynamo.Equal, fmt.Sprintf("ARCHIVED_HABITS#%s", hid)).
		OneWithContext(ctx, &habit)
	if err != nil {
		return nil, fmt.Errorf("dynamo: %w", err)
	}

	return habit, nil
}

func (r *DynamoRepository) CreateHabit(ctx context.Context, uid UserID, title string) (*DynamoHabit, error) {
	id := uuid.New()
	now := time.Now()

	h := &DynamoHabit{
		PK:        fmt.Sprintf("USER#%s", uid),
		SK:        fmt.Sprintf("HABITS#%s", id),
		UUID:      id,
		UserID:    uid,
		Title:     title,
		CreatedAt: now,
		UpdatedAt: now,
	}
	err := r.Table.Put(h).RunWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("dynamo: %w", err)
	}

	return h, nil
}

func (r *DynamoRepository) DeleteHabit(ctx context.Context, uid UserID, hid uuid.UUID) error {
	err := r.Table.Delete("PK", fmt.Sprintf("USER#%s", uid)).
		Range("SK", fmt.Sprintf("HABITS#%s", hid)).
		RunWithContext(ctx)
	if err != nil {
		return fmt.Errorf("dynamo: %w", err)
	}

	return nil
}

func (r *DynamoRepository) ArchiveHabit(ctx context.Context, uid UserID, hid uuid.UUID) (*DynamoHabit, error) {
	h, err := r.FindHabit(ctx, uid, hid)
	if err != nil {
		return nil, fmt.Errorf("failed to find a habit [%s]: %w", hid, err)
	}

	delete := r.Table.Delete("PK", h.PK).
		Range("SK", h.SK)

	h.SK = fmt.Sprintf("ARCHIVED_HABITS#%s", hid)
	h.UpdatedAt = time.Now()
	put := r.Table.Put(h)

	if err := r.DB.WriteTx().Delete(delete).Put(put).RunWithContext(ctx); err != nil {
		return nil, fmt.Errorf("dynamo: %w", err)
	}

	return h, nil
}

type DynamoRepositoryUpdateHabitInput struct {
	UserID    UserID
	HabitUUID uuid.UUID
	Title     string
}

func (r *DynamoRepository) UpdateHabit(ctx context.Context, in *DynamoRepositoryUpdateHabitInput) error {
	return r.Table.Update("PK", fmt.Sprintf("USER#%s", in.UserID)).
		Range("SK", fmt.Sprintf("HABITS#%s", in.HabitUUID)).
		Set("Title", in.Title).
		RunWithContext(ctx)
}

func (r *DynamoRepository) UnarchiveHabit(ctx context.Context, uid UserID, hid uuid.UUID) (*DynamoHabit, error) {
	h, err := r.FindArchivedHabit(ctx, uid, hid)
	if err != nil {
		return nil, fmt.Errorf("failed to find a habit [%s]: %w", hid, err)
	}

	delete := r.Table.Delete("PK", h.PK).
		Range("SK", h.SK)

	h.SK = fmt.Sprintf("HABITS#%s", hid)
	h.UpdatedAt = time.Now()
	put := r.Table.Put(h)

	if err := r.DB.WriteTx().Delete(delete).Put(put).RunWithContext(ctx); err != nil {
		return nil, fmt.Errorf("dynamo: %w", err)
	}

	return h, nil
}

func (r *DynamoRepository) ListChecks(ctx context.Context, uid UserID, hid uuid.UUID) ([]*DynamoCheck, error) {
	var checks []*DynamoCheck
	err := r.Table.Get("PK", fmt.Sprintf("USER#%s", uid)).
		Range("SK", dynamo.BeginsWith, fmt.Sprintf("HABIT#%s__CHECK_DATE#", hid)).
		Order(dynamo.Descending).
		AllWithContext(ctx, &checks)
	if err != nil {
		return nil, fmt.Errorf("dynamo: %w", err)
	}

	return checks, nil
}

func (r *DynamoRepository) ListLatestChecksWithLimit(ctx context.Context, uid UserID, hid uuid.UUID, limit int64) ([]*DynamoCheck, error) {
	var checks []*DynamoCheck
	err := r.Table.Get("PK", fmt.Sprintf("USER#%s", uid)).
		Range("SK",
			dynamo.BeginsWith,
			fmt.Sprintf("HABIT#%s__CHECK_DATE#", hid)).
		Order(dynamo.Descending).
		Limit(limit).
		AllWithContext(ctx, &checks)
	if err != nil {
		return nil, fmt.Errorf("dynamo: %w", err)
	}

	return checks, nil
}

func (r *DynamoRepository) ListLastWeekChecksInAllHabits(ctx context.Context, uid UserID) ([]*DynamoCheck, error) {
	var checks []*DynamoCheck
	err := r.Table.Get("PK", fmt.Sprintf("USER#%s", uid)).
		Range("CheckDateLSISK",
			dynamo.GreaterOrEqual,
			fmt.Sprintf("CHECK_DATE#%s", time.Now().Add(time.Hour*24*7*-1).Format("2006-01-02"))).
		Index("CheckDateLSI").
		AllWithContext(ctx, &checks)
	if err != nil {
		return nil, fmt.Errorf("dynamo: %w", err)
	}

	return checks, nil
}

func (r *DynamoRepository) CreateCheck(ctx context.Context, uid UserID, hid uuid.UUID, date string) (*DynamoCheck, error) {
	now := time.Now()
	c := &DynamoCheck{
		PK:             fmt.Sprintf("USER#%s", uid),
		SK:             fmt.Sprintf("HABIT#%s__CHECK_DATE#%s", hid, date),
		CheckDateLSISK: fmt.Sprintf("CHECK_DATE#%s__HABIT#%s", date, hid),
		HabitUUID:      hid,
		Date:           date,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	put := r.Table.Put(c)
	increment := r.Table.Update("PK", fmt.Sprintf("USER#%s", uid)).
		Range("SK", fmt.Sprintf("HABITS#%s", hid)).
		SetExpr("'ChecksCount' = 'ChecksCount' + ?", 1)

	err := r.DB.WriteTx().Put(put).Update(increment).RunWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("dynamo: %w", err)
	}

	return c, nil
}

func (r *DynamoRepository) DeleteCheck(ctx context.Context, uid UserID, hid uuid.UUID, date string) error {
	del := r.Table.Delete("PK", fmt.Sprintf("USER#%s", uid)).
		Range("SK", fmt.Sprintf("HABIT#%s__CHECK_DATE#%s", hid, date))
	decrement := r.Table.Update("PK", fmt.Sprintf("USER#%s", uid)).
		Range("SK", fmt.Sprintf("HABITS#%s", hid)).
		SetExpr("'ChecksCount' = 'ChecksCount' - ?", 1)

	err := r.DB.WriteTx().Delete(del).Update(decrement).RunWithContext(ctx)
	if err != nil {
		return fmt.Errorf("dynamo: %w", err)
	}

	return nil
}
