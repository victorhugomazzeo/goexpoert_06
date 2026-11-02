package viacep

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/victorhugomazzeo/goexpoert_06/internal/weather"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type response struct {
	Localidade string `json:"localidade"`
	Erro       string `json:"erro"`
}

type Client struct {
	client  *http.Client
	baseURL string
}

var tracer = otel.Tracer("github.com/victorhugomazzeo/goexpoert_06/internal/viacep")

const baseURL = "https://viacep.com.br/ws/"

func NewClient(client *http.Client) *Client {
	return &Client{
		client:  client,
		baseURL: baseURL,
	}
}

func (c *Client) Search(ctx context.Context, cep string) (city string, err error) {

	ctx, span := tracer.Start(ctx, "busca-cep")
	defer func() {
		if err != nil && !errors.Is(err, weather.ErrCEPNotFound) {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	span.SetAttributes(attribute.String("cep", cep))

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

	span.SetAttributes(attribute.String("city", viacepResp.Localidade))

	return viacepResp.Localidade, nil
}
