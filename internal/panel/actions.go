package panel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

func (c *Client) SetTrafficLimitStrategy(ctx context.Context, userID int64, strategy string) error {
	body, err := json.Marshal(map[string]any{
		"id":                   userID,
		"trafficLimitStrategy": strategy,
	})
	if err != nil {
		return err
	}
	return c.request(ctx, http.MethodPatch, c.baseURL+"/api/users", body)
}

func (c *Client) ResetTraffic(ctx context.Context, userID int64) error {
	return c.request(ctx, http.MethodPost, c.baseURL+"/api/users/"+strconv.FormatInt(userID, 10)+"/actions/reset-traffic", nil)
}

func (c *Client) FixUser(ctx context.Context, userID int64, strategy string) error {
	if err := c.SetTrafficLimitStrategy(ctx, userID, strategy); err != nil {
		return err
	}
	return c.ResetTraffic(ctx, userID)
}

func (c *Client) GetUser(ctx context.Context, userID int64) (User, error) {
	raw, err := c.requestBytes(ctx, http.MethodGet, c.baseURL+"/api/users/"+strconv.FormatInt(userID, 10), nil)
	if err != nil {
		return User{}, err
	}
	var parsed struct {
		Response User `json:"response"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return User{}, fmt.Errorf("decode user: %w", err)
	}
	return parsed.Response, nil
}

func (c *Client) request(ctx context.Context, method, url string, body []byte) error {
	_, err := c.requestBytes(ctx, method, url, body)
	return err
}

func (c *Client) requestBytes(ctx context.Context, method, url string, body []byte) ([]byte, error) {
	var r io.Reader
	if len(body) > 0 {
		r = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, r)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("panel %s %s: %s", method, url, truncate(raw, 300))
	}
	return raw, nil
}
