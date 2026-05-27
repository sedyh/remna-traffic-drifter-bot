package panel

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type UserTraffic struct {
	UsedTrafficBytes int64 `json:"usedTrafficBytes"`
}

type internalSquad struct {
	Name string `json:"name"`
}

type User struct {
	UUID                   string          `json:"uuid"`
	Username               string          `json:"username"`
	Status                 string          `json:"status"`
	Tag                    string          `json:"tag"`
	TrafficLimitBytes      int64           `json:"trafficLimitBytes"`
	TrafficLimitStrategy   string          `json:"trafficLimitStrategy"`
	CreatedAt              *time.Time      `json:"createdAt"`
	LastTrafficResetAt     *time.Time      `json:"lastTrafficResetAt"`
	UserTraffic            *UserTraffic    `json:"userTraffic"`
	ActiveInternalSquads   []internalSquad `json:"activeInternalSquads"`
}

func (u User) SquadNames() string {
	if len(u.ActiveInternalSquads) == 0 {
		return ""
	}
	names := make([]string, 0, len(u.ActiveInternalSquads))
	for _, s := range u.ActiveInternalSquads {
		n := strings.TrimSpace(s.Name)
		if n != "" {
			names = append(names, n)
		}
	}
	return strings.Join(names, ", ")
}

func (u User) SkipUnlimited() bool {
	return u.TrafficLimitBytes == 0
}

func (u User) HasUsedTraffic() bool {
	return u.UserTraffic != nil && u.UserTraffic.UsedTrafficBytes > 0
}

type usersPage struct {
	Total int    `json:"total"`
	Users []User `json:"users"`
}

type usersResponse struct {
	Response usersPage `json:"response"`
}

func (c *Client) ListUsers(ctx context.Context, start, size int) (usersPage, error) {
	u, err := url.Parse(c.baseURL + "/api/users")
	if err != nil {
		return usersPage{}, err
	}
	q := u.Query()
	q.Set("start", fmt.Sprintf("%d", start))
	q.Set("size", fmt.Sprintf("%d", size))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return usersPage{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return usersPage{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return usersPage{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return usersPage{}, fmt.Errorf("panel GET /api/users: %s: %s", resp.Status, truncate(body, 300))
	}

	var parsed usersResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return usersPage{}, fmt.Errorf("decode users: %w", err)
	}
	return parsed.Response, nil
}

func (c *Client) ListAllUsers(ctx context.Context, pageSize int) ([]User, error) {
	var all []User
	start := 0
	var total int

	for {
		page, err := c.ListUsers(ctx, start, pageSize)
		if err != nil {
			return nil, err
		}
		if total == 0 {
			total = page.Total
		}
		if len(page.Users) == 0 {
			break
		}
		all = append(all, page.Users...)
		start += len(page.Users)
		if start >= total {
			break
		}
	}
	return all, nil
}

func truncate(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "..."
}
