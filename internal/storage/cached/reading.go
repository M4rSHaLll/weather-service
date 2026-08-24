package cached

import (
	"context"
	"errors"
	"log/slog"

	"github.com/M4rSHaLll/weather-service/internal/domain"
)

type ReadingStore interface {
	Latest(context.Context, string) (domain.Reading, error)
	Save(context.Context, domain.Reading) error
}

type ReadingRepository struct {
	primary ReadingStore
	cache   ReadingStore
	logger  *slog.Logger
}

func NewReadingRepository(primary, cache ReadingStore, logger *slog.Logger) *ReadingRepository {
	return &ReadingRepository{primary: primary, cache: cache, logger: logger}
}

func (r *ReadingRepository) Latest(ctx context.Context, city string) (domain.Reading, error) {
	reading, err := r.cache.Latest(ctx, city)
	if err == nil {
		return reading, nil
	}
	if !errors.Is(err, domain.ErrReadingNotFound) {
		r.logger.Warn("redis read failed; using postgres", "city", city, "error", err)
	}

	reading, err = r.primary.Latest(ctx, city)
	if err != nil {
		return domain.Reading{}, err
	}
	if err := r.cache.Save(ctx, reading); err != nil {
		r.logger.Warn("redis warm-up failed", "city", city, "error", err)
	}
	return reading, nil
}

func (r *ReadingRepository) Save(ctx context.Context, reading domain.Reading) error {
	if err := r.primary.Save(ctx, reading); err != nil {
		return err
	}
	if err := r.cache.Save(ctx, reading); err != nil {
		r.logger.Warn("redis update failed", "city", reading.Name, "error", err)
	}
	return nil
}
