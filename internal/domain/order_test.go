package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newOrder(status domain.Status) *domain.Order {
	now := time.Now()
	return &domain.Order{
		ID:        "order-1",
		Status:    status,
		Version:   0,
		DateIn:    now,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestOrder_ValidTransitions(t *testing.T) {
	tests := []struct {
		name    string
		initial domain.Status
		trigger func(*domain.Order) error
		want    domain.Status
	}{
		{"assign", domain.RECEIVED, func(o *domain.Order) error { return o.Assign("emp-1") }, domain.IN_ANALYSIS},
		{"complete analysis", domain.IN_ANALYSIS, func(o *domain.Order) error {
			return o.CompleteAnalysis("emp-1", nil, nil)
		}, domain.ANALYSIS_FINISHED},
		{"request approval", domain.ANALYSIS_FINISHED, func(o *domain.Order) error { return o.RequestApproval("emp-1") }, domain.AWAITING_APPROVAL},
		{"approve", domain.AWAITING_APPROVAL, func(o *domain.Order) error { return o.Approve() }, domain.AWAITING_STOCK_CONSULT},
		{"reject", domain.AWAITING_APPROVAL, func(o *domain.Order) error { return o.Reject() }, domain.REJECTED},
		{"stock available", domain.AWAITING_STOCK_CONSULT, func(o *domain.Order) error { return o.MarkStockAvailable() }, domain.IN_PROGRESS},
		{"stock unavailable", domain.AWAITING_STOCK_CONSULT, func(o *domain.Order) error { return o.MarkStockUnavailable() }, domain.AWAITING_STOCK_ORDER},
		{"stock ready", domain.AWAITING_STOCK_ORDER, func(o *domain.Order) error { return o.MarkStockReady() }, domain.IN_PROGRESS},
		{"complete work", domain.IN_PROGRESS, func(o *domain.Order) error { return o.CompleteWork("emp-1") }, domain.FINISHED},
		{"payment checkout created", domain.FINISHED, func(o *domain.Order) error { return o.MarkPaymentCheckoutCreated() }, domain.AWAITING_PAYMENT},
		{"payment approved", domain.AWAITING_PAYMENT, func(o *domain.Order) error { return o.MarkPaymentApproved() }, domain.PAYMENT_APPROVED},
		{"payment failed", domain.AWAITING_PAYMENT, func(o *domain.Order) error { return o.MarkPaymentFailed("cancelled") }, domain.PAYMENT_FAILED},
		{"archive", domain.PAYMENT_APPROVED, func(o *domain.Order) error { return o.Archive("adm-1") }, domain.DELIVERED},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := newOrder(tt.initial)
			prevVersion := o.Version

			err := tt.trigger(o)

			require.NoError(t, err)
			assert.Equal(t, tt.want, o.Status)
			assert.Equal(t, prevVersion+1, o.Version)
			assert.NotNil(t, o.LastStatusAt)
			assert.Len(t, o.History, 1)
			assert.Equal(t, tt.initial, o.History[0].From)
			assert.Equal(t, tt.want, o.History[0].To)
		})
	}
}

func TestOrder_InvalidTransitions(t *testing.T) {
	invalid := []struct {
		name    string
		initial domain.Status
		trigger func(*domain.Order) error
	}{
		{"assign from analysis", domain.IN_ANALYSIS, func(o *domain.Order) error { return o.Assign("u") }},
		{"approve from received", domain.RECEIVED, func(o *domain.Order) error { return o.Approve() }},
		{"approve from in_progress", domain.IN_PROGRESS, func(o *domain.Order) error { return o.Approve() }},
		{"complete work from received", domain.RECEIVED, func(o *domain.Order) error { return o.CompleteWork("u") }},
		{"archive from finished", domain.FINISHED, func(o *domain.Order) error { return o.Archive("u") }},
		{"archive from delivered", domain.DELIVERED, func(o *domain.Order) error { return o.Archive("u") }},
		{"stock available from received", domain.RECEIVED, func(o *domain.Order) error { return o.MarkStockAvailable() }},
		{"payment checkout created from in_progress", domain.IN_PROGRESS, func(o *domain.Order) error { return o.MarkPaymentCheckoutCreated() }},
	}

	for _, tt := range invalid {
		t.Run(tt.name, func(t *testing.T) {
			o := newOrder(tt.initial)
			prevVersion := o.Version

			err := tt.trigger(o)

			require.Error(t, err)
			assert.True(t, errors.Is(err, domain.ErrInvalidTransition), "expected ErrInvalidTransition, got: %v", err)
			assert.Equal(t, tt.initial, o.Status, "status must not change on failure")
			assert.Equal(t, prevVersion, o.Version, "version must not change on failure")
			assert.Empty(t, o.History, "history must not grow on failure")
		})
	}
}

