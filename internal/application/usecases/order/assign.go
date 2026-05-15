package order

import (
	"context"
	"fmt"
)

func (s *Service) AssignOrder(ctx context.Context, orderID string, userID string) error {
	order, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		return err
	}
	if err := order.Assign(userID); err != nil {
		return err
	}
	order.EmployeeID = userID
	if err := s.repo.Save(ctx, order); err != nil {
		return fmt.Errorf("save order: %w", err)
	}
	return nil
}
