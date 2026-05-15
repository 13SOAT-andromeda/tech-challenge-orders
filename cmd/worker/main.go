package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	awssns "github.com/aws/aws-sdk-go-v2/service/sns"

	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/adapter/config"
	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/adapter/dynamo"
	dynamoidem "github.com/13SOAT-andromeda/tech-challenge-orders/internal/adapter/dynamo"
	dynamorepo "github.com/13SOAT-andromeda/tech-challenge-orders/internal/adapter/dynamo/repository"
	snspub "github.com/13SOAT-andromeda/tech-challenge-orders/internal/adapter/sns"
	sqsadapter "github.com/13SOAT-andromeda/tech-challenge-orders/internal/adapter/sqs"
	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/adapter/sqs/consumers"
	orderusecase "github.com/13SOAT-andromeda/tech-challenge-orders/internal/application/usecases/order"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar := logger.Sugar()

	cfg, err := config.Init()
	if err != nil {
		sugar.Fatalf("failed to load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// AWS config
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(cfg.DynamoDB.Region))
	if err != nil {
		sugar.Fatalf("failed to load aws config: %v", err)
	}

	// DynamoDB
	dynamoClient, err := dynamo.NewClient(ctx, dynamo.Config{
		Region:    cfg.DynamoDB.Region,
		Endpoint:  cfg.DynamoDB.Endpoint,
		TableName: cfg.DynamoDB.TableName,
	})
	if err != nil {
		sugar.Fatalf("failed to connect to DynamoDB: %v", err)
	}

	repo := dynamorepo.NewOrderRepository(dynamoClient, cfg.DynamoDB.TableName)
	idempotency := dynamoidem.NewIdempotencyStore(dynamoClient, cfg.DynamoDB.TableName)

	// SNS publisher and SQS client
	snsClient := awssns.NewFromConfig(awsCfg, func(o *awssns.Options) {
		o.BaseEndpoint = awsEndpoint()
	})
	sqsClient := awssqs.NewFromConfig(awsCfg, func(o *awssqs.Options) {
		o.BaseEndpoint = awsEndpoint()
	})

	publisher := snspub.NewPublisher(snsClient)

	// Service (users/stock clients not needed by Handle* use cases)
	svc := orderusecase.NewService(repo, publisher, nil, nil, idempotency, cfg.SNS.OrdersTopicARN, cfg.Http.PublicBaseURL)

	// Consumers
	type consumerDef struct {
		name     string
		queueURL string
		handler  sqsadapter.Handler
	}

	defs := []consumerDef{
		{"stock-available", cfg.SQS.StockAvailableURL, consumers.StockAvailable(svc)},
		{"stock-unavailable", cfg.SQS.StockUnavailableURL, consumers.StockUnavailable(svc)},
		{"stock-updated", cfg.SQS.StockUpdatedURL, consumers.StockUpdated(svc)},
		{"payment-events", cfg.SQS.PaymentEventsURL, consumers.PaymentEvents(svc)},
	}

	g, gctx := errgroup.WithContext(ctx)

	for _, d := range defs {
		consumer := sqsadapter.NewConsumer(sqsClient, d.queueURL, d.handler, logger.With(zap.String("consumer", d.name)))
		g.Go(func() error {
			sugar.Infow("consumer starting", "name", d.name, "queue", d.queueURL)
			return consumer.Run(gctx)
		})
	}

	sugar.Info("worker started")
	if err := g.Wait(); err != nil {
		sugar.Errorw("worker stopped with error", "error", err)
	} else {
		sugar.Info("worker stopped")
	}
}

// awsEndpoint returns the LocalStack endpoint URL if AWS_ENDPOINT_URL is set.
func awsEndpoint() *string {
	if v := os.Getenv("AWS_ENDPOINT_URL"); v != "" {
		return &v
	}
	return nil
}
