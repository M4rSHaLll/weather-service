package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/M4rSHaLll/weather-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReadingRepository struct {
	pool *pgxpool.Pool
}

func NewReadingRepository(pool *pgxpool.Pool) *ReadingRepository {
	return &ReadingRepository{pool: pool}
}

func (r *ReadingRepository) Latest(ctx context.Context, city string) (domain.Reading, error) {
	const query = `
		SELECT name, timestamp, temperature
		FROM reading
		WHERE name = $1
		ORDER BY timestamp DESC
		LIMIT 1`

	var reading domain.Reading
	err := r.pool.QueryRow(ctx, query, city).Scan(
		&reading.Name,
		&reading.Timestamp,
		&reading.Temperature,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Reading{}, domain.ErrReadingNotFound
	}
	if err != nil {
		return domain.Reading{}, fmt.Errorf("query latest reading: %w", err)
	}
	return reading, nil
}

func (r *ReadingRepository) Save(ctx context.Context, reading domain.Reading) error {
	const query = `
		INSERT INTO reading (name, temperature, timestamp)
		VALUES ($1, $2, $3)`

	if _, err := r.pool.Exec(ctx, query, reading.Name, reading.Temperature, reading.Timestamp); err != nil {
		return fmt.Errorf("insert reading: %w", err)
	}
	return nil
}
