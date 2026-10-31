package weather

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeWeatherSearcher struct {
	response Response
	err      error
}

func (f *fakeWeatherSearcher) Search(_ context.Context, _ string) (Response, error) {
	return f.response, f.err
}

func TestHandler_ServeHTTP(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		ws := &fakeWeatherSearcher{
			response: Response{
				City:  "São Paulo",
				TempC: 28.5,
				TempF: 83.3,
				TempK: 301.65,
			},
		}

		handler := NewHandler(ws)
		req := httptest.NewRequest(http.MethodGet, "/weather?cep=12345678", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Weather() status = %d, want %d", rec.Code, http.StatusOK)
		}
		if got := rec.Header().Get("Content-Type"); got != "application/json" {
			t.Fatalf("Weather() Content-Type = %s, want %s", got, "application/json")
		}
		var got Response
		if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
			t.Fatalf("Weather() failed to decode response body: %v", err)
		}

		want := Response{City: "São Paulo", TempC: 28.5, TempF: 83.3, TempK: 301.65}

		if got != want {
			t.Errorf("ServeHTTP() response = %+v, want %+v", got, want)
		}
	})

	tests := []struct {
		name       string
		cep        string
		searchErr  error
		wantStatus int
		wantBody   string
	}{

		{name: "invalid cep", cep: "1234", wantStatus: http.StatusUnprocessableEntity, wantBody: "invalid zipcode"},
		{name: "cep not found", cep: "87654321", searchErr: ErrCEPNotFound, wantStatus: http.StatusNotFound, wantBody: "can not find zipcode"},
		{name: "internal server error", cep: "12345678", searchErr: errors.New("some error"), wantStatus: http.StatusInternalServerError, wantBody: "error occurred while fetching weather data"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &fakeWeatherSearcher{err: tt.searchErr}
			handler := NewHandler(ws)
			req := httptest.NewRequest(http.MethodGet, "/weather?cep="+tt.cep, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("Weather() status = %d, want %d", rec.Code, tt.wantStatus)
			}

			gotBody := strings.TrimSpace(rec.Body.String())

			if gotBody != tt.wantBody {
				t.Errorf("Weather() body = %q, want %q", gotBody, tt.wantBody)
			}
		})
	}

}
