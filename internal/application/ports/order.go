package ports

import (
	"context"

	"github.com/13SOAT-andromeda/tech-challenge-s1/internal/adapter/database/model/order"
	"github.com/13SOAT-andromeda/tech-challenge-s1/internal/domain"
)

type OrderSearch struct {
	Status   string
	Enabled  bool
	OrderBy  string
	SortDesc bool
}

type CreateOrderInput struct {
	VehicleKilometers int
	Note              *string
	DiagnosticNote    *string
	CustomerVehicleID uint
	CompanyID         uint
}

type CreateCompleteOrderAnalysisInput struct {
	DiagnosticNote *string
	Products       []domain.StockItem
	Maintenances   []uint
}

type OrderRepository interface {
	Repository[order.Model]
	Search(ctx context.Context, params OrderSearch) ([]order.Model, error)
	FindOrderByID(ctx context.Context, id uint) (*order.Model, error)
}

type OrderService interface {
	Create(ctx context.Context, u domain.Order) (*domain.Order, error)
	GetAll(ctx context.Context, params map[string]interface{}) (*[]domain.Order, error)
	GetByID(ctx context.Context, id uint) (*domain.Order, error)
	Update(ctx context.Context, u domain.Order) error
	Delete(ctx context.Context, id uint) error
	GetApprovalTemplate(order domain.Order, customer domain.Customer, apiUrl string) (string, error)
}

type OrderUseCase interface {
	CreateOrder(ctx context.Context, userID uint, input CreateOrderInput) (*domain.Order, error)
	RequestApproval(ctx context.Context, id uint) error
	AssignOrder(ctx context.Context, orderID uint, userID uint) error
	CompleteOrderAnalysis(ctx context.Context, id uint, userID uint, input CreateCompleteOrderAnalysisInput) error
	ApproveOrder(ctx context.Context, id uint) error
	RejectOrder(ctx context.Context, id uint) error
	ArchiveOrder(ctx context.Context, id uint) error
	StartWorkOrder(ctx context.Context, id uint) error
	CompleteWorkOrder(ctx context.Context, id uint) error
}
