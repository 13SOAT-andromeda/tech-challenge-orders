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
	OrderID       string         `json:"order_id"`
	CustomerName  string         `json:"customer_name"`
	CustomerEmail string         `json:"customer_email"`
	Items         []ItemSnapshot `json:"items"`
	ApprovedAt    time.Time      `json:"approved_at"`
}

// OrderFinished is published after CompleteWork.
// Consumed by: Payments service (generates a payment request).
type OrderFinished struct {
	OrderID       string         `json:"order_id"`
	CustomerID    int64          `json:"customer_id"`
	CustomerEmail string         `json:"customer_email"`
	AmountCents   int64          `json:"amount_cents"`
	Items         []ItemSnapshot `json:"items"`
	FinishedAt    time.Time      `json:"finished_at"`
}
