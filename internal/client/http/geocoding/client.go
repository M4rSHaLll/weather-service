package geocoding

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

type Response struct {
	Name      string  `json:"name"`
	Country   string  `json:"country"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Client struct {
	httpClient *http.Client
}

func NewClient(httpClient *http.Client) *Client {
	return &Client{
		httpClient: httpClient,
	}
}

func (c *Client) GetCoords(ctx context.Context, city string) (Response, error) {
	query := url.Values{"name": {city}, "count": {"1"}, "language": {"ru"}, "format": {"json"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://geocoding-api.open-meteo.com/v1/search?"+query.Encode(), nil)
	if err != nil {
		return Response{}, fmt.Errorf("create request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("send request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Response{}, errors.New(resp.Status)
	}

	var geoResp struct {
		Results []Response `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&geoResp); err != nil {
		return Response{}, fmt.Errorf("decode response: %w", err)
	}
	if len(geoResp.Results) == 0 {
		return Response{}, errors.New("city not found")
	}

	return geoResp.Results[0], nil
}
