package gateway

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type WeatherClient struct {
	httpClient *http.Client
	baseURL    string
}

func NewWeatherClient(baseURL string, httpClient *http.Client) *WeatherClient {
	return &WeatherClient{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

func (w *WeatherClient) Get(ctx context.Context, cep string) (*http.Response, error) {
	q := url.Values{}
	q.Set("cep", cep)

	endpoint := w.baseURL + "/weather?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)

	if err != nil {
		return nil, fmt.Errorf("gateway: creating request to weather service: %w", err)

	}

	resp, err := w.httpClient.Do(req)

	if err != nil {
		return nil, fmt.Errorf("gateway: call weather service: %w", err)
	}
	return resp, nil

}
