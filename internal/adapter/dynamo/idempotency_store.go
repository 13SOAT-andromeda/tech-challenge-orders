package dynamo

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/adapter/dynamo/model"
	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/application/ports"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const idempotencyTTL = 7 * 24 * time.Hour

// IdempotencyStore uses DynamoDB to deduplicate event processing.
// Items are stored in the same orders table with PK=PROCESSED#<event_id>.
type IdempotencyStore struct {
	client    *dynamodb.Client
	tableName string
}

func NewIdempotencyStore(client *dynamodb.Client, tableName string) *IdempotencyStore {
	return &IdempotencyStore{client: client, tableName: tableName}
}

// MarkProcessed writes a record for eventID with a 7-day TTL.
// Returns (true, nil) on first call; (false, nil) if already recorded.
func (s *IdempotencyStore) MarkProcessed(ctx context.Context, eventID string) (bool, error) {
	pk := fmt.Sprintf("PROCESSED#%s", eventID)
	expiresAt := time.Now().Add(idempotencyTTL).Unix()

	_, err := s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName),
		Item: map[string]ddbtypes.AttributeValue{
			"PK":        &ddbtypes.AttributeValueMemberS{Value: pk},
			"SK":        &ddbtypes.AttributeValueMemberS{Value: model.SKMeta},
			"ExpiresAt": &ddbtypes.AttributeValueMemberN{Value: strconv.FormatInt(expiresAt, 10)},
		},
		ConditionExpression: aws.String("attribute_not_exists(PK)"),
	})
	if err != nil {
		var ccf *ddbtypes.ConditionalCheckFailedException
		if errors.As(err, &ccf) {
			return false, nil // already processed — caller should skip
		}
		return false, fmt.Errorf("idempotency mark processed: %w", err)
	}
	return true, nil
}

var _ ports.IdempotencyStore = (*IdempotencyStore)(nil)
