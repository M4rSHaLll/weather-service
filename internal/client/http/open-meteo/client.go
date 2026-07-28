package open_meteo

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Responce struct {
	Current struct {
		Time          string  `json:"time"`
		Temperature2m float64 `json:"temperature_2m"`
	}
}

type client struct {
	httpClient *http.Client
}

func NewClient(httpClient *http.Client) *client {
	return &client{
		httpClient: httpClient,
	}
}

func (c *client) GetTemperature(lat, long float64) (Responce, error) {
	resp, err := c.httpClient.Get(
		fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&current=temperature_2m", lat, long),
	)
	if err != nil {
		return Responce{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Responce{}, fmt.Errorf("status %s", resp.Status)
	}

	var res Responce
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return Responce{}, err
	}

	return res, nil

}
