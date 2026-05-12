package metrics

import (
	"context"

	"github.com/DataDog/datadog-go/v5/statsd"

	"github.com/13SOAT-andromeda/tech-challenge-s1/internal/domain"
)

type OrderStatsd struct {
	client *statsd.Client
}

func NewOrderStatsd(client *statsd.Client) *OrderStatsd {
	return &OrderStatsd{client: client}
}

func (m *OrderStatsd) OrderCreated(ctx context.Context) {
	_ = ctx
	_ = m.client.Incr("order.created", nil, 1)
}

func (m *OrderStatsd) OrderStatusTransition(ctx context.Context, from, to domain.Status, durationMin float64) {
	_ = ctx
	tags := []string{
		"from:" + orderStatusTag(from),
		"to:" + orderStatusTag(to),
		"phase:" + orderPhaseForPreviousStatus(from),
	}
	_ = m.client.Distribution("order.status.transition.duration_min", durationMin, tags, 1)
}

func (m *OrderStatsd) OrderApproved(ctx context.Context) {
	_ = ctx
	_ = m.client.Incr("order.approved", nil, 1)
}

func (m *OrderStatsd) OrderRejected(ctx context.Context) {
	_ = ctx
	_ = m.client.Incr("order.rejected", nil, 1)
}