func TestOrder_Approve_SetsDateApproved(t *testing.T) {
	o := newOrder(domain.AWAITING_APPROVAL)
	err := o.Approve()
	require.NoError(t, err)
	require.NotNil(t, o.DateApproved)
	assert.Equal(t, o.LastStatusAt, o.DateApproved)
}

func TestOrder_Reject_SetsDateRejected(t *testing.T) {
	o := newOrder(domain.AWAITING_APPROVAL)
	err := o.Reject()
	require.NoError(t, err)
	require.NotNil(t, o.DateRejected)
	assert.Equal(t, o.LastStatusAt, o.DateRejected)
}

func TestOrder_Archive_SetsDateOut(t *testing.T) {
	o := newOrder(domain.PAYMENT_APPROVED)
	err := o.Archive("adm-1")
	require.NoError(t, err)
	require.NotNil(t, o.DateOut)
	assert.Equal(t, o.LastStatusAt, o.DateOut)
}

func TestOrder_PaymentApproved_SetsPaymentDate(t *testing.T) {
	o := newOrder(domain.AWAITING_PAYMENT)
	err := o.MarkPaymentApproved()
	require.NoError(t, err)
	require.NotNil(t, o.PaymentDate)
	assert.Equal(t, o.LastStatusAt, o.PaymentDate)
}

func TestOrder_CompleteAnalysis_CalculatesTotalAndFreezesItems(t *testing.T) {
	o := newOrder(domain.IN_ANALYSIS)
	note := "trocar correia"
	items := []domain.ItemSnapshot{
		{ID: "p1", Kind: domain.ItemKindProduct, Name: "Correia", Quantity: 2, UnitPriceCents: 5000},
		{ID: "m1", Kind: domain.ItemKindMaintenance, Name: "Troca", Quantity: 1, UnitPriceCents: 8000},
	}

	err := o.CompleteAnalysis("emp-1", &note, items)

	require.NoError(t, err)
	assert.Equal(t, domain.ANALYSIS_FINISHED, o.Status)
	assert.Equal(t, &note, o.DiagnosticNote)
	assert.Equal(t, items, o.Items)
	require.NotNil(t, o.PriceCents)
	assert.Equal(t, int64(18000), *o.PriceCents) // 2*5000 + 1*8000
}

func TestOrder_HistoryAccumulates(t *testing.T) {
	o := newOrder(domain.RECEIVED)

	_ = o.Assign("emp-1")
	_ = o.CompleteAnalysis("emp-1", nil, nil)
	_ = o.RequestApproval("emp-1")

	assert.Len(t, o.History, 3)
	assert.Equal(t, domain.RECEIVED, o.History[0].From)
	assert.Equal(t, domain.IN_ANALYSIS, o.History[0].To)
	assert.Equal(t, domain.AWAITING_APPROVAL, o.History[2].To)
}

func TestCanTransition(t *testing.T) {
	assert.True(t, domain.CanTransition(domain.RECEIVED, domain.IN_ANALYSIS))
	assert.False(t, domain.CanTransition(domain.RECEIVED, domain.FINISHED))
	assert.False(t, domain.CanTransition(domain.DELIVERED, domain.RECEIVED))
}
