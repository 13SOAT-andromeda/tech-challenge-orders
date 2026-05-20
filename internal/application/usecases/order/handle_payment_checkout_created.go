package order

import (
	"context"
	"fmt"
	"log"
)

func (s *Service) HandlePaymentCheckoutCreated(ctx context.Context, eventID, orderID, paymentURL string) error {
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
	if err := order.MarkPaymentCheckoutCreated(); err != nil {
		return err
	}
	if err := s.repo.Save(ctx, order); err != nil {
		return fmt.Errorf("save order: %w", err)
	}

	if s.notifPublisher != nil {
		customer, err := s.users.GetCustomer(ctx, order.CustomerID)
		if err != nil {
			log.Printf("warn: awaiting payment notification: get customer failed: %v", err)
		} else {
			if err := s.notifPublisher.PublishAwaitingPayment(ctx, order, customer.Email, customer.Name, paymentURL); err != nil {
				log.Printf("warn: awaiting payment notification failed: %v", err)
			}
		}
	}

	return nil
}
