package consumers

import (
	"context"
	"encoding/json"
	"fmt"

	sqsadapter "github.com/13SOAT-andromeda/tech-challenge-orders/internal/adapter/sqs"
	orderusecase "github.com/13SOAT-andromeda/tech-challenge-orders/internal/application/usecases/order"
)

// stockPayload is the raw JSON published by the catalog API.
// It has no event envelope, so order_id+type is used as the idempotency key.
type stockPayload struct {
	OrderID string `json:"order_id"`
	Type    string `json:"type"`
}

// StockReserved handles STOCK_RESERVED events from catalog-stock-reserved-topic.
func StockReserved(svc *orderusecase.Service) sqsadapter.Handler {
	return func(ctx context.Context, _, _ string, data json.RawMessage) error {
		var p stockPayload
		if err := json.Unmarshal(data, &p); err != nil {
			return fmt.Errorf("STOCK_RESERVED: %w", err)
		}
		return svc.HandleStockAvailable(ctx, p.OrderID+"#STOCK_RESERVED", p.OrderID)
	}
}

// StockInsufficient handles STOCK_INSUFFICIENT events from catalog-stock-insufficient-topic.
func StockInsufficient(svc *orderusecase.Service) sqsadapter.Handler {
	return func(ctx context.Context, _, _ string, data json.RawMessage) error {
		var p stockPayload
		if err := json.Unmarshal(data, &p); err != nil {
			return fmt.Errorf("STOCK_INSUFFICIENT: %w", err)
		}
		return svc.HandleStockUnavailable(ctx, p.OrderID+"#STOCK_INSUFFICIENT", p.OrderID)
	}
}

// BackorderCreated handles BACKORDER_CREATED events from catalog-backorder-created-topic.
func BackorderCreated(svc *orderusecase.Service) sqsadapter.Handler {
	return func(ctx context.Context, _, _ string, data json.RawMessage) error {
		var p stockPayload
		if err := json.Unmarshal(data, &p); err != nil {
			return fmt.Errorf("BACKORDER_CREATED: %w", err)
		}
		return svc.HandleStockUpdated(ctx, p.OrderID+"#BACKORDER_CREATED", p.OrderID)
	}
}

// PaymentEvents routes payment.checkout_created, payment.approved and payment.failed
// from the single payments-events SNS topic to the appropriate use case.
func PaymentEvents(svc *orderusecase.Service) sqsadapter.Handler {
	return func(ctx context.Context, eventType, eventID string, data json.RawMessage) error {
		switch eventType {
		case "payment.checkout_created":
			var p struct {
				OrderID string `json:"order_id"`
			}
			if err := json.Unmarshal(data, &p); err != nil {
				return fmt.Errorf("payment.checkout_created: %w", err)
			}
			return svc.HandlePaymentCheckoutCreated(ctx, eventID, p.OrderID)

		case "payment.approved":
			var p struct {
				OrderID string `json:"order_id"`
			}
			if err := json.Unmarshal(data, &p); err != nil {
				return fmt.Errorf("payment.approved: %w", err)
			}
			return svc.HandlePaymentApproved(ctx, eventID, p.OrderID)

		case "payment.failed":
			var p struct {
				OrderID string `json:"order_id"`
				Reason  string `json:"reason"`
			}
			if err := json.Unmarshal(data, &p); err != nil {
				return fmt.Errorf("payment.failed: %w", err)
			}
			return svc.HandlePaymentFailed(ctx, eventID, p.OrderID, p.Reason)

		default:
			return fmt.Errorf("unknown payment event type: %s", eventType)
		}
	}
}
