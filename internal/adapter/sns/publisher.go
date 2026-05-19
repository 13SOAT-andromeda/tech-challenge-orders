package sns

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
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

// paymentsEnvelope matches the contract expected by the payments service.
type paymentsEnvelope struct {
	EventType string               `json:"event_type"`
	Payload   orderFinishedPayload `json:"payload"`
}

type orderFinishedPayload struct {
	OrderID       string           `json:"order_id"`
	CustomerID    string           `json:"customer_id"`
	CustomerEmail string           `json:"customer_email"`
	Amount        float64          `json:"amount"`
	Currency      string           `json:"currency"`
	Items         []paymentItemDTO `json:"items"`
}

type paymentItemDTO struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	Quantity  uint    `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

func toPaymentsPayload(e domain.OrderFinished) orderFinishedPayload {
	items := make([]paymentItemDTO, len(e.Items))
	for i, it := range e.Items {
		items[i] = paymentItemDTO{
			ID:        it.ID,
			Title:     it.Name,
			Quantity:  it.Quantity,
			UnitPrice: float64(it.UnitPriceCents) / 100,
		}
	}
	return orderFinishedPayload{
		OrderID:       e.OrderID,
		CustomerID:    strconv.FormatInt(e.CustomerID, 10),
		CustomerEmail: e.CustomerEmail,
		Amount:        float64(e.AmountCents) / 100,
		Currency:      "BRL",
		Items:         items,
	}
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

func marshalEvent(eventType string, event any) ([]byte, error) {
	if ev, ok := event.(domain.OrderFinished); ok {
		return json.Marshal(paymentsEnvelope{
			EventType: eventType,
			Payload:   toPaymentsPayload(ev),
		})
	}
	return json.Marshal(eventEnvelope{
		EventID:      uuid.New().String(),
		EventType:    eventType,
		EventVersion: "1",
		OccurredAt:   time.Now().UTC(),
		Data:         event,
	})
}

func (p *Publisher) Publish(ctx context.Context, topicARN string, event any) error {
	eventType, err := eventTypeOf(event)
	if err != nil {
		return err
	}

	body, err := marshalEvent(eventType, event)
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
