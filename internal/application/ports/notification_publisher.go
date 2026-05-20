package ports

import (
	"context"

	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/domain"
)

type NotificationPublisher interface {
	PublishApprovalRequest(ctx context.Context, evt domain.OrderApprovalRequested) error
	PublishAwaitingPayment(ctx context.Context, order *domain.Order, customerEmail, customerName, paymentURL string) error
}
