package weatherapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

const baseURL = "https://api.weatherapi.com/v1/"

type Client struct {
	apiKey  string
	client  *http.Client
	baseURL string
}

type response struct {
	Current struct {
		TempC float64 `json:"temp_c"`
	} `json:"current"`
}

func NewClient(apiKey string, client *http.Client) *Client {
	return &Client{
		apiKey:  apiKey,
		client:  client,
		baseURL: baseURL,
	}
}

var tracer = otel.Tracer("github.com/victorhugomazzeo/goexpoert_06/internal/weatherapi")

func (c *Client) Search(ctx context.Context, city string) (temp float64, err error) {

	ctx, span := tracer.Start(ctx, "busca-clima")
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	span.SetAttributes(attribute.String("city", city))

	q := url.Values{}
	q.Set("key", c.apiKey)
	q.Set("q", city)

	endpoint := c.baseURL + "current.json?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)

	if err != nil {
		return 0, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			return 0, fmt.Errorf("weatherapi: failed to make request: %w", urlErr.Err)
		}
		return 0, fmt.Errorf("weatherapi: failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("weatherapi: error while fetching data: %s", resp.Status)
	}

	var weatherResp response
	err = json.NewDecoder(resp.Body).Decode(&weatherResp)
	if err != nil {
		return 0, err
	}

	span.SetAttributes(attribute.Float64("temp_c", weatherResp.Current.TempC))

	return weatherResp.Current.TempC, nil
}
