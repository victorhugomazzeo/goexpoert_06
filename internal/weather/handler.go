package weather

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/victorhugomazzeo/goexpoert_06/internal/zipcode"
)

type Handler struct {
	searcher Searcher
}

type Searcher interface {
	Search(ctx context.Context, cep string) (Response, error)
}

func NewHandler(s Searcher) *Handler {
	return &Handler{
		searcher: s,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cep := r.URL.Query().Get("cep")

	if !zipcode.IsValid(cep) {
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	resp, err := h.searcher.Search(ctx, cep)

	if err != nil {
		if errors.Is(err, ErrCEPNotFound) {
			http.Error(w, "can not find zipcode", http.StatusNotFound)
			return
		}

		http.Error(w, "error occurred while fetching weather data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
