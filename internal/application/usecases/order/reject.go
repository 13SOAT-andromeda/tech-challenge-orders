package order

import (
	"context"
	"fmt"
)

func (s *Service) RejectOrder(ctx context.Context, orderID string) error {
	order, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		return err
	}
	if err := order.Reject(); err != nil {
		return err
	}
	if err := s.repo.Save(ctx, order); err != nil {
		return fmt.Errorf("save order: %w", err)
	}
	return nil
}
