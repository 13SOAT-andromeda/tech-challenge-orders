package ports

import (
	"context"

	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/domain"
)

type CustomerVehicle struct {
	ID         string
	CustomerID string
	Vehicle    domain.VehicleSnapshot
}

type Customer struct {
	ID    string
	Name  string
	Email string
}

type User struct {
	ID    string
	Email string
	Role  string
}

type UsersClient interface {
	GetCustomerVehicle(ctx context.Context, id string) (*CustomerVehicle, error)
	GetCustomer(ctx context.Context, id string) (*Customer, error)
	GetUser(ctx context.Context, id string) (*User, error)
	GetMaintenancesBatch(ctx context.Context, ids []string) ([]domain.ItemSnapshot, error)
}
