package order

import (
	"context"
	"fmt"
	"time"

	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/application/ports"
	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/domain"
	"github.com/google/uuid"
)

func (s *Service) CreateOrder(ctx context.Context, userID int64, input ports.CreateOrderInput) (*domain.Order, error) {
	cv, err := s.users.GetCustomerVehicle(ctx, input.CustomerVehicleID)
	if err != nil {
		return nil, fmt.Errorf("get customer vehicle: %w", err)
	}

	now := time.Now()
	order := &domain.Order{
		ID:                uuid.New().String(),
		Status:            domain.RECEIVED,
		DateIn:            now,
		CreatedAt:         now,
		UpdatedAt:         now,
		VehicleKilometers: input.VehicleKilometers,
		Note:              input.Note,
		CustomerVehicleID: input.CustomerVehicleID,
		CustomerID:        cv.CustomerID,
		CompanyID:         input.CompanyID,
		Vehicle:           &cv.Vehicle,
	}

	if err := s.repo.Save(ctx, order); err != nil {
		return nil, fmt.Errorf("save order: %w", err)
	}
	return order, nil
}
