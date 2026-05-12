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

	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/adapter/config"
	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/adapter/database"
	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/adapter/database/model/order"
	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/adapter/database/repository"
	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/adapter/email"
	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/adapter/http"
	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/adapter/http/handlers"
	appmetrics "github.com/13SOAT-andromeda/tech-challenge-orders/internal/adapter/metrics"
	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/application/ports"
	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/application/services"
	orderUsecase "github.com/13SOAT-andromeda/tech-challenge-orders/internal/application/usecases/order"
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
	db, err := database.Init(ctx, *cfg.Database)
	if err != nil {
		sugar.Fatalf("failed to connect database: %v", err)
	}

	sugar.Infof("Connecting to database")

	err = db.AutoMigrate(
		&order.Model{},
	)

	if err != nil {
		sugar.Fatalf("Error to executing migration: %s", err)
	}

	dbase := db.GetDB()

	if err = database.Seed(dbase); err != nil {
		sugar.Fatalf("failed to seed database: %v", err)
	}
	apiUrl := cfg.Http.ApiUrl

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
			sugar.Warnw("dogstatsd indisponível, métricas de ordem desativadas", "error", errStatsd)
		} else {
			defer statsdClient.Close()
			orderMetrics = appmetrics.NewOrderStatsd(statsdClient)
		}
	}

	// Repositories
	orderRepository := repository.NewOrderRepository(dbase)

	// Services
	orderService := services.NewOrderService(orderRepository)
	emailService := email.NewSendtrap(cfg.MailTrap.ApiKey, cfg.MailTrap.ApiUrl)

	// UseCases
	createOrderUseCase := orderUsecase.NewOrderUseCase(orderService, productService, maintenanceService, customerService, userService, employeeService, emailService, orderRepository, orderProductRepository, orderMaintenanceRepository, apiUrl, orderMetrics)

	// Handlers
	orderHandler := handlers.NewOrderHandler(orderService, createOrderUseCase)

	router := http.NewRouter(*cfg, logger, *orderHandler)
	sugar.Info("Starting HTTP server on port %s", cfg.Http.Port)

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
