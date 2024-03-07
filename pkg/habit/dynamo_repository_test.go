package habit

import (
	"context"
	"encoding/json"
	"sort"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"
	"github.com/hareku/habit-tracker-app/dynamoconf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newDynamoRepositoryTest(t *testing.T) *DynamoRepository {
	time.Local = nil

	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		t.Fatalf("load aws config: %v", err)
	}
	cfg.BaseEndpoint = aws.String("http://localhost:8000")
	cfg.Region = "ap-northeast-1"

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

	myUserID := UserID("MyUserID")
	otherUserID := UserID("OtherUserID")

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

	myUserID := UserID("MyUserID")

	h1, err := repo.CreateHabit(ctx, myUserID, "Habit1")
	require.NoError(t, err)
	_, err = repo.CreateHabit(ctx, myUserID, "Habit2")
	require.NoError(t, err)

	got, err := repo.FindHabit(ctx, myUserID, uuid.MustParse(h1.UUID))
	require.NoError(t, err)
	assert.Equal(t, h1, got)
}

func Test_DeleteHabit(t *testing.T) {
	repo := newDynamoRepositoryTest(t)
	ctx := context.Background()

	myUserID := UserID("MyUserID")

	h1, err := repo.CreateHabit(ctx, myUserID, "Habit1")
	require.NoError(t, err)

	h2, err := repo.CreateHabit(ctx, myUserID, "Habit2")
	require.NoError(t, err)

	require.NoError(t, repo.DeleteHabit(ctx, myUserID, uuid.MustParse(h1.UUID)))

	got1, err := repo.FindHabit(ctx, myUserID, uuid.MustParse(h1.UUID))
	require.Error(t, err, "Got habit1: %+v", got1)
	require.ErrorIs(t, err, ErrNotFound)

	got2, err := repo.FindHabit(ctx, myUserID, uuid.MustParse(h2.UUID))
	require.NoError(t, err)
	assert.Equal(t, h2, got2)
}

func Test_ListLatestChecksWithLimit(t *testing.T) {
	repo := newDynamoRepositoryTest(t)
	ctx := context.Background()

	myUserID := UserID("MyUserID")

	h1, err := repo.CreateHabit(ctx, myUserID, "Habit1")
	require.NoError(t, err)

	c1, err := repo.CreateCheck(ctx, myUserID, uuid.MustParse(h1.UUID), "2000-01-01")
	require.NoError(t, err)
	c2, err := repo.CreateCheck(ctx, myUserID, uuid.MustParse(h1.UUID), "2000-01-02")
	require.NoError(t, err)

	got1, err := repo.ListLatestChecksWithLimit(ctx, myUserID, uuid.MustParse(h1.UUID), 1)
	require.NoError(t, err)
	require.Len(t, got1, 1)
	require.Equal(t, []*DynamoCheck{c2}, got1)

	got2, err := repo.ListLatestChecksWithLimit(ctx, myUserID, uuid.MustParse(h1.UUID), 2)
	require.NoError(t, err)
	require.Len(t, got2, 2)
	require.Equal(t, []*DynamoCheck{c2, c1}, got2)
}
