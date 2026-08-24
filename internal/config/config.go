package config

import (
	"fmt"
	"os"
	"time"
)

const (
	defaultHTTPAddr       = ":3000"
	defaultDatabaseURL    = "postgresql://postgres:a864653K@localhost:54321/weather"
	defaultCity           = "moscow"
	defaultPollInterval   = 30 * time.Minute
	defaultRequestTimeout = 10 * time.Second
)

type Config struct {
	HTTPAddr        string
	DatabaseURL     string
	City            string
	PollInterval    time.Duration
	RequestTimeout  time.Duration
	ShutdownTimeout time.Duration
}

func Load() (Config, error) {
	pollInterval, err := durationFromEnv("WEATHER_POLL_INTERVAL", defaultPollInterval)
	if err != nil {
		return Config{}, err
	}
	requestTimeout, err := durationFromEnv("HTTP_CLIENT_TIMEOUT", defaultRequestTimeout)
	if err != nil {
		return Config{}, err
	}
	shutdownTimeout, err := durationFromEnv("SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}

	return Config{
		HTTPAddr:        valueOrDefault("HTTP_ADDR", defaultHTTPAddr),
		DatabaseURL:     valueOrDefault("DATABASE_URL", defaultDatabaseURL),
		City:            valueOrDefault("WEATHER_CITY", defaultCity),
		PollInterval:    pollInterval,
		RequestTimeout:  requestTimeout,
		ShutdownTimeout: shutdownTimeout,
	}, nil
}

func valueOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func durationFromEnv(name string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(name)
	if value == "" {
		return fallback, nil
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", name, err)
	}
	if duration <= 0 {
		return 0, fmt.Errorf("%s must be positive", name)
	}
	return duration, nil
}
