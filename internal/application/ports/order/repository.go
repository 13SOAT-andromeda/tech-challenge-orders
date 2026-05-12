package order

import (
	"context"

	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/domain"
)

type Page struct {
	Limit  int32
	Cursor string
}

type PageResult struct {
	Orders     []domain.Order
	NextCursor string
}

type Repository interface {
	Save(ctx context.Context, order *domain.Order) error
	FindByID(ctx context.Context, id string) (*domain.Order, error)
	ListByStatus(ctx context.Context, status domain.Status, page Page) (PageResult, error)
	ListByCustomerVehicle(ctx context.Context, customerVehicleID string, page Page) (PageResult, error)
}
