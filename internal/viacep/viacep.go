package viacep

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/victorhugomazzeo/goexpoert_06/internal/weather"
)

type response struct {
	Localidade string `json:"localidade"`
	Erro       string `json:"erro"`
}

type Client struct {
	client  *http.Client
	baseURL string
}

const baseURL = "https://viacep.com.br/ws/"

func NewClient(client *http.Client) *Client {
	return &Client{
		client:  client,
		baseURL: baseURL,
	}
}

func (c *Client) Search(ctx context.Context, cep string) (string, error) {

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+cep+"/json/", nil)
	if err != nil {
		return "", err
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("viaCEP: failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("viaCEP: error while fetching data: %s", resp.Status)
	}

	var viacepResp response
	err = json.NewDecoder(resp.Body).Decode(&viacepResp)
	if err != nil {
		return "", err
	}

	if viacepResp.Erro == "true" {
		return "", weather.ErrCEPNotFound
	}

	return viacepResp.Localidade, nil
}
