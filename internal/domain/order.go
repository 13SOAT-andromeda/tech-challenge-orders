package domain

import (
	"errors"
	"time"
)

var ErrOrderNotFound = errors.New("order not found")

type Order struct {
	ID      string
	Status  Status
	Version int

	DateIn       time.Time
	DateOut      *time.Time
	DateApproved *time.Time
	DateRejected *time.Time
	PaymentDate  *time.Time
	LastStatusAt *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time

	VehicleKilometers int
	Note              *string
	DiagnosticNote    *string
	PriceCents        *int64

	CustomerVehicleID string
	CustomerID        string
	EmployeeID        string
	CompanyID         string

	Vehicle *VehicleSnapshot
	Items   []ItemSnapshot
	History []HistoryEntry
}

func (o *Order) transition(to Status, actor, reason string) error {
	if !CanTransition(o.Status, to) {
		return InvalidTransitionError(o.Status, to)
	}
	now := time.Now()
	o.History = append(o.History, HistoryEntry{
		From:   o.Status,
		To:     to,
		At:     now,
		Actor:  actor,
		Reason: reason,
	})
	o.Status = to
	o.LastStatusAt = &now
	o.UpdatedAt = now
	o.Version++
	return nil
}

func (o *Order) Assign(actorID string) error {
	return o.transition(IN_ANALYSIS, actorID, "")
}

func (o *Order) CompleteAnalysis(actorID string, note *string, items []ItemSnapshot) error {
	if err := o.transition(ANALYSIS_FINISHED, actorID, ""); err != nil {
		return err
	}
	if note != nil {
		o.DiagnosticNote = note
	}
	o.Items = items
	var total int64
	for _, item := range items {
		total += int64(item.Quantity) * item.UnitPriceCents
	}
	o.PriceCents = &total
	return nil
}

func (o *Order) RequestApproval(actorID string) error {
	return o.transition(AWAITING_APPROVAL, actorID, "")
}

func (o *Order) Approve() error {
	if err := o.transition(AWAITING_STOCK_CONSULT, "system", ""); err != nil {
		return err
	}
	o.DateApproved = o.LastStatusAt
	return nil
}

func (o *Order) Reject() error {
	if err := o.transition(REJECTED, "system", ""); err != nil {
		return err
	}
	o.DateRejected = o.LastStatusAt
	return nil
}

func (o *Order) MarkStockAvailable() error {
	return o.transition(IN_PROGRESS, "system", "stock available")
}

func (o *Order) MarkStockUnavailable() error {
	return o.transition(AWAITING_STOCK_ORDER, "system", "stock unavailable")
}

func (o *Order) MarkStockReady() error {
	return o.transition(IN_PROGRESS, "system", "stock replenished")
}

func (o *Order) CompleteWork(actorID string) error {
	return o.transition(FINISHED, actorID, "")
}

func (o *Order) MarkPaymentGenerated() error {
	return o.transition(AWAITING_PAYMENT, "system", "payment generated")
}

func (o *Order) MarkPaymentApproved() error {
	if err := o.transition(PAYMENT_APPROVED, "system", "payment approved"); err != nil {
		return err
	}
	o.PaymentDate = o.LastStatusAt
	return nil
}

func (o *Order) MarkPaymentFailed(reason string) error {
	return o.transition(PAYMENT_FAILED, "system", reason)
}

func (o *Order) Archive(actorID string) error {
	if err := o.transition(DELIVERED, actorID, ""); err != nil {
		return err
	}
	o.DateOut = o.LastStatusAt
	return nil
}
