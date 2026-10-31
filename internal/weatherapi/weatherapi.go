package weatherapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
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

func (c *Client) Search(ctx context.Context, city string) (float64, error) {

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

	return weatherResp.Current.TempC, nil
}
