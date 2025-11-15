package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"

	appcache "github.com/example/interview-stack/go-service/internal/application/cache"
	"github.com/example/interview-stack/go-service/internal/application/usecase"
	"github.com/example/interview-stack/go-service/internal/config"
	"github.com/example/interview-stack/go-service/internal/db"
	"github.com/example/interview-stack/go-service/internal/docs"
	"github.com/example/interview-stack/go-service/internal/domain"
	cacheinfra "github.com/example/interview-stack/go-service/internal/infrastructure/cache"
	clockinfra "github.com/example/interview-stack/go-service/internal/infrastructure/clock"
	"github.com/example/interview-stack/go-service/internal/infrastructure/logging"
	"github.com/example/interview-stack/go-service/internal/infrastructure/nosql"
	queueinfra "github.com/example/interview-stack/go-service/internal/infrastructure/queue"
	"github.com/example/interview-stack/go-service/internal/infrastructure/repository"
	"github.com/example/interview-stack/go-service/internal/infrastructure/sql"
	httpInterface "github.com/example/interview-stack/go-service/internal/interface/http"
	obsmetrics "github.com/example/interview-stack/go-service/internal/observability/metrics"
)

const serviceName = "go-service"

func main() {
	baseLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(baseLogger)
	logger := logging.NewStructuredLogger(baseLogger)

	docs.SwaggerInfo.Title = "Products API (Go)"
	docs.SwaggerInfo.Description = "Auto-generated documentation for the Fiber service."
	docs.SwaggerInfo.Version = "1.0.0"
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Schemes = []string{"http"}

	cfg, err := config.Load("")
	if err != nil {
		baseLogger.Error("failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	obsmetrics.Init(serviceName, cfg.Metrics.Enabled)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	clk := clockinfra.NewSystemClock()

	cacheCfg := appcache.Config{
		Enabled:    cfg.Cache.Enabled,
		DefaultTTL: cfg.Cache.DefaultTTL.Duration(),
		StaleTTL:   cfg.Cache.StaleTTL.Duration(),
	}
	var redisClient *cacheinfra.RedisClient
	if cfg.Cache.Enabled {
		redisClient = cacheinfra.NewRedisClient(cfg.Cache.Redis)
		if redisClient != nil {
			pingCtx, cancel := context.WithTimeout(ctx, cfg.Server.RequestTimeouts.Read.Duration())
			if err := redisClient.Ping(pingCtx); err != nil {
				baseLogger.Warn("redis unavailable, disabling cache", slog.String("error", err.Error()))
				redisClient.Close()
				redisClient = nil
				cacheCfg.Enabled = false
			}
			cancel()
		} else {
			baseLogger.Warn("redis not configured, disabling cache")
			cacheCfg.Enabled = false
		}
	}
	cacheSvc := appcache.NewService(redisClient, cacheCfg, logger, clk)
	if redisClient != nil {
		defer redisClient.Close()
		baseLogger.Info("redis cache connected",
			slog.String("host", cfg.Cache.Redis.Host),
			slog.Int("port", cfg.Cache.Redis.Port),
			slog.Duration("default_ttl", cacheCfg.DefaultTTL),
			slog.Duration("stale_ttl", cacheCfg.StaleTTL),
		)
	}

	var store repository.ProductStore
	switch cfg.Database.Type {
	case "sql":
		pgPool, err := db.NewPostgresPool(ctx, cfg.Database.Postgres)
		if err != nil {
			baseLogger.Error("failed to connect to postgres", slog.String("error", err.Error()))
			os.Exit(1)
		}
		defer pgPool.Close()
		store = sql.NewProductStore(pgPool)
	case "mongo":
		mongoCollection, err := db.NewMongoCollection(ctx, cfg.Database.Mongo)
		if err != nil {
			baseLogger.Error("failed to connect to mongo", slog.String("error", err.Error()))
			os.Exit(1)
		}
		mongoStore := nosql.NewProductStore(mongoCollection, cfg.Database.Mongo.OperationTimeout.Duration())
		if err := mongoStore.EnsureIndexes(ctx); err != nil {
			baseLogger.Warn("failed to ensure mongo indexes", slog.String("error", err.Error()))
		}
		store = mongoStore
	default:
		baseLogger.Error("unsupported database.type", slog.String("value", cfg.Database.Type))
		os.Exit(1)
	}

	var streamClient *redis.Client
	var writeQueue domain.WriteQueue
	if cfg.WriteBehind.Enabled {
		streamClient = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", cfg.Cache.Redis.Host, cfg.Cache.Redis.Port),
			Password: cfg.Cache.Redis.Password,
			DB:       cfg.Cache.Redis.DB,
		})
		if err := streamClient.Ping(ctx).Err(); err != nil {
			baseLogger.Error("failed to connect to redis stream", slog.String("error", err.Error()))
			os.Exit(1)
		}
		writeQueue = queueinfra.NewRedisStreamQueue(streamClient, cfg.WriteBehind.StreamName)
		defer streamClient.Close()
	}

	productRepo := repository.NewProductRepository(store, clk)
	baseLogger.Info("product store configured", slog.String("db_type", cfg.Database.Type))

	listUC := usecase.NewListProductsUseCase(productRepo, logger, cacheSvc)
	getUC := usecase.NewGetProductUseCase(productRepo, logger, cacheSvc)
	createUC := usecase.NewCreateProductUseCase(productRepo, logger, clk, cacheSvc, writeQueue, cfg.WriteBehind.Enabled, serviceName)
	updateUC := usecase.NewUpdateProductUseCase(productRepo, logger, clk, cacheSvc, writeQueue, cfg.WriteBehind.Enabled, serviceName)
	deleteUC := usecase.NewDeleteProductUseCase(productRepo, logger, clk, cacheSvc, writeQueue, cfg.WriteBehind.Enabled, serviceName)
	healthUC := usecase.NewHealthCheckUseCase(productRepo)

	controller := httpInterface.NewProductController(
		listUC,
		getUC,
		createUC,
		updateUC,
		deleteUC,
		healthUC,
		httpInterface.RequestTimeouts{
			Read:   cfg.Server.RequestTimeouts.Read.Duration(),
			Write:  cfg.Server.RequestTimeouts.Write.Duration(),
			Health: cfg.Server.RequestTimeouts.Health.Duration(),
		},
	)

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ReadTimeout:           cfg.Server.ReadTimeout.Duration(),
		WriteTimeout:          cfg.Server.WriteTimeout.Duration(),
		IdleTimeout:           cfg.Server.IdleTimeout.Duration(),
	})
	app.Use(recover.New())
	app.Use(loggerMiddleware())
	app.Use(httpMetricsMiddleware(cfg.Metrics.Enabled))

	httpInterface.RegisterRoutes(app, controller)

	if cfg.WriteBehind.Enabled && streamClient != nil {
		consumerName := fmt.Sprintf("%s-worker-%d", serviceName, time.Now().UnixNano())
		worker := queueinfra.NewWorker(
			streamClient,
			cfg.WriteBehind.StreamName,
			serviceName,
			consumerName,
			serviceName,
			cfg.WriteBehind.BatchSize,
			cfg.WriteBehind.FlushInterval.Duration(),
			store,
			logger,
		)
		go worker.Run(ctx)
	}

	metricsServer := registerMetricsEndpoint(app, cfg, baseLogger)

	go gracefulShutdown(app, metricsServer, baseLogger, cfg.Server.ShutdownTimeout.Duration(), cancel)

	baseLogger.Info("starting go-service", slog.Int("port", cfg.Server.Port))
	if err := app.Listen(fmt.Sprintf(":%d", cfg.Server.Port)); err != nil {
		baseLogger.Error("fiber stopped", slog.String("error", err.Error()))
	}
}

