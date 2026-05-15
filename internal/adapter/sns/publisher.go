package sns

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/domain"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	snstypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/google/uuid"
)

type Publisher struct {
	client *sns.Client
}

func NewPublisher(client *sns.Client) *Publisher {
	return &Publisher{client: client}
}

type eventEnvelope struct {
	EventID      string    `json:"event_id"`
	EventType    string    `json:"event_type"`
	EventVersion string    `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	Data         any       `json:"data"`
}

func eventTypeOf(event any) (string, error) {
	switch event.(type) {
	case domain.OrderApprovalRequested:
		return "order.approval-requested", nil
	case domain.OrderApproved:
		return "order.approved", nil
	case domain.OrderFinished:
		return "order.finished", nil
	default:
		return "", fmt.Errorf("unknown event type: %T", event)
	}
}

func (p *Publisher) Publish(ctx context.Context, topicARN string, event any) error {
	eventType, err := eventTypeOf(event)
	if err != nil {
		return err
	}

	envelope := eventEnvelope{
		EventID:      uuid.New().String(),
		EventType:    eventType,
		EventVersion: "1",
		OccurredAt:   time.Now().UTC(),
		Data:         event,
	}

	body, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	_, err = p.client.Publish(ctx, &sns.PublishInput{
		TopicArn: aws.String(topicARN),
		Message:  aws.String(string(body)),
		MessageAttributes: map[string]snstypes.MessageAttributeValue{
			"event_type": {
				DataType:    aws.String("String"),
				StringValue: aws.String(eventType),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("sns publish %s: %w", eventType, err)
	}
	return nil
}
