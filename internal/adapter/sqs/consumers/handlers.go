package consumers

import (
	"context"
	"encoding/json"
	"fmt"

	sqsadapter "github.com/13SOAT-andromeda/tech-challenge-orders/internal/adapter/sqs"
	orderusecase "github.com/13SOAT-andromeda/tech-challenge-orders/internal/application/usecases/order"
)

type catalogPayload struct {
	OrderID string `json:"order_id"`
}

// CatalogEvents routes STOCK_RESERVED, STOCK_RESERVATION_FAILED and BACKORDER_CREATED
// events from the single catalog-events SNS topic to the appropriate use case.
func CatalogEvents(svc *orderusecase.Service) sqsadapter.Handler {
	return func(ctx context.Context, eventType, _ string, data json.RawMessage) error {
		var p catalogPayload
		if err := json.Unmarshal(data, &p); err != nil {
			return fmt.Errorf("%s: %w", eventType, err)
		}
		switch eventType {
		case "STOCK_RESERVED":
			return svc.HandleStockAvailable(ctx, p.OrderID+"#STOCK_RESERVED", p.OrderID)
		case "STOCK_RESERVATION_FAILED":
			return svc.HandleStockUnavailable(ctx, p.OrderID+"#STOCK_RESERVATION_FAILED", p.OrderID)
		case "BACKORDER_CREATED":
			return svc.HandleStockUpdated(ctx, p.OrderID+"#BACKORDER_CREATED", p.OrderID)
		default:
			return fmt.Errorf("unknown catalog event type: %s", eventType)
		}
	}
}

// PaymentEvents routes payment.checkout_created, payment.approved and payment.failed
// from the single payments-events SNS topic to the appropriate use case.
func PaymentEvents(svc *orderusecase.Service) sqsadapter.Handler {
	return func(ctx context.Context, eventType, eventID string, data json.RawMessage) error {
		switch eventType {
		case "payment.checkout_created":
			var p struct {
				OrderID    string `json:"order_id"`
				PaymentURL string `json:"payment_url"`
			}
			if err := json.Unmarshal(data, &p); err != nil {
				return fmt.Errorf("payment.checkout_created: %w", err)
			}
			return svc.HandlePaymentCheckoutCreated(ctx, eventID, p.OrderID, p.PaymentURL)

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
