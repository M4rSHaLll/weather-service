package openmeteo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type Response struct {
	Current struct {
		Time          string  `json:"time"`
		Temperature2m float64 `json:"temperature_2m"`
	} `json:"current"`
}

type Client struct {
	httpClient *http.Client
}

func NewClient(httpClient *http.Client) *Client {
	return &Client{httpClient: httpClient}
}

func (c *Client) GetTemperature(ctx context.Context, lat, long float64) (Response, error) {
	query := url.Values{
		"latitude":  {strconv.FormatFloat(lat, 'f', 6, 64)},
		"longitude": {strconv.FormatFloat(long, 'f', 6, 64)},
		"current":   {"temperature_2m"},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.open-meteo.com/v1/forecast?"+query.Encode(), nil)
	if err != nil {
		return Response{}, fmt.Errorf("create request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Response{}, fmt.Errorf("status %s", resp.Status)
	}

	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Response{}, fmt.Errorf("decode response: %w", err)
	}
	return result, nil
}
