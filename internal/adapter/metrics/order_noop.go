package metrics

import (
	"context"

	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/domain"
)

type NoopOrderMetrics struct{}

func (NoopOrderMetrics) OrderCreated(context.Context) {}

func (NoopOrderMetrics) OrderStatusTransition(context.Context, domain.Status, domain.Status, float64) {
}

func (NoopOrderMetrics) OrderApproved(context.Context) {}

func (NoopOrderMetrics) OrderRejected(context.Context) {}
