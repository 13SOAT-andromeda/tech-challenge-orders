package order

import (
	"context"
	"fmt"
	"time"

	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/domain"
)

func (s *Service) CompleteWorkOrder(ctx context.Context, orderID string) error {
	order, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		return err
	}

	customer, err := s.users.GetCustomer(ctx, order.CustomerID)
	if err != nil {
		return fmt.Errorf("get customer: %w", err)
	}

	if err := order.CompleteWork("system"); err != nil {
		return err
	}
	if err := s.repo.Save(ctx, order); err != nil {
		return fmt.Errorf("save order: %w", err)
	}

	priceCents := int64(0)
	if order.PriceCents != nil {
		priceCents = *order.PriceCents
	}

	evt := domain.OrderFinished{
		OrderID:       order.ID,
		CustomerID:    customer.ID,
		CustomerEmail: customer.Email,
		AmountCents:   priceCents,
		Items:         order.Items,
		FinishedAt:    time.Now(),
	}

	return s.publisher.Publish(ctx, s.topicARN, evt)
}
