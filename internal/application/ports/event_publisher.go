package ports

import "context"

// EventPublisher publishes domain events to a messaging topic (SNS).
type EventPublisher interface {
	Publish(ctx context.Context, topicARN string, event any) error
}
