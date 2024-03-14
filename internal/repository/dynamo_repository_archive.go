package repository

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/hareku/habit-tracker-app/internal/auth"
)

func NewArchivedDynamoHabit(userID auth.UserID, habitUUID uuid.UUID) *DynamoHabit {
	h := NewDynamoHabit(userID, habitUUID)
	h.SK = fmt.Sprintf("ARCHIVED_HABITS#%s", habitUUID)
	return h
}

func (r *DynamoRepository) AllArchivedHabits(ctx context.Context, uid auth.UserID) ([]*DynamoHabit, error) {
	expr, err := expression.NewBuilder().
		WithKeyCondition(
			expression.Key("PK").Equal(expression.Value(fmt.Sprintf("USER#%s", uid))).
				And(expression.Key("SK").BeginsWith("ARCHIVED_HABITS#")),
		).
		Build()
	if err != nil {
		return nil, fmt.Errorf("build expression: %w", err)
	}

	var habits []*DynamoHabit
	paginator := dynamodb.NewQueryPaginator(r.Client, &dynamodb.QueryInput{
		TableName:                 &r.TableName,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		ProjectionExpression:      expr.Projection(),
	})
	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("query paginator: %w", err)
		}

		var pageItems []*DynamoHabit
		if err := attributevalue.UnmarshalListOfMaps(resp.Items, &pageItems); err != nil {
			return nil, fmt.Errorf("unmarshal items: %w", err)
		}
		habits = append(habits, pageItems...)
	}

	return habits, nil
}

func (r *DynamoRepository) FindArchivedHabit(ctx context.Context, uid auth.UserID, hid uuid.UUID) (*DynamoHabit, error) {
	h := NewArchivedDynamoHabit(uid, hid)

	resp, err := r.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &r.TableName,
		Key:       h.GetKey(),
	})
	if err != nil {
		return nil, fmt.Errorf("get item: %w", err)
	}
	if err := attributevalue.UnmarshalMap(resp.Item, &h); err != nil {
		return nil, fmt.Errorf("unmarshal item: %w", err)
	}
	return h, nil
}

func (r *DynamoRepository) ArchiveHabit(ctx context.Context, uid auth.UserID, hid uuid.UUID) error {
	h, err := r.FindHabit(ctx, uid, hid)
	if err != nil {
		return fmt.Errorf("find a habit [%s]: %w", hid, err)
	}
	deleteKey := h.GetKey()

	h.SK = fmt.Sprintf("ARCHIVED_HABITS#%s", hid)
	item, err := attributevalue.MarshalMap(h)
	if err != nil {
		return fmt.Errorf("marshal habit: %w", err)
	}

	if _, err := r.Client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			{
				Delete: &types.Delete{
					TableName: &r.TableName,
					Key:       deleteKey,
				},
			},
			{
				Put: &types.Put{
					TableName: &r.TableName,
					Item:      item,
				},
			},
		},
	}); err != nil {
		return fmt.Errorf("transact write items: %w", err)
	}

	return nil
}

func (r *DynamoRepository) UnarchiveHabit(ctx context.Context, uid auth.UserID, hid uuid.UUID) error {
	h, err := r.FindArchivedHabit(ctx, uid, hid)
	if err != nil {
		return fmt.Errorf("find a habit [%s]: %w", hid, err)
	}
	deleteKey := h.GetKey()

	h.SK = fmt.Sprintf("HABITS#%s", hid)
	item, err := attributevalue.MarshalMap(h)
	if err != nil {
		return fmt.Errorf("marshal habit: %w", err)
	}

	if _, err := r.Client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			{
				Delete: &types.Delete{
					TableName: &r.TableName,
					Key:       deleteKey,
				},
			},
			{
				Put: &types.Put{
					TableName: &r.TableName,
					Item:      item,
				},
			},
		},
	}); err != nil {
		return fmt.Errorf("transact write items: %w", err)
	}

	return nil
}
