package ports

import (
	"context"

	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/domain"
)

type CreateOrderInput struct {
	VehicleKilometers int
	Note              *string
	CustomerVehicleID string
	CompanyID         string
}

type CreateCompleteOrderAnalysisInput struct {
	DiagnosticNote *string
	Products       []domain.StockItem
	Maintenances   []string
}

// OrderUseCase is the interface consumed by HTTP handlers.
// Implementations live in internal/application/usecases/order/.
type OrderUseCase interface {
	CreateOrder(ctx context.Context, userID string, input CreateOrderInput) (*domain.Order, error)
	AssignOrder(ctx context.Context, orderID string, userID string) error
	CompleteOrderAnalysis(ctx context.Context, id string, userID string, input CreateCompleteOrderAnalysisInput) error
	RequestApproval(ctx context.Context, id string) error
	ApproveOrder(ctx context.Context, id string) error
	RejectOrder(ctx context.Context, id string) error
	ArchiveOrder(ctx context.Context, id string) error
	CompleteWorkOrder(ctx context.Context, id string) error
}
