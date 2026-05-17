package order

import (
	"context"
	"sync"

	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/application/ports"
)

func (s *Service) GetOrderByID(ctx context.Context, id string) (*ports.OrderDetail, error) {
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	detail := &ports.OrderDetail{Order: *order}

	var mu sync.Mutex
	var wg sync.WaitGroup
	errs := make([]error, 0, 3)

	fetch := func(fn func() error) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := fn(); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		}()
	}

	fetch(func() error {
		c, err := s.users.GetCustomer(ctx, order.CustomerID)
		if err != nil {
			return err
		}
		mu.Lock()
		detail.Customer = c
		mu.Unlock()
		return nil
	})

	fetch(func() error {
		e, err := s.users.GetEmployee(ctx, order.EmployeeID)
		if err != nil {
			return err
		}
		mu.Lock()
		detail.Employee = e
		mu.Unlock()
		return nil
	})

	fetch(func() error {
		co, err := s.users.GetCompany(ctx, order.CompanyID)
		if err != nil {
			return err
		}
		mu.Lock()
		detail.Company = co
		mu.Unlock()
		return nil
	})

	wg.Wait()

	if len(errs) > 0 {
		return nil, errs[0]
	}

	return detail, nil
}
