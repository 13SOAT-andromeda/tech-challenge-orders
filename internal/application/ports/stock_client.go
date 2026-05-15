package ports

import (
	"context"

	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/domain"
)

type StockClient interface {
	CheckProductsBatch(ctx context.Context, items []domain.StockItem) ([]domain.ItemSnapshot, error)
}
