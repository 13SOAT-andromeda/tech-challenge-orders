package order

import (
	"context"
	"fmt"
	"strconv"
)

func (s *Service) AssignOrder(ctx context.Context, orderID string, userID int64) error {
	employee, err := s.users.GetEmployeeByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get employee by user: %w", err)
	}

	order, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		return err
	}
	if err := order.Assign(strconv.FormatInt(userID, 10)); err != nil {
		return err
	}
	order.EmployeeID = employee.ID
	if err := s.repo.Save(ctx, order); err != nil {
		return fmt.Errorf("save order: %w", err)
	}
	from, to, dur := lastTransitionDuration(order)
	s.metrics.OrderStatusTransition(ctx, from, to, dur)
	return nil
}
