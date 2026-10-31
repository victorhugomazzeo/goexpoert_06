package gateway

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/victorhugomazzeo/goexpoert_06/internal/zipcode"
)

type WeatherGetter interface {
	Get(ctx context.Context, cep string) (*http.Response, error)
}

type Handler struct {
	weather WeatherGetter
}

func NewHandler(weatherGetter WeatherGetter) *Handler {
	return &Handler{
		weather: weatherGetter,
	}
}

type request struct {
	CEP string `json:"cep"`
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	if !zipcode.IsValid(req.CEP) {
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	resp, err := h.weather.Get(r.Context(), req.CEP)
	if err != nil {
		http.Error(w, "weather service unavailable", http.StatusBadGateway)
		slog.Error("call weather service", "error", err)

		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.WriteHeader(resp.StatusCode)

	if _, err := io.Copy(w, resp.Body); err != nil {
		slog.Error("copy weather response", "error", err)
	}
}
