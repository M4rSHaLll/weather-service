package redisstorage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/M4rSHaLll/weather-service/internal/domain"
	"github.com/redis/go-redis/v9"
)

type ReadingCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewReadingCache(client *redis.Client, ttl time.Duration) *ReadingCache {
	return &ReadingCache{client: client, ttl: ttl}
}

func (c *ReadingCache) Latest(ctx context.Context, city string) (domain.Reading, error) {
	raw, err := c.client.Get(ctx, key(city)).Bytes()
	if errors.Is(err, redis.Nil) {
		return domain.Reading{}, domain.ErrReadingNotFound
	}
	if err != nil {
		return domain.Reading{}, fmt.Errorf("get cached reading: %w", err)
	}

	var reading domain.Reading
	if err := json.Unmarshal(raw, &reading); err != nil {
		return domain.Reading{}, fmt.Errorf("decode cached reading: %w", err)
	}
	return reading, nil
}

func (c *ReadingCache) Save(ctx context.Context, reading domain.Reading) error {
	raw, err := json.Marshal(reading)
	if err != nil {
		return fmt.Errorf("encode cached reading: %w", err)
	}
	if err := c.client.Set(ctx, key(reading.Name), raw, c.ttl).Err(); err != nil {
		return fmt.Errorf("cache reading: %w", err)
	}
	return nil
}

func key(city string) string {
	return "weather:latest:" + city
}
