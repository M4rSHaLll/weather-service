package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/M4rSHaLll/weather-service/internal/domain"
)

type ReadingRepository interface {
	Latest(context.Context, string) (domain.Reading, error)
}

type ReadingLoader interface {
	Load(context.Context, string) (domain.Reading, error)
}

type Weather struct {
	repository ReadingRepository
	loader     ReadingLoader
}

func NewWeather(repository ReadingRepository, loader ReadingLoader) *Weather {
	return &Weather{repository: repository, loader: loader}
}

func (s *Weather) Latest(ctx context.Context, city string) (domain.Reading, error) {
	reading, err := s.repository.Latest(ctx, city)
	if err == nil {
		return reading, nil
	}
	if !errors.Is(err, domain.ErrReadingNotFound) {
		return domain.Reading{}, fmt.Errorf("get latest reading for %q: %w", city, err)
	}

	reading, err = s.loader.Load(ctx, city)
	if err != nil {
		return domain.Reading{}, fmt.Errorf("load current reading for %q: %w", city, err)
	}
	return reading, nil
}