func loggerMiddleware() fiber.Handler {
	return logger.New(logger.Config{
		Format:     "${time} ${status} - ${latency} ${method} ${path}\n",
		TimeFormat: time.RFC3339,
	})
}

func httpMetricsMiddleware(enabled bool) fiber.Handler {
	if !enabled {
		return func(c *fiber.Ctx) error {
			return c.Next()
		}
	}
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		path := c.Path()
		if route := c.Route(); route != nil && route.Path != "" {
			path = route.Path
		}
		status := c.Response().StatusCode()
		obsmetrics.ObserveHTTPRequest(c.Method(), path, status, time.Since(start))
		return err
	}
}

func registerMetricsEndpoint(app *fiber.App, cfg *config.Config, log *slog.Logger) *http.Server {
	if cfg == nil || !cfg.Metrics.Enabled {
		return nil
	}
	handler := promhttp.Handler()
	if cfg.Metrics.Port == cfg.Server.Port || cfg.Metrics.Port == 0 {
		app.Get(cfg.Metrics.Path, adaptor.HTTPHandler(handler))
		log.Info("metrics endpoint registered", slog.String("path", cfg.Metrics.Path), slog.Int("port", cfg.Server.Port))
		return nil
	}
	mux := http.NewServeMux()
	mux.Handle(cfg.Metrics.Path, handler)
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Metrics.Port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("metrics server stopped unexpectedly", slog.String("error", err.Error()))
		}
	}()
	log.Info("metrics server listening", slog.String("path", cfg.Metrics.Path), slog.Int("port", cfg.Metrics.Port))
	return server
}

func gracefulShutdown(app *fiber.App, metricsSrv *http.Server, log *slog.Logger, timeout time.Duration, cancel context.CancelFunc) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	cancel()
	ctx, shutdownCancel := context.WithTimeout(context.Background(), timeout)
	defer shutdownCancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Error("failed to shutdown server", slog.String("error", err.Error()))
	}
	if metricsSrv != nil {
		if err := metricsSrv.Shutdown(ctx); err != nil {
			log.Warn("failed to shutdown metrics server", slog.String("error", err.Error()))
		}
	}
}
