package habit

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

type DynamoRepository struct {
	Client    *dynamodb.Client
	TableName string
}

type DynamoHabit struct {
	PK          string
	SK          string
	UUID        string
	UserID      UserID
	Title       string
	ChecksCount int
	CreatedAt   time.Time
	UpdatedAt   time.Time

	LatestCheck *DynamoCheck
}

// GetKey returns the composite primary key of the movie in a format that can be
// sent to DynamoDB.
func (h *DynamoHabit) GetKey() map[string]types.AttributeValue {
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

func (h *DynamoHabit) beforeReturn() {
	h.CreatedAt = h.CreatedAt.UTC()
	h.UpdatedAt = h.UpdatedAt.UTC()
}

type DynamoCheck struct {
	PK             string
	SK             string
	CheckDateLSISK string
	HabitUUID      string
	Date           string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// GetKey returns the composite primary key of the movie in a format that can be
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

func (r *DynamoRepository) AllHabits(ctx context.Context, uid UserID) ([]*DynamoHabit, error) {
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

	for _, h := range habits {
		h.beforeReturn()
	}
	return habits, nil
}

func (r *DynamoRepository) AllArchivedHabits(ctx context.Context, uid UserID) ([]*DynamoHabit, error) {
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

	for _, h := range habits {
		h.beforeReturn()
	}
	return habits, nil
}

func (r *DynamoRepository) FindHabit(ctx context.Context, uid UserID, hid uuid.UUID) (*DynamoHabit, error) {
	h := &DynamoHabit{
		PK: fmt.Sprintf("USER#%s", uid),
		SK: fmt.Sprintf("HABITS#%s", hid),
	}

	resp, err := r.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &r.TableName,
		Key:       h.GetKey(),
	})
	if err != nil {
		return nil, fmt.Errorf("dynamo: %w", err)
	}

	if err := attributevalue.UnmarshalMap(resp.Item, &h); err != nil {
		return nil, fmt.Errorf("unmarshal item: %w", err)
	}

	h.beforeReturn()
	return h, nil
}

func (r *DynamoRepository) FindArchivedHabit(ctx context.Context, uid UserID, hid uuid.UUID) (*DynamoHabit, error) {
	var habit *DynamoHabit
	habit.PK = fmt.Sprintf("USER#%s", uid)
	habit.SK = fmt.Sprintf("ARCHIVED_HABITS#%s", hid)

	resp, err := r.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &r.TableName,
		Key:       habit.GetKey(),
	})
	if err != nil {
		return nil, fmt.Errorf("dynamo: %w", err)
	}

	if err := attributevalue.UnmarshalMap(resp.Item, &habit); err != nil {
		return nil, fmt.Errorf("unmarshal item: %w", err)
	}
	return habit, nil
}

func (r *DynamoRepository) CreateHabit(ctx context.Context, uid UserID, title string) (*DynamoHabit, error) {
	id := uuid.New()
	now := time.Now().UTC()

	h := &DynamoHabit{
		PK:        fmt.Sprintf("USER#%s", uid),
		SK:        fmt.Sprintf("HABITS#%s", id),
		UUID:      id.String(),
		UserID:    uid,
		Title:     title,
		CreatedAt: now,
		UpdatedAt: now,
	}
	item, err := attributevalue.MarshalMap(h)
	if err != nil {
		return nil, fmt.Errorf("marshal habit: %w", err)
	}
	if _, err := r.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &r.TableName,
		Item:      item,
	}); err != nil {
		return nil, fmt.Errorf("put item: %w", err)
	}

	h.beforeReturn()
	return h, nil
}

func (r *DynamoRepository) DeleteHabit(ctx context.Context, uid UserID, hid uuid.UUID) error {
	h := &DynamoHabit{
		PK: fmt.Sprintf("USER#%s", uid),
		SK: fmt.Sprintf("HABITS#%s", hid),
	}

	if _, err := r.Client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &r.TableName,
		Key:       h.GetKey(),
	}); err != nil {
		return fmt.Errorf("delete item: %w", err)
	}
	return nil
}

func (r *DynamoRepository) ArchiveHabit(ctx context.Context, uid UserID, hid uuid.UUID) (*DynamoHabit, error) {
	h, err := r.FindHabit(ctx, uid, hid)
	if err != nil {
		return nil, fmt.Errorf("find a habit [%s]: %w", hid, err)
	}
	deleteKey := h.GetKey()

	h.SK = fmt.Sprintf("ARCHIVED_HABITS#%s", hid)
	item, err := attributevalue.MarshalMap(h)
	if err != nil {
		return nil, fmt.Errorf("marshal habit: %w", err)
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
		return nil, fmt.Errorf("transact write items: %w", err)
	}

	h.beforeReturn()
	return h, nil
}

