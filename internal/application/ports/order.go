package ports

import (
	"context"

	orderport "github.com/13SOAT-andromeda/tech-challenge-orders/internal/application/ports/order"
	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/domain"
)

type CreateOrderInput struct {
	VehicleKilometers int
	Note              *string
	CustomerVehicleID int64
	EmployeeID        int64
	CompanyID         int64
}

type CreateCompleteOrderAnalysisInput struct {
	DiagnosticNote *string
	Products       []domain.StockItem
	Maintenances   []int64
}

type OrderDetail struct {
	domain.Order
	Customer *Customer
	Employee *Employee
	Company  *Company
}

// OrderUseCase is the interface consumed by HTTP handlers.
// Implementations live in internal/application/usecases/order/.
type OrderUseCase interface {
	CreateOrder(ctx context.Context, userID int64, input CreateOrderInput) (*domain.Order, error)
	AssignOrder(ctx context.Context, orderID string, userID int64) error
	CompleteOrderAnalysis(ctx context.Context, id string, userID int64, input CreateCompleteOrderAnalysisInput) error
	RequestApproval(ctx context.Context, id string) error
	ApproveOrder(ctx context.Context, id string) error
	RejectOrder(ctx context.Context, id string) error
	ArchiveOrder(ctx context.Context, id string) error
	CompleteWorkOrder(ctx context.Context, id string) error
	GetAllOrders(ctx context.Context, page orderport.Page) (orderport.PageResult, error)
	GetOrderByID(ctx context.Context, id string) (*OrderDetail, error)
	GetInProgressOrders(ctx context.Context, page orderport.Page) (orderport.PageResult, error)
}
