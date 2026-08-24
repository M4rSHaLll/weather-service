package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/M4rSHaLll/weather-service/internal/client/http/geocoding"
	"github.com/M4rSHaLll/weather-service/internal/client/http/openmeteo"
	"github.com/M4rSHaLll/weather-service/internal/collector"
	"github.com/M4rSHaLll/weather-service/internal/config"
	"github.com/M4rSHaLll/weather-service/internal/service"
	"github.com/M4rSHaLll/weather-service/internal/storage/cached"
	"github.com/M4rSHaLll/weather-service/internal/storage/postgres"
	redisstorage "github.com/M4rSHaLll/weather-service/internal/storage/redis"
	httptransport "github.com/M4rSHaLll/weather-service/internal/transport/http"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func Run(ctx context.Context, cfg config.Config, logger *slog.Logger) error {
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("create postgres pool: %w", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}

	redisOptions, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return fmt.Errorf("parse redis URL: %w", err)
	}
	redisOptions.DialTimeout = cfg.RedisTimeout
	redisOptions.ReadTimeout = cfg.RedisTimeout
	redisOptions.WriteTimeout = cfg.RedisTimeout
	redisOptions.MaxRetries = 1
	redisClient := redis.NewClient(redisOptions)
	defer redisClient.Close()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Warn("redis unavailable; cache will be bypassed", "error", err)
	}

	postgresRepository := postgres.NewReadingRepository(pool)
	redisCache := redisstorage.NewReadingCache(redisClient, cfg.CacheTTL)
	repository := cached.NewReadingRepository(postgresRepository, redisCache, logger)
	httpClient := &http.Client{Timeout: cfg.RequestTimeout}
	weatherCollector := collector.New(
		cfg.City,
		geocoding.NewClient(httpClient),
		openmeteo.NewClient(httpClient),
		repository,
		logger,
	)
	weatherService := service.NewWeather(repository, weatherCollector)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           httptransport.NewHandler(weatherService, logger),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	collectorCtx, stopCollector := context.WithCancel(ctx)
	defer stopCollector()
	go weatherCollector.Run(collectorCtx, cfg.PollInterval)

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("http server started", "address", cfg.HTTPAddr)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve http: %w", err)
		}
	}

	stopCollector()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown http server: %w", err)
	}
	logger.Info("application stopped gracefully")
	return nil
}
