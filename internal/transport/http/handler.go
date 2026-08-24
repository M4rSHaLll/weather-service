package httptransport

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/M4rSHaLll/weather-service/internal/domain"
)

type WeatherService interface {
	Latest(context.Context, string) (domain.Reading, error)
}

type Handler struct {
	weather WeatherService
	logger  *slog.Logger
}

func NewHandler(weather WeatherService, logger *slog.Logger) http.Handler {
	h := &Handler{weather: weather, logger: logger}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{city}", h.latest)
	return requestLogger(logger, mux)
}

func (h *Handler) latest(w http.ResponseWriter, r *http.Request) {
	reading, err := h.weather.Latest(r.Context(), r.PathValue("city"))
	if errors.Is(err, domain.ErrReadingNotFound) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		h.logger.Error("get latest weather", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(reading); err != nil {
		h.logger.Error("encode response", "error", err)
	}
}

func requestLogger(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("http request", "method", r.Method, "path", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
