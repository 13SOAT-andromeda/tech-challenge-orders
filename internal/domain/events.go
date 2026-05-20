package domain

import "time"

// OrderApprovalRequested is published after RequestApproval.
// Consumed by: Lambda Notification (sends approval email with approve/reject links).
type OrderApprovalRequested struct {
	OrderID        string
	CustomerID     int64
	CustomerName   string
	CustomerEmail  string
	Vehicle        VehicleSnapshot
	DiagnosticNote string
	Items          []ItemSnapshot
	TotalCents     int64
	ApprovalURL    string
	RejectURL      string
	ExpiresAt      time.Time
}

// OrderApproved is published after Approve.
// Consumed by: Stock service (reserves/decrements product items).
type OrderApproved struct {
	OrderID       string
	CustomerName  string
	CustomerEmail string
	Items         []ItemSnapshot // only Kind == ItemKindProduct entries
	ApprovedAt    time.Time
}

// OrderFinished is published after CompleteWork.
// Consumed by: Payments service (generates a payment request).
type OrderFinished struct {
	OrderID       string
	CustomerID    int64
	CustomerEmail string
	AmountCents   int64
	Items         []ItemSnapshot
	FinishedAt    time.Time
}
