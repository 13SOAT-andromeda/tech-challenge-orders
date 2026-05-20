package order

import (
	"context"
	"fmt"
)

func (s *Service) ArchiveOrder(ctx context.Context, orderID string) error {
	order, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		return err
	}
	if err := order.Archive("system"); err != nil {
		return err
	}
	if err := s.repo.Save(ctx, order); err != nil {
		return fmt.Errorf("save order: %w", err)
	}
	from, to, dur := lastTransitionDuration(order)
	s.metrics.OrderStatusTransition(ctx, from, to, dur)
	return nil
}
