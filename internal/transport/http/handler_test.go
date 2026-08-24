package httptransport

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/M4rSHaLll/weather-service/internal/domain"
)

type weatherServiceStub struct {
	reading domain.Reading
	err     error
	city    string
}

func (s *weatherServiceStub) Latest(_ context.Context, city string) (domain.Reading, error) {
	s.city = city
	return s.reading, s.err
}

func TestHandlerLatest(t *testing.T) {
	timestamp := time.Date(2026, time.August, 23, 12, 30, 0, 0, time.UTC)
	tests := []struct {
		name       string
		service    *weatherServiceStub
		wantStatus int
		wantBody   string
	}{
		{
			name: "returns latest reading",
			service: &weatherServiceStub{reading: domain.Reading{
				Name: "moscow", Timestamp: timestamp, Temperature: 18.5,
			}},
			wantStatus: http.StatusOK,
			wantBody:   `{"Name":"moscow","Timestamp":"2026-08-23T12:30:00Z","Temperature":18.5}`,
		},
		{
			name:       "returns not found",
			service:    &weatherServiceStub{err: domain.ErrReadingNotFound},
			wantStatus: http.StatusNotFound,
			wantBody:   "not found",
		},
		{
			name:       "returns internal error",
			service:    &weatherServiceStub{err: errors.New("database unavailable")},
			wantStatus: http.StatusInternalServerError,
			wantBody:   "internal error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			handler := NewHandler(tt.service, logger)
			request := httptest.NewRequest(http.MethodGet, "/moscow", nil)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, tt.wantStatus)
			}
			if body := strings.TrimSpace(response.Body.String()); body != tt.wantBody {
				t.Fatalf("body = %q, want %q", body, tt.wantBody)
			}
			if tt.service.city != "moscow" {
				t.Fatalf("city = %q, want moscow", tt.service.city)
			}
		})
	}
}
