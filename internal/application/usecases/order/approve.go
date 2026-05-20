package order

import (
	"context"
	"fmt"

	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/domain"
)

func (s *Service) ApproveOrder(ctx context.Context, orderID string) error {
	order, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		return err
	}

	customer, err := s.users.GetCustomer(ctx, order.CustomerID)
	if err != nil {
		return fmt.Errorf("get customer: %w", err)
	}

	if err := order.Approve(); err != nil {
		return err
	}
	if err := s.repo.Save(ctx, order); err != nil {
		return fmt.Errorf("save order: %w", err)
	}

	productItems := make([]domain.ItemSnapshot, 0)
	for _, item := range order.Items {
		if item.Kind == domain.ItemKindProduct {
			productItems = append(productItems, item)
		}
	}

	approvedAt := order.DateIn
	if order.DateApproved != nil {
		approvedAt = *order.DateApproved
	}

	evt := domain.OrderApproved{
		OrderID:       order.ID,
		CustomerName:  customer.Name,
		CustomerEmail: customer.Email,
		Items:         productItems,
		ApprovedAt:    approvedAt,
	}

	return s.publisher.Publish(ctx, s.topicARN, evt)
}
