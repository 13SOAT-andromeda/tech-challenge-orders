package order

import (
	"context"

	orderport "github.com/13SOAT-andromeda/tech-challenge-orders/internal/application/ports/order"
	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/domain"
)

func (s *Service) GetInProgressOrders(ctx context.Context, page orderport.Page) (orderport.PageResult, error) {
	return s.repo.ListByStatus(ctx, domain.IN_PROGRESS, page)
}
