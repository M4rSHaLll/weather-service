package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	defaultHTTPAddr       = ":3000"
	defaultCity           = "moscow"
	defaultPollInterval   = 30 * time.Minute
	defaultRequestTimeout = 10 * time.Second
	defaultCacheTTL       = 2 * time.Minute
	defaultRedisTimeout   = 500 * time.Millisecond
)

type Config struct {
	HTTPAddr        string
	DatabaseURL     string
	RedisURL        string
	CacheTTL        time.Duration
	RedisTimeout    time.Duration
	City            string
	PollInterval    time.Duration
	RequestTimeout  time.Duration
	ShutdownTimeout time.Duration
}

func Load() (Config, error) {
	if err := loadDotEnv(".env"); err != nil {
		return Config{}, err
	}

	databaseURL, err := requiredEnv("DATABASE_URL")
	if err != nil {
		return Config{}, err
	}
	redisURL, err := requiredEnv("REDIS_URL")
	if err != nil {
		return Config{}, err
	}
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

	cacheTTL, err := durationFromEnv("CACHE_TTL", defaultCacheTTL)
	if err != nil {
		return Config{}, err
	}
	redisTimeout, err := durationFromEnv("REDIS_TIMEOUT", defaultRedisTimeout)
	if err != nil {
		return Config{}, err
	}

	return Config{
		HTTPAddr:        valueOrDefault("HTTP_ADDR", defaultHTTPAddr),
		DatabaseURL:     databaseURL,
		RedisURL:        redisURL,
		CacheTTL:        cacheTTL,
		RedisTimeout:    redisTimeout,
		City:            valueOrDefault("WEATHER_CITY", defaultCity),
		PollInterval:    pollInterval,
		RequestTimeout:  requestTimeout,
		ShutdownTimeout: shutdownTimeout,
	}, nil
}

func requiredEnv(name string) (string, error) {
	value := os.Getenv(name)
	if value == "" {
		return "", fmt.Errorf("%s is required", name)
	}
	return value, nil
}

func loadDotEnv(path string) error {
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		name, value, ok := strings.Cut(line, "=")
		if !ok {
			return fmt.Errorf("parse %s: invalid line %q", path, line)
		}
		name = strings.TrimSpace(name)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if name == "" {
			return fmt.Errorf("parse %s: empty variable name", path)
		}
		if _, exists := os.LookupEnv(name); !exists {
			if err := os.Setenv(name, value); err != nil {
				return fmt.Errorf("set %s: %w", name, err)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	return nil
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
