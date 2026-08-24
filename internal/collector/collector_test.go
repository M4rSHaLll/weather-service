package collector

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/M4rSHaLll/weather-service/internal/client/http/geocoding"
	"github.com/M4rSHaLll/weather-service/internal/client/http/openmeteo"
	"github.com/M4rSHaLll/weather-service/internal/domain"
)

type geocoderStub struct {
	response geocoding.Response
	city     string
}

func (s *geocoderStub) GetCoords(_ context.Context, city string) (geocoding.Response, error) {
	s.city = city
	return s.response, nil
}

type temperatureProviderStub struct {
	response openmeteo.Response
	lat      float64
	long     float64
}

func (s *temperatureProviderStub) GetTemperature(_ context.Context, lat, long float64) (openmeteo.Response, error) {
	s.lat, s.long = lat, long
	return s.response, nil
}

type writerStub struct {
	reading domain.Reading
}

func (s *writerStub) Save(_ context.Context, reading domain.Reading) error {
	s.reading = reading
	return nil
}

func TestCollectorCollect(t *testing.T) {
	weather := &temperatureProviderStub{}
	weather.response.Current.Time = "2026-08-23T12:30"
	weather.response.Current.Temperature2m = 18.5
	writer := &writerStub{}
	geocoder := &geocoderStub{response: geocoding.Response{Latitude: 55.75, Longitude: 37.62}}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	collector := New(
		"moscow",
		geocoder,
		weather,
		writer,
		logger,
	)

	if err := collector.Collect(context.Background()); err != nil {
		t.Fatalf("Collect() error = %v", err)
	}

	if weather.lat != 55.75 || weather.long != 37.62 {
		t.Fatalf("coordinates = (%v, %v), want (55.75, 37.62)", weather.lat, weather.long)
	}
	if geocoder.city != "moscow" {
		t.Fatalf("geocoding city = %q, want moscow", geocoder.city)
	}
	wantTime := time.Date(2026, time.August, 23, 12, 30, 0, 0, time.UTC)
	if writer.reading.Name != "moscow" || writer.reading.Temperature != 18.5 || !writer.reading.Timestamp.Equal(wantTime) {
		t.Fatalf("saved reading = %+v", writer.reading)
	}
}

func TestCollectorLoadRequestedCity(t *testing.T) {
	weather := &temperatureProviderStub{}
	weather.response.Current.Time = "2026-08-23T12:30"
	weather.response.Current.Temperature2m = 21
	writer := &writerStub{}
	geocoder := &geocoderStub{response: geocoding.Response{Latitude: 59.93, Longitude: 30.31}}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	collector := New("moscow", geocoder, weather, writer, logger)

	reading, err := collector.Load(context.Background(), "saint-petersburg")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if geocoder.city != "saint-petersburg" {
		t.Fatalf("geocoding city = %q, want saint-petersburg", geocoder.city)
	}
	if reading.Name != "saint-petersburg" || writer.reading.Name != "saint-petersburg" {
		t.Fatalf("reading was not saved for requested city: %+v", writer.reading)
	}
}