type DynamoRepositoryUpdateHabitInput struct {
	UserID    UserID
	HabitUUID uuid.UUID
	Title     string
}

func (r *DynamoRepository) UpdateHabit(ctx context.Context, in *DynamoRepositoryUpdateHabitInput) error {
	h := &DynamoHabit{
		PK: fmt.Sprintf("USER#%s", in.UserID),
		SK: fmt.Sprintf("HABITS#%s", in.HabitUUID),
	}

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

func (r *DynamoRepository) UnarchiveHabit(ctx context.Context, uid UserID, hid uuid.UUID) (*DynamoHabit, error) {
	h, err := r.FindArchivedHabit(ctx, uid, hid)
	if err != nil {
		return nil, fmt.Errorf("find a habit [%s]: %w", hid, err)
	}

	deleteKey := h.GetKey()

	h.SK = fmt.Sprintf("HABITS#%s", hid)
	item, err := attributevalue.MarshalMap(h)
	if err != nil {
		return nil, fmt.Errorf("marshal habit: %w", err)
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
		return nil, fmt.Errorf("transact write items: %w", err)
	}

	h.beforeReturn()
	return h, nil
}

func (r *DynamoRepository) ListLatestChecksWithLimit(ctx context.Context, uid UserID, hid uuid.UUID, limit int32) ([]*DynamoCheck, error) {
	expr, err := expression.NewBuilder().
		WithKeyCondition(
			expression.Key("PK").Equal(expression.Value(fmt.Sprintf("USER#%s", uid))).
				And(expression.Key("SK").BeginsWith(fmt.Sprintf("HABIT#%s__CHECK_DATE#", hid))),
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
		Limit:                     &limit,
		ScanIndexForward:          aws.Bool(false),
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

func (r *DynamoRepository) ListLastWeekChecksInAllHabits(ctx context.Context, uid UserID) ([]*DynamoCheck, error) {
	minTime := time.Now().Add(time.Hour * 24 * 7 * -1).UTC().Format("2006-01-02")

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

func (r *DynamoRepository) CreateCheck(ctx context.Context, uid UserID, hid uuid.UUID, date string) (*DynamoCheck, error) {
	now := time.Now()
	c := &DynamoCheck{
		PK:             fmt.Sprintf("USER#%s", uid),
		SK:             fmt.Sprintf("HABIT#%s__CHECK_DATE#%s", hid, date),
		CheckDateLSISK: fmt.Sprintf("CHECK_DATE#%s__HABIT#%s", date, hid),
		HabitUUID:      hid.String(),
		Date:           date,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	item, err := attributevalue.MarshalMap(c)
	if err != nil {
		return nil, fmt.Errorf("marshal check: %w", err)
	}

	h := &DynamoHabit{
		PK: fmt.Sprintf("USER#%s", uid),
		SK: fmt.Sprintf("HABITS#%s", hid),
	}

	update := expression.Set(expression.Name("ChecksCount"), expression.Name("ChecksCount").Plus(expression.Value(1)))
	updateExpr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		return nil, fmt.Errorf("build expression: %w", err)
	}

	if _, err := r.Client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			{
				Put: &types.Put{
					TableName: &r.TableName,
					Item:      item,
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
		return nil, fmt.Errorf("transact write items: %w", err)
	}
	return c, nil
}

func (r *DynamoRepository) DeleteCheck(ctx context.Context, uid UserID, hid uuid.UUID, date string) error {
	c := &DynamoCheck{
		PK: fmt.Sprintf("USER#%s", uid),
		SK: fmt.Sprintf("HABIT#%s__CHECK_DATE#%s", hid, date),
	}
	h := &DynamoHabit{
		PK: fmt.Sprintf("USER#%s", uid),
		SK: fmt.Sprintf("HABITS#%s", hid),
	}

	update := expression.Set(expression.Name("ChecksCount"), expression.Name("ChecksCount").Plus(expression.Value(-1)))
	updateExpr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		return fmt.Errorf("build expression: %w", err)
	}

	if _, err := r.Client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			{
				Delete: &types.Delete{
					TableName: &r.TableName,
					Key:       c.GetKey(),
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
		return fmt.Errorf("transact write items: %w", err)
	}
	return nil
}
