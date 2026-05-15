package consumers

import (
	"context"
	"encoding/json"
	"fmt"

	sqsadapter "github.com/13SOAT-andromeda/tech-challenge-orders/internal/adapter/sqs"
	orderusecase "github.com/13SOAT-andromeda/tech-challenge-orders/internal/application/usecases/order"
)

// StockAvailable handles stock.available events.
func StockAvailable(svc *orderusecase.Service) sqsadapter.Handler {
	return func(ctx context.Context, eventID string, data json.RawMessage) error {
		var payload struct {
			OrderID string `json:"order_id"`
		}
		if err := json.Unmarshal(data, &payload); err != nil {
			return fmt.Errorf("stock.available: %w", err)
		}
		return svc.HandleStockAvailable(ctx, eventID, payload.OrderID)
	}
}

// StockUnavailable handles stock.unavailable events.
func StockUnavailable(svc *orderusecase.Service) sqsadapter.Handler {
	return func(ctx context.Context, eventID string, data json.RawMessage) error {
		var payload struct {
			OrderID string `json:"order_id"`
		}
		if err := json.Unmarshal(data, &payload); err != nil {
			return fmt.Errorf("stock.unavailable: %w", err)
		}
		return svc.HandleStockUnavailable(ctx, eventID, payload.OrderID)
	}
}

// StockUpdated handles stock.updated events.
func StockUpdated(svc *orderusecase.Service) sqsadapter.Handler {
	return func(ctx context.Context, eventID string, data json.RawMessage) error {
		var payload struct {
			OrderID string `json:"order_id"`
		}
		if err := json.Unmarshal(data, &payload); err != nil {
			return fmt.Errorf("stock.updated: %w", err)
		}
		return svc.HandleStockUpdated(ctx, eventID, payload.OrderID)
	}
}

// PaymentGenerated handles payment.generated events.
func PaymentGenerated(svc *orderusecase.Service) sqsadapter.Handler {
	return func(ctx context.Context, eventID string, data json.RawMessage) error {
		var payload struct {
			OrderID string `json:"order_id"`
		}
		if err := json.Unmarshal(data, &payload); err != nil {
			return fmt.Errorf("payment.generated: %w", err)
		}
		return svc.HandlePaymentGenerated(ctx, eventID, payload.OrderID)
	}
}

// PaymentApproved handles payment.approved events.
func PaymentApproved(svc *orderusecase.Service) sqsadapter.Handler {
	return func(ctx context.Context, eventID string, data json.RawMessage) error {
		var payload struct {
			OrderID string `json:"order_id"`
		}
		if err := json.Unmarshal(data, &payload); err != nil {
			return fmt.Errorf("payment.approved: %w", err)
		}
		return svc.HandlePaymentApproved(ctx, eventID, payload.OrderID)
	}
}

// PaymentFailed handles payment.failed events.
func PaymentFailed(svc *orderusecase.Service) sqsadapter.Handler {
	return func(ctx context.Context, eventID string, data json.RawMessage) error {
		var payload struct {
			OrderID string `json:"order_id"`
			Reason  string `json:"reason"`
		}
		if err := json.Unmarshal(data, &payload); err != nil {
			return fmt.Errorf("payment.failed: %w", err)
		}
		return svc.HandlePaymentFailed(ctx, eventID, payload.OrderID, payload.Reason)
	}
}
