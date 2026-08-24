package collector

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/M4rSHaLll/weather-service/internal/client/http/geocoding"
	"github.com/M4rSHaLll/weather-service/internal/client/http/openmeteo"
	"github.com/M4rSHaLll/weather-service/internal/domain"
)

type Geocoder interface {
	GetCoords(context.Context, string) (geocoding.Response, error)
}

type TemperatureProvider interface {
	GetTemperature(context.Context, float64, float64) (openmeteo.Response, error)
}

type ReadingWriter interface {
	Save(context.Context, domain.Reading) error
}

type Collector struct {
	city       string
	geocoder   Geocoder
	weather    TemperatureProvider
	repository ReadingWriter
	logger     *slog.Logger
}

func New(city string, geocoder Geocoder, weather TemperatureProvider, repository ReadingWriter, logger *slog.Logger) *Collector {
	return &Collector{city: city, geocoder: geocoder, weather: weather, repository: repository, logger: logger}
}

func (c *Collector) Collect(ctx context.Context) error {
	_, err := c.Load(ctx, c.city)
	return err
}

func (c *Collector) Load(ctx context.Context, city string) (domain.Reading, error) {
	coords, err := c.geocoder.GetCoords(ctx, city)
	if err != nil {
		return domain.Reading{}, fmt.Errorf("get coordinates: %w", err)
	}

	weather, err := c.weather.GetTemperature(ctx, coords.Latitude, coords.Longitude)
	if err != nil {
		return domain.Reading{}, fmt.Errorf("get temperature: %w", err)
	}

	timestamp, err := time.Parse("2006-01-02T15:04", weather.Current.Time)
	if err != nil {
		return domain.Reading{}, fmt.Errorf("parse observation time: %w", err)
	}

	reading := domain.Reading{
		Name:        city,
		Timestamp:   timestamp,
		Temperature: weather.Current.Temperature2m,
	}
	if err := c.repository.Save(ctx, reading); err != nil {
		return domain.Reading{}, err
	}

	c.logger.Info("weather data updated", "city", city)
	return reading, nil
}

func (c *Collector) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := c.Collect(ctx); err != nil {
				c.logger.Error("collect weather data", "city", c.city, "error", err)
			}
		}
	}
}
