package weather

import (
	"context"
	"errors"
	"math"
	"testing"
)

type fakeTempSearcher struct {
	temp float64
	err  error
}

type fakeCitySearcher struct {
	city string
	err  error
}

func (f *fakeCitySearcher) Search(_ context.Context, _ string) (string, error) {
	return f.city, f.err
}

func (f *fakeTempSearcher) Search(_ context.Context, _ string) (float64, error) {
	return f.temp, f.err
}

func TestToFahrenheit(t *testing.T) {
	tests := []struct {
		name    string
		celsius float64
		want    float64
	}{
		{name: "zero", celsius: 0, want: 32},
		{name: "positive", celsius: 100, want: 212},
		{name: "negative", celsius: -40, want: -40},
		{name: "decimal from API contract", celsius: 28.5, want: 83.3},
		{name: "body temperature", celsius: 36.6, want: 97.88},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toFahrenheit(tt.celsius)
			if !almostEqual(got, tt.want) {
				t.Errorf("toFahrenheit(%v) = %v, want %v", tt.celsius, got, tt.want)
			}
		})
	}
}

func TestToKelvin(t *testing.T) {
	tests := []struct {
		name    string
		celsius float64
		want    float64
	}{
		{name: "zero", celsius: 0, want: 273.15},
		{name: "positive", celsius: 100, want: 373.15},
		{name: "negative at absolute zero", celsius: -273.15, want: 0},
		{name: "decimal from API contract", celsius: 28.5, want: 301.65},
		{name: "body temperature", celsius: 36.6, want: 309.75},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toKelvin(tt.celsius)
			if !almostEqual(got, tt.want) {
				t.Errorf("toKelvin(%v) = %v, want %v", tt.celsius, got, tt.want)
			}
		})
	}
}

func almostEqual(a, b float64) bool {
	const tolerance = 1e-9
	return math.Abs(a-b) < tolerance
}

func TestService_Search(t *testing.T) {
	cs := &fakeCitySearcher{city: "São Paulo"}
	ts := &fakeTempSearcher{temp: 28.5}
	f := 83.3
	k := 301.65

	s := NewService(ts, cs)
	resp, err := s.Search(t.Context(), "01001000")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !almostEqual(resp.TempC, 28.5) {
		t.Errorf("Search().TempC = %v, want %v", resp.TempC, 28.5)
	}
	if !almostEqual(resp.TempF, f) {
		t.Errorf("Search().TempF = %v, want %v", resp.TempF, f)
	}
	if !almostEqual(resp.TempK, k) {
		t.Errorf("Search().TempK = %v, want %v", resp.TempK, k)
	}
	if resp.City != "São Paulo" {
		t.Errorf("Search().City = %q, want %q", resp.City, "São Paulo")
	}
}

func TestWeather_Search_Errors(t *testing.T) {
	tempErr := errors.New("temp searcher error")

	tests := []struct {
		name    string
		cityErr error
		tempErr error
		wantErr error
	}{
		{name: "city not found", cityErr: ErrCEPNotFound, wantErr: ErrCEPNotFound},
		{name: "temp searcher error", tempErr: tempErr, wantErr: tempErr},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &fakeTempSearcher{err: tt.tempErr}
			cs := &fakeCitySearcher{err: tt.cityErr}
			s := NewService(ts, cs)
			_, err := s.Search(t.Context(), "01001000")
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Search() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
