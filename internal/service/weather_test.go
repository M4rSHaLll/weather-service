package service

import (
	"context"
	"errors"
	"testing"

	"github.com/M4rSHaLll/weather-service/internal/domain"
)

type repositoryStub struct {
	reading domain.Reading
	err     error
	city    string
}

func (s *repositoryStub) Latest(_ context.Context, city string) (domain.Reading, error) {
	s.city = city
	return s.reading, s.err
}

type loaderStub struct {
	reading domain.Reading
	err     error
	city    string
	calls   int
}

func (s *loaderStub) Load(_ context.Context, city string) (domain.Reading, error) {
	s.city = city
	s.calls++
	return s.reading, s.err
}

func TestWeatherLatestReturnsStoredReading(t *testing.T) {
	repository := &repositoryStub{reading: domain.Reading{Name: "moscow", Temperature: 18}}
	loader := &loaderStub{}
	weather := NewWeather(repository, loader)

	reading, err := weather.Latest(context.Background(), "moscow")
	if err != nil {
		t.Fatalf("Latest() error = %v", err)
	}
	if reading != repository.reading {
		t.Fatalf("reading = %+v, want %+v", reading, repository.reading)
	}
	if loader.calls != 0 {
		t.Fatalf("loader calls = %d, want 0", loader.calls)
	}
}

func TestWeatherLatestLoadsMissingCity(t *testing.T) {
	repository := &repositoryStub{err: domain.ErrReadingNotFound}
	loader := &loaderStub{reading: domain.Reading{Name: "kazan", Temperature: 20}}
	weather := NewWeather(repository, loader)

	reading, err := weather.Latest(context.Background(), "kazan")
	if err != nil {
		t.Fatalf("Latest() error = %v", err)
	}
	if reading != loader.reading {
		t.Fatalf("reading = %+v, want %+v", reading, loader.reading)
	}
	if loader.calls != 1 || loader.city != "kazan" {
		t.Fatalf("loader calls = %d, city = %q", loader.calls, loader.city)
	}
}

func TestWeatherLatestDoesNotLoadOnDatabaseError(t *testing.T) {
	databaseErr := errors.New("database unavailable")
	repository := &repositoryStub{err: databaseErr}
	loader := &loaderStub{}
	weather := NewWeather(repository, loader)

	_, err := weather.Latest(context.Background(), "kazan")
	if !errors.Is(err, databaseErr) {
		t.Fatalf("Latest() error = %v, want %v", err, databaseErr)
	}
	if loader.calls != 0 {
		t.Fatalf("loader calls = %d, want 0", loader.calls)
	}
}
