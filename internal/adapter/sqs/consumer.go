package sqs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"go.uber.org/zap"
)

// Handler is called for each message with the parsed event_type, event_id and raw data payload.
// Return nil to delete the message from the queue; return an error to leave it for retry/DLQ.
type Handler func(ctx context.Context, eventType, eventID string, data json.RawMessage) error

// Consumer polls a single SQS queue and dispatches each message to Handler.
type Consumer struct {
	client   *sqs.Client
	queueURL string
	handler  Handler
	logger   *zap.Logger
}

func NewConsumer(client *sqs.Client, queueURL string, handler Handler, logger *zap.Logger) *Consumer {
	return &Consumer{
		client:   client,
		queueURL: queueURL,
		handler:  handler,
		logger:   logger,
	}
}

// snsNotification is the outer envelope SNS places around our message when delivering to SQS.
type snsNotification struct {
	Message string `json:"Message"`
}

type eventEnvelope struct {
	EventID   string          `json:"event_id"`
	EventType string          `json:"event_type"`
	Data      json.RawMessage `json:"data"`
	Payload   json.RawMessage `json:"payload"`
}

// Run polls the queue until ctx is cancelled. It never returns a non-nil error for
// transient receive failures — those are logged and retried on the next iteration.
func (c *Consumer) Run(ctx context.Context) error {
	for {
		if ctx.Err() != nil {
			return nil
		}

		output, err := c.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(c.queueURL),
			MaxNumberOfMessages: 10,
			WaitTimeSeconds:     20, // long-polling
		})
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			c.logger.Warn("sqs receive error", zap.String("queue", c.queueURL), zap.Error(err))
			continue
		}

		for _, msg := range output.Messages {
			if err := c.process(ctx, *msg.Body); err != nil {
				c.logger.Error("message processing failed",
					zap.String("queue", c.queueURL),
					zap.Stringp("message_id", msg.MessageId),
					zap.Error(err),
				)
				continue // leave message for retry / DLQ
			}
			if _, err := c.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
				QueueUrl:      aws.String(c.queueURL),
				ReceiptHandle: msg.ReceiptHandle,
			}); err != nil {
				c.logger.Warn("failed to delete message",
					zap.String("queue", c.queueURL),
					zap.Stringp("message_id", msg.MessageId),
					zap.Error(err),
				)
			}
		}
	}
}

func (c *Consumer) process(ctx context.Context, body string) error {
	// SNS wraps our message in a notification envelope
	var notification snsNotification
	if err := json.Unmarshal([]byte(body), &notification); err != nil {
		return fmt.Errorf("unmarshal sns notification: %w", err)
	}

	var envelope eventEnvelope
	if err := json.Unmarshal([]byte(notification.Message), &envelope); err != nil {
		return fmt.Errorf("unmarshal event envelope: %w", err)
	}

	data := envelope.Data
	if len(data) == 0 {
		data = envelope.Payload
	}
	// If neither data nor payload fields exist, the inner message is itself the payload
	// (e.g. publishers that send raw JSON without an envelope wrapper).
	if len(data) == 0 {
		data = json.RawMessage(notification.Message)
	}

	return c.handler(ctx, envelope.EventType, envelope.EventID, data)
}
