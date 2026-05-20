package order

import (
	"context"
	"fmt"
	"time"

	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/domain"
)

func (s *Service) RequestApproval(ctx context.Context, orderID string) error {
	order, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		return err
	}

	customer, err := s.users.GetCustomer(ctx, order.CustomerID)
	if err != nil {
		return fmt.Errorf("get customer: %w", err)
	}

	if err := order.RequestApproval("system"); err != nil {
		return err
	}
	if err := s.repo.Save(ctx, order); err != nil {
		return fmt.Errorf("save order: %w", err)
	}
	from, to, dur := lastTransitionDuration(order)
	s.metrics.OrderStatusTransition(ctx, from, to, dur)

	vehicle := domain.VehicleSnapshot{}
	if order.Vehicle != nil {
		vehicle = *order.Vehicle
	}

	diagnosticNote := ""
	if order.DiagnosticNote != nil {
		diagnosticNote = *order.DiagnosticNote
	}

	priceCents := int64(0)
	if order.PriceCents != nil {
		priceCents = *order.PriceCents
	}

	evt := domain.OrderApprovalRequested{
		OrderID:        order.ID,
		CustomerID:     customer.ID,
		CustomerName:   customer.Name,
		CustomerEmail:  customer.Email,
		Vehicle:        vehicle,
		DiagnosticNote: diagnosticNote,
		Items:          order.Items,
		TotalCents:     priceCents,
		ApprovalURL:    s.publicBaseURL + "/api/orders/" + order.ID + "/approve",
		RejectURL:      s.publicBaseURL + "/api/orders/" + order.ID + "/reject",
		ExpiresAt:      time.Now().Add(7 * 24 * time.Hour),
	}

	return s.publisher.Publish(ctx, s.topicARN, evt)
}
