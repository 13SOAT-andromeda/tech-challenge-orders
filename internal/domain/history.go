package domain

import "time"

// HistoryEntry records a single status transition on an Order.
type HistoryEntry struct {
	From   Status    `json:"from"`
	To     Status    `json:"to"`
	At     time.Time `json:"at"`
	Actor  string    `json:"actor"`  // user ID or "system"
	Reason string    `json:"reason,omitempty"`
}
