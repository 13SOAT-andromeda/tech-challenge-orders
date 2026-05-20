package order

import (
	"time"

	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/application/ports"
	orderport "github.com/13SOAT-andromeda/tech-challenge-orders/internal/application/ports/order"
	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/domain"
)

type Service struct {
	repo          orderport.Repository
	publisher     ports.EventPublisher
	notifPublisher ports.NotificationPublisher
	users         ports.UsersClient
	stock         ports.StockClient
	idempotency   ports.IdempotencyStore
	metrics       ports.OrderMetrics
	topicARN      string
	publicBaseURL string
}

func NewService(
	repo orderport.Repository,
	publisher ports.EventPublisher,
	notifPublisher ports.NotificationPublisher,
	users ports.UsersClient,
	stock ports.StockClient,
	idempotency ports.IdempotencyStore,
	metrics ports.OrderMetrics,
	topicARN string,
	publicBaseURL string,
) *Service {
	return &Service{
		repo:          repo,
		publisher:     publisher,
		notifPublisher: notifPublisher,
		users:         users,
		stock:         stock,
		idempotency:   idempotency,
		metrics:       metrics,
		topicARN:      topicARN,
		publicBaseURL: publicBaseURL,
	}
}

// lastTransitionDuration returns the from/to statuses and how long the order
// spent in the previous status, based on the last history entry.
func lastTransitionDuration(o *domain.Order) (from, to domain.Status, durationMin float64) {
	n := len(o.History)
	last := o.History[n-1]
	var prevAt time.Time
	if n >= 2 {
		prevAt = o.History[n-2].At
	} else {
		prevAt = o.DateIn
	}
	return last.From, last.To, last.At.Sub(prevAt).Minutes()
}

var _ ports.OrderUseCase = (*Service)(nil)
