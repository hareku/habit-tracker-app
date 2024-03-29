package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/hareku/habit-tracker-app/internal/apperrors"
	"github.com/hareku/habit-tracker-app/internal/auth"
)

type DynamoRepository struct {
	Client    *dynamodb.Client
	TableName string
}

type DynamoHabit struct {
	PK          string
	SK          string
	ID          string `dynamodbav:"UUID"`
	UserID      auth.UserID
	Title       string
	ChecksCount int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewDynamoHabit(userID auth.UserID, habitID string) *DynamoHabit {
	return &DynamoHabit{
		PK:     fmt.Sprintf("USER#%s", userID),
		SK:     fmt.Sprintf("HABITS#%s", habitID),
		UserID: userID,
		ID:     habitID,
	}
}

// GetKey returns the composite primary key of the habit in a format that can be
// sent to DynamoDB.
func (h *DynamoHabit) GetKey() map[string]types.AttributeValue {
	pk, err := attributevalue.Marshal(h.PK)
	if err != nil {
		panic(fmt.Errorf("marshal PK: %w", err))
	}
	sk, err := attributevalue.Marshal(h.SK)
	if err != nil {
		panic(fmt.Errorf("marshal SK: %w", err))
	}
	return map[string]types.AttributeValue{"PK": pk, "SK": sk}
}

type DynamoCheck struct {
	PK             string
	SK             string
	CheckDateLSISK string
	HabitID        string `dynamodbav:"HabitUUID"`
	Date           string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewDynamoCheck(userID auth.UserID, habitID, date string) *DynamoCheck {
	return &DynamoCheck{
		PK:             fmt.Sprintf("USER#%s", userID),
		SK:             fmt.Sprintf("HABIT#%s__CHECK_DATE#%s", habitID, date),
		CheckDateLSISK: fmt.Sprintf("CHECK_DATE#%s__HABIT#%s", date, habitID),
		HabitID:        habitID,
		Date:           date,
	}
}

// GetKey returns the composite primary key of the check in a format that can be
// sent to DynamoDB.
func (h *DynamoCheck) GetKey() map[string]types.AttributeValue {
	pk, err := attributevalue.Marshal(h.PK)
	if err != nil {
		panic(err)
	}
	sk, err := attributevalue.Marshal(h.SK)
	if err != nil {
		panic(err)
	}
	return map[string]types.AttributeValue{"PK": pk, "SK": sk}
}

func (r *DynamoRepository) AllHabits(ctx context.Context, uid auth.UserID) ([]*DynamoHabit, error) {
	expr, err := expression.NewBuilder().
		WithKeyCondition(
			expression.Key("PK").Equal(expression.Value(fmt.Sprintf("USER#%s", uid))).
				And(expression.Key("SK").BeginsWith("HABITS#")),
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
		if err := attributevalue.UnmarshalListOfMapsWithOptions(resp.Items, &pageItems); err != nil {
			return nil, fmt.Errorf("unmarshal items: %w", err)
		}
		habits = append(habits, pageItems...)
	}

	return habits, nil
}

func (r *DynamoRepository) FindHabit(ctx context.Context, uid auth.UserID, hid string) (*DynamoHabit, error) {
	h := NewDynamoHabit(uid, hid)
	expr, err := expression.NewBuilder().
		WithKeyCondition(
			expression.Key("PK").Equal(expression.Value(h.PK)).
				And(expression.Key("SK").Equal(expression.Value(h.SK))),
		).
		Build()
	if err != nil {
		return nil, fmt.Errorf("build expression: %w", err)
	}

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
		if err := attributevalue.UnmarshalListOfMapsWithOptions(resp.Items, &pageItems); err != nil {
			return nil, fmt.Errorf("unmarshal items: %w", err)
		}
		if len(pageItems) == 1 {
			return pageItems[0], nil
		}
	}
	return nil, apperrors.ErrNotFound
}

func (r *DynamoRepository) CreateHabit(ctx context.Context, uid auth.UserID, title string) (*DynamoHabit, error) {
	h := NewDynamoHabit(uid, uuid.New().String())
	h.Title = title
	h.CreatedAt = time.Now().Round(time.Nanosecond)
	h.UpdatedAt = h.CreatedAt

	item, err := attributevalue.MarshalMap(h)
	if err != nil {
		return nil, fmt.Errorf("marshal habit: %w", err)
	}
	resp, err := r.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &r.TableName,
		Item:      item,
	})
	if err != nil {
		return nil, fmt.Errorf("put item: %w", err)
	}

	if err := attributevalue.UnmarshalMap(resp.Attributes, &h); err != nil {
		return nil, fmt.Errorf("unmarshal item: %w", err)
	}

	return h, nil
}

