package order

import (
	"context"
	"fmt"
	"sync"

	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/application/ports"
	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/domain"
)

func (s *Service) CompleteOrderAnalysis(ctx context.Context, id string, userID string, input ports.CreateCompleteOrderAnalysisInput) error {
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	var (
		productItems []domain.ItemSnapshot
		maintItems   []domain.ItemSnapshot
		productErr   error
		maintErr     error
		wg           sync.WaitGroup
	)

	wg.Add(2)
	go func() {
		defer wg.Done()
		productItems, productErr = s.stock.CheckProductsBatch(ctx, input.Products)
	}()
	go func() {
		defer wg.Done()
		maintItems, maintErr = s.users.GetMaintenancesBatch(ctx, input.Maintenances)
	}()
	wg.Wait()

	if productErr != nil {
		return fmt.Errorf("check products: %w", productErr)
	}
	if maintErr != nil {
		return fmt.Errorf("get maintenances: %w", maintErr)
	}

	items := append(productItems, maintItems...)

	if err := order.CompleteAnalysis(userID, input.DiagnosticNote, items); err != nil {
		return err
	}
	if err := s.repo.Save(ctx, order); err != nil {
		return fmt.Errorf("save order: %w", err)
	}
	return nil
}
