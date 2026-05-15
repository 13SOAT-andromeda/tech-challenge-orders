package http

import (
	"net/http"
	"time"

	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/adapter/config"
	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/adapter/http/handlers"
	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/adapter/http/middlewares"
	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/adapter/http/response"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	gintrace "github.com/DataDog/dd-trace-go/contrib/gin-gonic/gin/v2"
	"github.com/DataDog/dd-trace-go/v2/ddtrace/tracer"
	ginzap "github.com/gin-contrib/zap"
	swaggerFiles "github.com/swaggo/files"
	swagger "github.com/swaggo/gin-swagger"
)

type Router struct {
	*gin.Engine
}

func NewRouter(
	config config.Config,
	logger *zap.Logger,
	orderHandler handlers.OrderHandler,
) *Router {
	if config.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = config.Http.AllowedOrigins
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}

	router := gin.New()
	router.Use(gintrace.Middleware("tech-challenge-api",
		gintrace.WithUseGinErrors(),
		gintrace.WithAnalytics(true),
		gintrace.WithIgnoreRequest(func(c *gin.Context) bool {
			return c.Request.URL.Path == "/api/health"
		}),
		gintrace.WithStatusCheck(func(statusCode int) bool {
			return statusCode >= 400
		}),
	))
	router.Use(ginzap.GinzapWithConfig(logger, &ginzap.Config{
		UTC:        true,
		TimeFormat: time.RFC3339,
		Context: ginzap.Fn(func(c *gin.Context) []zapcore.Field {
			fields := []zapcore.Field{}
			span, ok := tracer.SpanFromContext(c.Request.Context())

			if ok {
				fields = append(fields, zap.String("trace_id", span.Context().TraceID()))
				fields = append(fields, zap.String("span_id", span.String()))
			}
			return fields
		}),
	}))
	router.Use(ginzap.RecoveryWithZap(logger, true))
	router.Use(
		cors.New(corsConfig),
	)
	api := router.Group("/api")
	protected := api.Group("/")
	protected.Use(middlewares.AuthRequired())
	{
		orderGroup := protected.Group("/orders")
		orderGroup.Use(middlewares.RoleRequired("administrator", "attendant", "mechanic"))
		{
			orderGroup.GET("", orderHandler.GetAll)
			orderGroup.GET("/:id", orderHandler.GetByID)
			orderGroup.GET("/in-progress", orderHandler.GetInProgress)
			orderGroup.POST("", orderHandler.Create)
			orderGroup.POST("/:id/assign", orderHandler.Assign)
			orderGroup.POST("/:id/complete-analysis", orderHandler.CompleteAnalysis)
			orderGroup.POST("/:id/approve", orderHandler.ApproveOrder)
			orderGroup.POST("/:id/reject", orderHandler.RejectOrder)
			orderGroup.POST("/:id/request-approval", orderHandler.RequestApproval)
			orderGroup.POST("/:id/complete-work", orderHandler.CompleteWork)
			orderGroup.POST("/:id/archive", orderHandler.ArchiveOrder)
		}
	}

	// Public routes — no authorization middleware applied.
	publicOrders := api.Group("/orders")
	{
		publicOrders.GET("/:id/approve", orderHandler.ApproveOrder)
		publicOrders.GET("/:id/reject", orderHandler.RejectOrder)
	}

	api.GET("/health", func(c *gin.Context) {
		response.RespondSuccess(c, http.StatusOK, "ok")
	})

	// serve static swagger files from ./swagger so /api/swagger/swagger.yaml is available
	api.Static("/swagger", "./swagger")

	// Serve the swagger UI under /api/docs and point it to the static spec at /api/swagger/swagger.yaml
	api.GET("/docs/*any", swagger.WrapHandler(swaggerFiles.Handler, swagger.URL("/api/swagger/swagger.yaml")))

	api.StaticFile("/redoc", "./swagger/redoc.html")

	return &Router{router}
}

func (r *Router) Server(listenAddr string) error {
	return r.Run(listenAddr)
}
