package sns

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/domain"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	snstypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
)

type NotificationPublisher struct {
	client   *sns.Client
	topicARN string
}

func NewNotificationPublisher(client *sns.Client, topicARN string) *NotificationPublisher {
	return &NotificationPublisher{client: client, topicARN: topicARN}
}

type notifRecipient struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type notifMessage struct {
	Recipient notifRecipient    `json:"recipient"`
	Data      map[string]string `json:"data"`
}

func (p *NotificationPublisher) publish(ctx context.Context, templateType string, recipient notifRecipient, data map[string]string) error {
	body, err := json.Marshal(notifMessage{Recipient: recipient, Data: data})
	if err != nil {
		return fmt.Errorf("marshal notification: %w", err)
	}

	_, err = p.client.Publish(ctx, &sns.PublishInput{
		Message:  aws.String(string(body)),
		TopicArn: aws.String(p.topicARN),
		MessageAttributes: map[string]snstypes.MessageAttributeValue{
			"templateType": {
				DataType:    aws.String("String"),
				StringValue: aws.String(templateType),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("sns publish notification [%s]: %w", templateType, err)
	}
	return nil
}

func formatCents(cents int64) string {
	reais := cents / 100
	centavos := cents % 100
	if centavos < 0 {
		centavos = -centavos
	}
	return fmt.Sprintf("R$ %d,%02d", reais, centavos)
}

func itemRow(name string, qty uint, unitPriceCents int64) string {
	return fmt.Sprintf(
		`<tr><td style="padding:8px;border-bottom:1px solid #eee">%s</td><td style="padding:8px;border-bottom:1px solid #eee;text-align:center">%dx</td><td style="padding:8px;border-bottom:1px solid #eee;text-align:right">%s</td></tr>`,
		name, qty, formatCents(unitPriceCents),
	)
}

func renderItemRows(items []domain.ItemSnapshot) string {
	var sb strings.Builder
	for _, it := range items {
		sb.WriteString(itemRow(it.Name, it.Quantity, it.UnitPriceCents))
	}
	return sb.String()
}

func (p *NotificationPublisher) PublishApprovalRequest(ctx context.Context, evt domain.OrderApprovalRequested) error {
	return p.publish(ctx, "ORDER_APPROVAL_REQUEST",
		notifRecipient{Email: evt.CustomerEmail, Name: evt.CustomerName},
		map[string]string{
			"order_id":          evt.OrderID,
			"items":             renderItemRows(evt.Items),
			"total_amount":      formatCents(evt.TotalCents),
			"link_confirmation": evt.ApprovalURL,
			"link_cancellation": evt.RejectURL,
		},
	)
}

func (p *NotificationPublisher) PublishAwaitingPayment(ctx context.Context, order *domain.Order, customerEmail, customerName, paymentURL string) error {
	var products, services []domain.ItemSnapshot
	for _, it := range order.Items {
		if it.Kind == domain.ItemKindProduct {
			products = append(products, it)
		} else {
			services = append(services, it)
		}
	}

	totalCents := int64(0)
	if order.PriceCents != nil {
		totalCents = *order.PriceCents
	}

	return p.publish(ctx, "ORDER_AWAITING_PAYMENT",
		notifRecipient{Email: customerEmail, Name: customerName},
		map[string]string{
			"order_id":     order.ID,
			"products":     renderItemRows(products),
			"services":     renderItemRows(services),
			"total_amount": formatCents(totalCents),
			"link_payment": paymentURL,
		},
	)
}