func (r *DynamoRepository) DeleteHabit(ctx context.Context, uid auth.UserID, hid string) error {
	_, err := r.Client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &r.TableName,
		Key:       NewDynamoHabit(uid, hid).GetKey(),
	})
	if err != nil {
		return fmt.Errorf("delete item: %w", err)
	}
	return nil
}

type DynamoRepositoryUpdateHabitInput struct {
	UserID  auth.UserID
	HabitID string
	Title   string
}

func (r *DynamoRepository) UpdateHabit(ctx context.Context, in *DynamoRepositoryUpdateHabitInput) error {
	h := NewDynamoHabit(in.UserID, in.HabitID)

	update := expression.Set(expression.Name("Title"), expression.Value(in.Title))
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		return fmt.Errorf("build expression: %w", err)
	}
	_, err = r.Client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:                 &r.TableName,
		Key:                       h.GetKey(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
		ReturnValues:              types.ReturnValueNone,
	})
	if err != nil {
		return fmt.Errorf("update item: %w", err)
	}
	return nil
}

func (r *DynamoRepository) ListLatestChecksWithLimit(ctx context.Context, uid auth.UserID, hid string, limit int32) ([]*DynamoCheck, error) {
	expr, err := expression.NewBuilder().
		WithKeyCondition(
			expression.Key("PK").Equal(expression.Value(fmt.Sprintf("USER#%s", uid))).
				And(expression.Key("SK").BeginsWith(fmt.Sprintf("HABIT#%s__CHECK_DATE#", hid))),
		).
		Build()
	if err != nil {
		return nil, fmt.Errorf("build expression: %w", err)
	}

	paginator := dynamodb.NewQueryPaginator(r.Client, &dynamodb.QueryInput{
		TableName:                 &r.TableName,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		ProjectionExpression:      expr.Projection(),
		Limit:                     &limit,
		ScanIndexForward:          aws.Bool(false),
	})

	resp, err := paginator.NextPage(ctx)
	if err != nil {
		return nil, fmt.Errorf("query paginator: %w", err)
	}
	var pageItems []*DynamoCheck
	if err := attributevalue.UnmarshalListOfMapsWithOptions(resp.Items, &pageItems); err != nil {
		return nil, fmt.Errorf("unmarshal items: %w", err)
	}
	return pageItems, nil
}

func (r *DynamoRepository) ListLastWeekChecksInAllHabits(ctx context.Context, uid auth.UserID) ([]*DynamoCheck, error) {
	minTime := time.Now().Add(time.Hour * 24 * 7 * -1).Format("2006-01-02")

	expr, err := expression.NewBuilder().
		WithKeyCondition(
			expression.Key("PK").Equal(expression.Value(fmt.Sprintf("USER#%s", uid))).
				And(expression.Key("CheckDateLSISK").
					GreaterThanEqual(expression.Value(fmt.Sprintf("CHECK_DATE#%s", minTime))),
				),
		).
		Build()
	if err != nil {
		return nil, fmt.Errorf("build expression: %w", err)
	}

	var checks []*DynamoCheck
	paginator := dynamodb.NewQueryPaginator(r.Client, &dynamodb.QueryInput{
		TableName:                 &r.TableName,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		ProjectionExpression:      expr.Projection(),
		IndexName:                 aws.String("CheckDateLSI"),
	})
	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("query paginator: %w", err)
		}

		var pageItems []*DynamoCheck
		if err := attributevalue.UnmarshalListOfMapsWithOptions(resp.Items, &pageItems); err != nil {
			return nil, fmt.Errorf("unmarshal items: %w", err)
		}
		checks = append(checks, pageItems...)
	}

	return checks, nil
}

