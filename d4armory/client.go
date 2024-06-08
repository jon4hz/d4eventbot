package d4armory

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const baseURL = "https://d4armory.io/api"

type Client struct {
	http *http.Client
}

func New() *Client {
	c := Client{}
	c.http = &http.Client{
		Timeout: time.Second * 30,
	}
	return &c
}

func (c *Client) doRequest(ctx context.Context, endpoint, method string, expRes any) (int, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return 0, fmt.Errorf("failed to parse url: %w", err)
	}
	u = u.JoinPath(endpoint)
	req, err := http.NewRequestWithContext(ctx, method, u.String(), nil)
	if err != nil {
		return 0, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	switch resp.StatusCode {
	case 200:
		if expRes != nil {
			err = json.Unmarshal(body, expRes)
			if err != nil {
				return 0, err
			}
		}
		return resp.StatusCode, nil

	default:
		return resp.StatusCode, fmt.Errorf("%s", body)
	}
}

type RecentEvents struct {
	Boss struct {
		Name             string `json:"name"`
		ExpectedName     string `json:"expectedName"`
		NextExpectedName string `json:"nextExpectedName"`
		Timestamp        int    `json:"timestamp"`
		Expected         int    `json:"expected"`
		NextExpected     int    `json:"nextExpected"`
		Territory        string `json:"territory"`
		Zone             string `json:"zone"`
	} `json:"boss"`
	Helltide struct {
		Timestamp int    `json:"timestamp"`
		Zone      string `json:"zone"`
		Refresh   int    `json:"refresh"`
	} `json:"helltide"`
	Legion struct {
		Timestamp int    `json:"timestamp"`
		Territory string `json:"territory"`
		Zone      string `json:"zone"`
	} `json:"legion"`
}

func (c *Client) GetRecent(ctx context.Context) (*RecentEvents, error) {
	var res RecentEvents
	_, err := c.doRequest(ctx, "/events/recent", http.MethodGet, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

var ZoneMap = map[string]string{
	"step": "Dry Steppes",
	"hawe": "Hawezar",
	"frac": "Fractured Peaks",
	"kehj": "Kehjistan",
	"scos": "Scosglen", // guessed
}
