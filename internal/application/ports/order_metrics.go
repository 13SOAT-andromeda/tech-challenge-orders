package ports

import (
	"context"

	"github.com/13SOAT-andromeda/tech-challenge-s1/internal/domain"
)

type OrderMetrics interface {
	OrderCreated(ctx context.Context)
	OrderStatusTransition(ctx context.Context, from, to domain.Status, durationMin float64)
	OrderApproved(ctx context.Context)
	OrderRejected(ctx context.Context)
}
