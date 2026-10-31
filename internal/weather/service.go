package weather

import (
	"context"
	"errors"
	"fmt"
)

var ErrCEPNotFound = errors.New("CEP não encontrado")

type Response struct {
	City  string  `json:"city"`
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

// CitySearcher precisa retornar o nome da cidade a partir do CEP informado, caso não encontre, deve retornar ErrCEPNotFound
type CitySearcher interface {
	Search(ctx context.Context, cep string) (string, error)
}

type TempSearcher interface {
	Search(ctx context.Context, city string) (float64, error)
}

type Service struct {
	tempSearcher TempSearcher
	citySearcher CitySearcher
}

func NewService(ts TempSearcher, cs CitySearcher) *Service {
	return &Service{
		tempSearcher: ts,
		citySearcher: cs,
	}
}

func (s *Service) Search(ctx context.Context, cep string) (Response, error) {
	city, err := s.citySearcher.Search(ctx, cep)
	if err != nil {
		return Response{}, fmt.Errorf("failed to locate city: %w", err)
	}

	temp, err := s.tempSearcher.Search(ctx, city)
	if err != nil {
		return Response{}, fmt.Errorf("failed to fetch weather data: %w", err)
	}

	return Response{
		City:  city,
		TempC: temp,
		TempF: toFahrenheit(temp),
		TempK: toKelvin(temp),
	}, nil
}

func toFahrenheit(celsius float64) float64 {
	return (celsius * 9 / 5) + 32
}

func toKelvin(celsius float64) float64 {
	return celsius + 273.15
}
