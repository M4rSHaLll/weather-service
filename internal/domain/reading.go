package domain

import (
	"errors"
	"time"
)

var ErrReadingNotFound = errors.New("reading not found")

type Reading struct {
	Name        string
	Timestamp   time.Time
	Temperature float64
}
