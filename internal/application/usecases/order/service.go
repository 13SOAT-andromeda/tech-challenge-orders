package order

import (
	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/application/ports"
	orderport "github.com/13SOAT-andromeda/tech-challenge-orders/internal/application/ports/order"
)

type Service struct {
	repo          orderport.Repository
	publisher     ports.EventPublisher
	users         ports.UsersClient
	stock         ports.StockClient
	idempotency   ports.IdempotencyStore
	topicARN      string
	publicBaseURL string
}

func NewService(
	repo orderport.Repository,
	publisher ports.EventPublisher,
	users ports.UsersClient,
	stock ports.StockClient,
	idempotency ports.IdempotencyStore,
	topicARN string,
	publicBaseURL string,
) *Service {
	return &Service{
		repo:          repo,
		publisher:     publisher,
		users:         users,
		stock:         stock,
		idempotency:   idempotency,
		topicARN:      topicARN,
		publicBaseURL: publicBaseURL,
	}
}

var _ ports.OrderUseCase = (*Service)(nil)
