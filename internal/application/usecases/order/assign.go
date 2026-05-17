package order

import (
	"context"
	"fmt"
	"strconv"
)

func (s *Service) AssignOrder(ctx context.Context, orderID string, userID int64) error {
	order, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		return err
	}
	if err := order.Assign(strconv.FormatInt(userID, 10)); err != nil {
		return err
	}
	order.EmployeeID = userID
	if err := s.repo.Save(ctx, order); err != nil {
		return fmt.Errorf("save order: %w", err)
	}
	return nil
}
