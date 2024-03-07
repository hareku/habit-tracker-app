package habit

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/hareku/habit-tracker-app/dynamoconf"
	"github.com/stretchr/testify/require"
)

func newDynamoRepositoryTest(t *testing.T) *DynamoRepository {
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

func TestHello(t *testing.T) {
	repo := newDynamoRepositoryTest(t)
	ctx := context.Background()

	myUserID := UserID("MyUserID")
	otherUserID := UserID("OtherUserID")

	h1, err := repo.CreateHabit(ctx, myUserID, "Habit1")
	require.NoError(t, err)

	_, err = repo.CreateHabit(ctx, otherUserID, "Habit2")
	require.NoError(t, err)

	got, err := repo.AllHabits(ctx, myUserID)
	require.NoError(t, err)
	require.Equal(t, 1, len(got))

	require.Equal(t, h1, got[0])
}
