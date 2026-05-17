package ports

import (
	"context"

	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/domain"
)

type CustomerVehicle struct {
	ID         int64
	CustomerID int64
	VehicleID  int64
	Vehicle    domain.VehicleSnapshot
	Customer   *Customer
}

type Customer struct {
	ID    int64
	Name  string
	Email string
}

type User struct {
	ID    int64
	Email string
	Role  string
}

type UsersClient interface {
	GetCustomerVehicle(ctx context.Context, id int64) (*CustomerVehicle, error)
	GetCustomer(ctx context.Context, id int64) (*Customer, error)
	GetUser(ctx context.Context, id int64) (*User, error)
	GetMaintenancesBatch(ctx context.Context, ids []int64) ([]domain.ItemSnapshot, error)
}
