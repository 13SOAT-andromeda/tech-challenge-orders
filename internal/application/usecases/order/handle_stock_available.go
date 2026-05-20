package order

import (
	"context"
	"fmt"
)

func (s *Service) HandleStockAvailable(ctx context.Context, eventID, orderID string) error {
	ok, err := s.idempotency.MarkProcessed(ctx, eventID)
	if err != nil {
		return fmt.Errorf("idempotency check: %w", err)
	}
	if !ok {
		return nil
	}

	order, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		return err
	}
	if err := order.MarkStockAvailable(); err != nil {
		return err
	}
	if err := s.repo.Save(ctx, order); err != nil {
		return fmt.Errorf("save order: %w", err)
	}
	from, to, dur := lastTransitionDuration(order)
	s.metrics.OrderStatusTransition(ctx, from, to, dur)
	return nil
}
