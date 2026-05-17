package repository

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/adapter/dynamo/model"
	orderport "github.com/13SOAT-andromeda/tech-challenge-orders/internal/application/ports/order"
	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/domain"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// ErrConcurrencyConflict is returned when a concurrent write modified the item
// between our read and the attempted PutItem.
var ErrConcurrencyConflict = errors.New("optimistic concurrency conflict")

const (
	gsiStatus          = "GSI1"
	gsiCustomerVehicle = "GSI2"
	defaultPageSize    = 50
)

type OrderRepository struct {
	client    *dynamodb.Client
	tableName string
}

func NewOrderRepository(client *dynamodb.Client, tableName string) *OrderRepository {
	return &OrderRepository{client: client, tableName: tableName}
}

// Save persists the full order aggregate.
// Optimistic concurrency: on first save (Version=0) it requires the PK to be absent;
// on subsequent saves it requires the stored Version to equal order.Version-1.
func (r *OrderRepository) Save(ctx context.Context, order *domain.Order) error {
	item := model.FromDomain(order)
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("marshal order: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      av,
		// Allow creation when PK is new; require the stored Version to be
		// order.Version-1 for updates (the domain increments Version before Save).
		ConditionExpression: aws.String(
			"attribute_not_exists(PK) OR #ver = :v_prev",
		),
		ExpressionAttributeNames: map[string]string{
			"#ver": "Version",
		},
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":v_prev": &ddbtypes.AttributeValueMemberN{
				Value: strconv.Itoa(order.Version - 1),
			},
		},
	})
	if err != nil {
		var ccf *ddbtypes.ConditionalCheckFailedException
		if errors.As(err, &ccf) {
			return fmt.Errorf("%w: order %s", ErrConcurrencyConflict, order.ID)
		}
		return fmt.Errorf("put order: %w", err)
	}
	return nil
}

func (r *OrderRepository) FindByID(ctx context.Context, id string) (*domain.Order, error) {
	pk, sk := model.PrimaryKey(id)

	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]ddbtypes.AttributeValue{
			"PK": &ddbtypes.AttributeValueMemberS{Value: pk},
			"SK": &ddbtypes.AttributeValueMemberS{Value: sk},
		},
		ConsistentRead: aws.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}

	if len(out.Item) == 0 {
		return nil, domain.ErrOrderNotFound
	}

	var item model.OrderItem
	if err := attributevalue.UnmarshalMap(out.Item, &item); err != nil {
		return nil, fmt.Errorf("unmarshal order: %w", err)
	}

	return item.ToDomain(), nil
}

func (r *OrderRepository) ListAll(ctx context.Context, page orderport.Page) (orderport.PageResult, error) {
	limit := page.Limit
	if limit <= 0 {
		limit = defaultPageSize
	}

	startKey, err := decodeCursor(page.Cursor)
	if err != nil {
		return orderport.PageResult{}, fmt.Errorf("decode cursor: %w", err)
	}

	out, err := r.client.Scan(ctx, &dynamodb.ScanInput{
		TableName:         aws.String(r.tableName),
		Limit:             aws.Int32(limit),
		ExclusiveStartKey: startKey,
		FilterExpression:  aws.String("begins_with(PK, :prefix)"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":prefix": &ddbtypes.AttributeValueMemberS{Value: "ORDER#"},
		},
	})
	if err != nil {
		return orderport.PageResult{}, fmt.Errorf("scan orders: %w", err)
	}

	orders := make([]domain.Order, 0, len(out.Items))
	for _, raw := range out.Items {
		var item model.OrderItem
		if err := attributevalue.UnmarshalMap(raw, &item); err != nil {
			return orderport.PageResult{}, fmt.Errorf("unmarshal order: %w", err)
		}
		orders = append(orders, *item.ToDomain())
	}

	next, err := encodeCursor(out.LastEvaluatedKey)
	if err != nil {
		return orderport.PageResult{}, fmt.Errorf("encode cursor: %w", err)
	}

	return orderport.PageResult{Orders: orders, NextCursor: next}, nil
}

func (r *OrderRepository) ListByStatus(ctx context.Context, status domain.Status, page orderport.Page) (orderport.PageResult, error) {
	return r.queryGSI(ctx, gsiStatus, "GSI1PK", fmt.Sprintf("STATUS#%s", status), page)
}

func (r *OrderRepository) ListByCustomerVehicle(ctx context.Context, customerVehicleID string, page orderport.Page) (orderport.PageResult, error) {
	return r.queryGSI(ctx, gsiCustomerVehicle, "GSI2PK", fmt.Sprintf("CUSTOMERVEHICLE#%s", customerVehicleID), page)
}

func (r *OrderRepository) queryGSI(ctx context.Context, indexName, pkAttr, pkValue string, page orderport.Page) (orderport.PageResult, error) {
	limit := page.Limit
	if limit <= 0 {
		limit = defaultPageSize
	}

	startKey, err := decodeCursor(page.Cursor)
	if err != nil {
		return orderport.PageResult{}, fmt.Errorf("decode cursor: %w", err)
	}

	out, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String(indexName),
		KeyConditionExpression: aws.String("#pk = :pk"),
		ExpressionAttributeNames: map[string]string{
			"#pk": pkAttr,
		},
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":pk": &ddbtypes.AttributeValueMemberS{Value: pkValue},
		},
		Limit:             aws.Int32(limit),
		ExclusiveStartKey: startKey,
		ScanIndexForward:  aws.Bool(false),
	})
	if err != nil {
		return orderport.PageResult{}, fmt.Errorf("query %s: %w", indexName, err)
	}

	orders := make([]domain.Order, 0, len(out.Items))
	for _, raw := range out.Items {
		var item model.OrderItem
		if err := attributevalue.UnmarshalMap(raw, &item); err != nil {
			return orderport.PageResult{}, fmt.Errorf("unmarshal order: %w", err)
		}
		orders = append(orders, *item.ToDomain())
	}

	next, err := encodeCursor(out.LastEvaluatedKey)
	if err != nil {
		return orderport.PageResult{}, fmt.Errorf("encode cursor: %w", err)
	}

	return orderport.PageResult{Orders: orders, NextCursor: next}, nil
}

func encodeCursor(key map[string]ddbtypes.AttributeValue) (string, error) {
	if len(key) == 0 {
		return "", nil
	}
	plain := make(map[string]any, len(key))
	if err := attributevalue.UnmarshalMap(key, &plain); err != nil {
		return "", err
	}
	b, err := json.Marshal(plain)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func decodeCursor(cursor string) (map[string]ddbtypes.AttributeValue, error) {
	if cursor == "" {
		return nil, nil
	}
	b, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return nil, err
	}
	var plain map[string]any
	if err := json.Unmarshal(b, &plain); err != nil {
		return nil, err
	}
	return attributevalue.MarshalMap(plain)
}

var _ orderport.Repository = (*OrderRepository)(nil)
