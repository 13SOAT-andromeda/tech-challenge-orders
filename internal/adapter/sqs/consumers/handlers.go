package consumers

import (
	"context"
	"encoding/json"
	"fmt"

	sqsadapter "github.com/13SOAT-andromeda/tech-challenge-orders/internal/adapter/sqs"
	orderusecase "github.com/13SOAT-andromeda/tech-challenge-orders/internal/application/usecases/order"
)

func StockAvailable(svc *orderusecase.Service) sqsadapter.Handler {
	return func(ctx context.Context, _, eventID string, data json.RawMessage) error {
		var payload struct {
			OrderID string `json:"order_id"`
		}
		if err := json.Unmarshal(data, &payload); err != nil {
			return fmt.Errorf("stock.available: %w", err)
		}
		return svc.HandleStockAvailable(ctx, eventID, payload.OrderID)
	}
}

func StockUnavailable(svc *orderusecase.Service) sqsadapter.Handler {
	return func(ctx context.Context, _, eventID string, data json.RawMessage) error {
		var payload struct {
			OrderID string `json:"order_id"`
		}
		if err := json.Unmarshal(data, &payload); err != nil {
			return fmt.Errorf("stock.unavailable: %w", err)
		}
		return svc.HandleStockUnavailable(ctx, eventID, payload.OrderID)
	}
}

func StockUpdated(svc *orderusecase.Service) sqsadapter.Handler {
	return func(ctx context.Context, _, eventID string, data json.RawMessage) error {
		var payload struct {
			OrderID string `json:"order_id"`
		}
		if err := json.Unmarshal(data, &payload); err != nil {
			return fmt.Errorf("stock.updated: %w", err)
		}
		return svc.HandleStockUpdated(ctx, eventID, payload.OrderID)
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
