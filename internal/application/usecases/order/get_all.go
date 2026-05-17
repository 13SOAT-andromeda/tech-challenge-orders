package order

import (
	"context"

	orderport "github.com/13SOAT-andromeda/tech-challenge-orders/internal/application/ports/order"
)

func (s *Service) GetAllOrders(ctx context.Context, page orderport.Page) (orderport.PageResult, error) {
	return s.repo.ListAll(ctx, page)
}