func (r *DynamoRepository) CreateCheck(ctx context.Context, uid auth.UserID, hid, date string) (*DynamoCheck, error) {
	c := NewDynamoCheck(uid, hid, date)
	c.CreatedAt = time.Now().Round(time.Nanosecond)
	c.UpdatedAt = c.CreatedAt

	item, err := attributevalue.MarshalMap(c)
	if err != nil {
		return nil, fmt.Errorf("marshal check: %w", err)
	}

	h := NewDynamoHabit(uid, hid)

	update := expression.Set(expression.Name("ChecksCount"), expression.Name("ChecksCount").Plus(expression.Value(1)))
	updateExpr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		return nil, fmt.Errorf("build update expression: %w", err)
	}

	condition := expression.Not(
		expression.AttributeExists(expression.Name("PK")).
			And(expression.AttributeExists(expression.Name("SK"))),
	)
	conditionExpr, err := expression.NewBuilder().WithCondition(condition).Build()
	if err != nil {
		return nil, fmt.Errorf("build condition expression: %w", err)
	}

	if _, err := r.Client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			{
				Put: &types.Put{
					TableName:                 &r.TableName,
					Item:                      item,
					ConditionExpression:       conditionExpr.Condition(),
					ExpressionAttributeNames:  conditionExpr.Names(),
					ExpressionAttributeValues: conditionExpr.Values(),
				},
			},
			{
				Update: &types.Update{
					TableName:                 &r.TableName,
					Key:                       h.GetKey(),
					ExpressionAttributeNames:  updateExpr.Names(),
					ExpressionAttributeValues: updateExpr.Values(),
					UpdateExpression:          updateExpr.Update(),
				},
			},
		},
	}); err != nil {
		var tce *types.TransactionCanceledException
		if errors.As(err, &tce) && *tce.CancellationReasons[0].Code == string(types.BatchStatementErrorCodeEnumConditionalCheckFailed) {
			return nil, fmt.Errorf("condition check failed %w: %w", apperrors.ErrConflict, tce)
		}

		return nil, fmt.Errorf("transact write items: %w", err)
	}

	return c, nil
}

func (r *DynamoRepository) DeleteCheck(ctx context.Context, uid auth.UserID, hid, date string) error {
	c := &DynamoCheck{
		PK: fmt.Sprintf("USER#%s", uid),
		SK: fmt.Sprintf("HABIT#%s__CHECK_DATE#%s", hid, date),
	}
	h := NewDynamoHabit(uid, hid)

	update := expression.Set(expression.Name("ChecksCount"), expression.Name("ChecksCount").Plus(expression.Value(-1)))
	updateExpr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		return fmt.Errorf("build update expression: %w", err)
	}

	condition := expression.AttributeExists(expression.Name("PK")).
		And(expression.AttributeExists(expression.Name("SK")))
	conditionExpr, err := expression.NewBuilder().WithCondition(condition).Build()
	if err != nil {
		return fmt.Errorf("build condition expression: %w", err)
	}

	if _, err := r.Client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			{
				Delete: &types.Delete{
					TableName:                 &r.TableName,
					Key:                       c.GetKey(),
					ConditionExpression:       conditionExpr.Condition(),
					ExpressionAttributeNames:  conditionExpr.Names(),
					ExpressionAttributeValues: conditionExpr.Values(),
				},
			},
			{
				Update: &types.Update{
					TableName:                 &r.TableName,
					Key:                       h.GetKey(),
					ExpressionAttributeNames:  updateExpr.Names(),
					ExpressionAttributeValues: updateExpr.Values(),
					UpdateExpression:          updateExpr.Update(),
				},
			},
		},
	}); err != nil {
		var tce *types.TransactionCanceledException
		if errors.As(err, &tce) && *tce.CancellationReasons[0].Code == string(types.BatchStatementErrorCodeEnumConditionalCheckFailed) {
			return fmt.Errorf("condition check failed: %w: %w", apperrors.ErrNotFound, tce)
		}

		return fmt.Errorf("transact write items: %w", err)
	}
	return nil
}
