package repository

import (
	"context"
	"encoding/json"
	"sort"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/hareku/habit-tracker-app/dynamoconf"
	"github.com/hareku/habit-tracker-app/internal/apperrors"
	"github.com/hareku/habit-tracker-app/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newDynamoRepositoryTest(t *testing.T) *DynamoRepository {
	ctx := context.Background()
	cfg := aws.Config{
		BaseEndpoint: aws.String("http://localhost:8000"),
		Region:       "ap-northeast-1",
		Credentials:  credentials.NewStaticCredentialsProvider("dummy", "dummy", ""),
	}

	var in dynamodb.CreateTableInput
	if err := json.Unmarshal(dynamoconf.Table, &in); err != nil {
		t.Fatalf("unmarshal table config: %v", err)
	}
	in.TableName = aws.String("Test_" + *in.TableName)

	dynamoCli := dynamodb.NewFromConfig(cfg)
	if _, err := dynamoCli.CreateTable(ctx, &in); err != nil {
		t.Fatalf("create table: %v", err)
	}
	t.Cleanup(func() {
		if _, err := dynamoCli.DeleteTable(ctx, &dynamodb.DeleteTableInput{TableName: in.TableName}); err != nil {
			t.Fatalf("delete table: %v", err)
		}
	})

	return &DynamoRepository{
		Client:    dynamoCli,
		TableName: *in.TableName,
	}
}

func Test_AllHabits(t *testing.T) {
	repo := newDynamoRepositoryTest(t)
	ctx := context.Background()

	myUserID := auth.UserID("MyUserID")
	otherUserID := auth.UserID("OtherUserID")

	h1, err := repo.CreateHabit(ctx, myUserID, "Habit1")
	require.NoError(t, err)
	h2, err := repo.CreateHabit(ctx, myUserID, "Habit2")
	require.NoError(t, err)

	_, err = repo.CreateHabit(ctx, otherUserID, "Habit3")
	require.NoError(t, err)

	got, err := repo.AllHabits(ctx, myUserID)
	require.NoError(t, err)
	require.Equal(t, 2, len(got))

	sort.Slice(got, func(i, j int) bool {
		return got[i].Title < got[j].Title
	})
	assert.Equal(t, []*DynamoHabit{h1, h2}, got)
}

func Test_FindHabit(t *testing.T) {
	repo := newDynamoRepositoryTest(t)
	ctx := context.Background()

	myUserID := auth.UserID("MyUserID")

	h1, err := repo.CreateHabit(ctx, myUserID, "Habit1")
	require.NoError(t, err)
	_, err = repo.CreateHabit(ctx, myUserID, "Habit2")
	require.NoError(t, err)

	got, err := repo.FindHabit(ctx, myUserID, h1.ID)
	require.NoError(t, err)
	assert.Equal(t, h1, got)
}

func Test_DeleteHabit(t *testing.T) {
	repo := newDynamoRepositoryTest(t)
	ctx := context.Background()

	myUserID := auth.UserID("MyUserID")

	h1, err := repo.CreateHabit(ctx, myUserID, "Habit1")
	require.NoError(t, err)

	h2, err := repo.CreateHabit(ctx, myUserID, "Habit2")
	require.NoError(t, err)

	require.NoError(t, repo.DeleteHabit(ctx, myUserID, h1.ID))

	got1, err := repo.FindHabit(ctx, myUserID, h1.ID)
	require.Error(t, err, "Got habit1: %+v", got1)
	require.ErrorIs(t, err, apperrors.ErrNotFound)

	got2, err := repo.FindHabit(ctx, myUserID, h2.ID)
	require.NoError(t, err)
	assert.Equal(t, h2, got2)
}

func Test_CreateCheck_Twice(t *testing.T) {
	repo := newDynamoRepositoryTest(t)
	ctx := context.Background()

	myUserID := auth.UserID("MyUserID")

	h1, err := repo.CreateHabit(ctx, myUserID, "Habit1")
	require.NoError(t, err)

	c1, err := repo.CreateCheck(ctx, myUserID, h1.ID, "2000-01-01")
	require.NoError(t, err)
	_, err = repo.CreateCheck(ctx, myUserID, h1.ID, c1.Date)
	require.Error(t, err)
	require.ErrorIs(t, err, apperrors.ErrConflict)

	h1, err = repo.FindHabit(ctx, myUserID, h1.ID)
	require.NoError(t, err)
	require.Equal(t, 1, h1.ChecksCount)
}

func Test_DeleteCheck_Twice(t *testing.T) {
	repo := newDynamoRepositoryTest(t)
	ctx := context.Background()

	myUserID := auth.UserID("MyUserID")

	h1, err := repo.CreateHabit(ctx, myUserID, "Habit1")
	require.NoError(t, err)

	c1, err := repo.CreateCheck(ctx, myUserID, h1.ID, "2000-01-01")
	require.NoError(t, err)
	require.NoError(t, repo.DeleteCheck(ctx, myUserID, h1.ID, c1.Date))

	h1, err = repo.FindHabit(ctx, myUserID, h1.ID)
	require.NoError(t, err)
	require.Equal(t, 0, h1.ChecksCount)

	require.ErrorIs(t, repo.DeleteCheck(ctx, myUserID, h1.ID, c1.Date), apperrors.ErrNotFound)

	h1, err = repo.FindHabit(ctx, myUserID, h1.ID)
	require.NoError(t, err)
	require.Equal(t, 0, h1.ChecksCount)
}

func Test_ListLatestChecksWithLimit(t *testing.T) {
	repo := newDynamoRepositoryTest(t)
	ctx := context.Background()

	myUserID := auth.UserID("MyUserID")

	h1, err := repo.CreateHabit(ctx, myUserID, "Habit1")
	require.NoError(t, err)

	c1, err := repo.CreateCheck(ctx, myUserID, h1.ID, "2000-01-01")
	require.NoError(t, err)
	c2, err := repo.CreateCheck(ctx, myUserID, h1.ID, "2000-01-02")
	require.NoError(t, err)

	got1, err := repo.ListLatestChecksWithLimit(ctx, myUserID, h1.ID, 1)
	require.NoError(t, err)
	require.Len(t, got1, 1)
	require.Equal(t, []*DynamoCheck{c2}, got1)

	got2, err := repo.ListLatestChecksWithLimit(ctx, myUserID, h1.ID, 2)
	require.NoError(t, err)
	require.Len(t, got2, 2)
	require.Equal(t, []*DynamoCheck{c2, c1}, got2)
}
