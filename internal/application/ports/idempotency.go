package ports

import "context"

// IdempotencyStore prevents duplicate processing of the same event.
type IdempotencyStore interface {
	// MarkProcessed atomically records eventID as processed.
	// Returns (true, nil) on first call for this eventID.
	// Returns (false, nil) if already processed — caller should skip the event.
	MarkProcessed(ctx context.Context, eventID string) (bool, error)
}
