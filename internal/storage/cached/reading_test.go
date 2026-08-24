package cached

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/M4rSHaLll/weather-service/internal/domain"
)

type readingStoreStub struct {
	reading     domain.Reading
	latestErr   error
	saveErr     error
	latestCalls int
	saveCalls   int
}

func (s *readingStoreStub) Latest(context.Context, string) (domain.Reading, error) {
	s.latestCalls++
	return s.reading, s.latestErr
}

func (s *readingStoreStub) Save(_ context.Context, reading domain.Reading) error {
	s.saveCalls++
	s.reading = reading
	return s.saveErr
}

func TestLatestReturnsCachedReading(t *testing.T) {
	primary := &readingStoreStub{}
	cache := &readingStoreStub{reading: domain.Reading{Name: "moscow", Temperature: 18}}
	repository := NewReadingRepository(primary, cache, discardLogger())

	reading, err := repository.Latest(context.Background(), "moscow")
	if err != nil {
		t.Fatalf("Latest() error = %v", err)
	}
	if reading != cache.reading {
		t.Fatalf("reading = %+v, want %+v", reading, cache.reading)
	}
	if primary.latestCalls != 0 {
		t.Fatalf("postgres calls = %d, want 0", primary.latestCalls)
	}
}

func TestLatestWarmsCacheAfterMiss(t *testing.T) {
	want := domain.Reading{Name: "kazan", Temperature: 20}
	primary := &readingStoreStub{reading: want}
	cache := &readingStoreStub{latestErr: domain.ErrReadingNotFound}
	repository := NewReadingRepository(primary, cache, discardLogger())

	reading, err := repository.Latest(context.Background(), "kazan")
	if err != nil {
		t.Fatalf("Latest() error = %v", err)
	}
	if reading != want || cache.reading != want {
		t.Fatalf("reading = %+v, cached = %+v, want %+v", reading, cache.reading, want)
	}
	if primary.latestCalls != 1 || cache.saveCalls != 1 {
		t.Fatalf("postgres calls = %d, cache saves = %d", primary.latestCalls, cache.saveCalls)
	}
}

func TestLatestFallsBackWhenRedisFails(t *testing.T) {
	want := domain.Reading{Name: "moscow", Temperature: 18}
	primary := &readingStoreStub{reading: want}
	cache := &readingStoreStub{latestErr: errors.New("redis unavailable"), saveErr: errors.New("redis unavailable")}
	repository := NewReadingRepository(primary, cache, discardLogger())

	reading, err := repository.Latest(context.Background(), "moscow")
	if err != nil {
		t.Fatalf("Latest() error = %v", err)
	}
	if reading != want {
		t.Fatalf("reading = %+v, want %+v", reading, want)
	}
}

func TestSaveKeepsPostgresAsSourceOfTruth(t *testing.T) {
	want := domain.Reading{Name: "moscow", Temperature: 18}
	primary := &readingStoreStub{}
	cache := &readingStoreStub{saveErr: errors.New("redis unavailable")}
	repository := NewReadingRepository(primary, cache, discardLogger())

	if err := repository.Save(context.Background(), want); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if primary.saveCalls != 1 || primary.reading != want {
		t.Fatalf("postgres did not save reading: %+v", primary.reading)
	}
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
