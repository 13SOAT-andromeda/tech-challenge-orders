package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/DataDog/dd-trace-go/v2/ddtrace/tracer"
	"github.com/DataDog/dd-trace-go/v2/profiler"
	"go.uber.org/zap"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awssns "github.com/aws/aws-sdk-go-v2/service/sns"

	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/adapter/clients"
	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/adapter/config"
	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/adapter/dynamo"
	dynamoidem "github.com/13SOAT-andromeda/tech-challenge-orders/internal/adapter/dynamo"
	dynamorepo "github.com/13SOAT-andromeda/tech-challenge-orders/internal/adapter/dynamo/repository"
	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/adapter/http"
	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/adapter/http/handlers"
	appmetrics "github.com/13SOAT-andromeda/tech-challenge-orders/internal/adapter/metrics"
	snspub "github.com/13SOAT-andromeda/tech-challenge-orders/internal/adapter/sns"
	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/application/ports"
	orderusecase "github.com/13SOAT-andromeda/tech-challenge-orders/internal/application/usecases/order"
)

func main() {
	logger, err := zap.NewProduction()

	defer func() {
		tracer.Stop()
		profiler.Stop()
		logger.Sync()
	}()

	sugar := logger.Sugar()

	if err != nil {
		sugar.Fatalf("Error on start logger zap: %s", err)
	}

	cfg, err := config.Init()
	if err != nil {
		sugar.Fatalf("failed to load config: %v", err)
	}

	err = profiler.Start(
		profiler.WithEnv(cfg.Env),
		profiler.WithService(cfg.Service),
		profiler.WithVersion(cfg.Version),
		profiler.WithTags("layer:api"),
		profiler.WithProfileTypes(
			profiler.CPUProfile,
			profiler.HeapProfile,
		),
	)
	if err != nil {
		sugar.Fatalf("Error on start datadog profiler: %s", err)
	}

	err = tracer.Start(
		tracer.WithEnv(cfg.Env),
		tracer.WithService(cfg.Service),
		tracer.WithServiceVersion(cfg.Version),
	)
	if err != nil {
		sugar.Fatalf("Error on start datadog tracer: %s", err)
	}

	ctx := context.Background()

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

	// AWS config for SNS
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(cfg.DynamoDB.Region))
	if err != nil {
		sugar.Fatalf("failed to load aws config: %v", err)
	}

	snsClient := awssns.NewFromConfig(awsCfg, func(o *awssns.Options) {
		if endpoint := os.Getenv("AWS_ENDPOINT_URL"); endpoint != "" {
			o.BaseEndpoint = &endpoint
		}
	})
	publisher := snspub.NewPublisher(snsClient)

	var notifPublisher ports.NotificationPublisher
	if cfg.SNS.NotificationTopicARN != "" {
		notifPublisher = snspub.NewNotificationPublisher(snsClient, cfg.SNS.NotificationTopicARN)
	}

	// HTTP clients
	usersClient := clients.NewUsersHTTPClient(cfg.Clients.UsersBaseURL, cfg.Clients.TimeoutMs)
	stockClient := clients.NewStockHTTPClient(cfg.Clients.StockBaseURL, cfg.Clients.TimeoutMs)


	// Metrics
	var orderMetrics ports.OrderMetrics = appmetrics.NoopOrderMetrics{}
	if !cfg.DogStatsD.Disabled && cfg.DogStatsD.Addr != "" {
		statsdClient, errStatsd := statsd.New(cfg.DogStatsD.Addr,
			statsd.WithNamespace("tech_challenge."),
			statsd.WithTags([]string{
				"env:" + cfg.Env,
				"service:" + cfg.Service,
				"version:" + cfg.Version,
			}),
		)
		if errStatsd != nil {
			sugar.Warnw("dogstatsd unavailable, order metrics disabled", "error", errStatsd)
		} else {
			defer statsdClient.Close()
			orderMetrics = appmetrics.NewOrderStatsd(statsdClient)
		}
	}

	// Use case service
	svc := orderusecase.NewService(
		repo, publisher, notifPublisher, usersClient, stockClient, idempotency, orderMetrics,
		cfg.SNS.OrdersTopicARN, cfg.Http.PublicBaseURL,
	)

	orderHandler := handlers.NewOrderHandler(svc)

	router := http.NewRouter(*cfg, logger, *orderHandler)
	sugar.Infof("Starting HTTP server on port %s", cfg.Http.Port)

	if err = router.Server(":" + cfg.Http.Port); err != nil {
		sugar.Fatalf("failed to start server: %v", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM)
	go func() {
		<-sigChan
		tracer.Stop()
	}()
}
