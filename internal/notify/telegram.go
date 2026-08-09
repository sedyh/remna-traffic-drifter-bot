package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/sedyh/remna-traffic-drifter-bot/internal/check"
)

const maxSendRetries = 5

type TelegramChat struct {
	ChatID   string
	ThreadID int
}

type Telegram struct {
	token        string
	chats        []TelegramChat
	sendInterval time.Duration
	panelBaseURL string
	client       *http.Client

	mu       sync.Mutex
	lastSend time.Time
}

func NewTelegram(token string, chats []TelegramChat, sendInterval time.Duration, proxyURL, panelBaseURL string) *Telegram {
	if sendInterval <= 0 {
		sendInterval = 3 * time.Second
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if proxyURL != "" {
		u, err := url.Parse(proxyURL)
		if err == nil {
			transport.Proxy = http.ProxyURL(u)
		}
	}
	return &Telegram{
		token:        token,
		chats:        chats,
		sendInterval: sendInterval,
		panelBaseURL: strings.TrimSuffix(strings.TrimSpace(panelBaseURL), "/"),
		client: &http.Client{
			Timeout:   60 * time.Second,
			Transport: transport,
		},
	}
}

type sendMessageReq struct {
	ChatID          string          `json:"chat_id"`
	MessageThreadID int             `json:"message_thread_id,omitempty"`
	Text            string          `json:"text"`
	ParseMode       string          `json:"parse_mode,omitempty"`
	ReplyMarkup     *inlineKeyboard `json:"reply_markup,omitempty"`
}

type telegramAPIError struct {
	OK          bool   `json:"ok"`
	ErrorCode   int    `json:"error_code"`
	Description string `json:"description"`
	Parameters  struct {
		RetryAfter int `json:"retry_after"`
	} `json:"parameters"`
}

func (t *Telegram) SendIssue(ctx context.Context, issue check.Issue) error {
	text := formatIssueFromCheck(issue, t.panelBaseURL)
	var errs []string
	for _, chat := range t.chats {
		if err := t.send(ctx, chat.ChatID, chat.ThreadID, text, issueKeyboard(issue)); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("telegram: %s", strings.Join(errs, "; "))
	}
	return nil
}

func formatStrategy(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "?"
	}
	return s
}

func formatIssueFromCheck(i check.Issue, panelBase string) string {
	return formatIssueLines(i.Kind, issueFields{
		username:         i.Username,
		userID:           i.UserID,
		panelBase:        panelBase,
		tag:              i.Tag,
		status:           i.Status,
		squad:            i.Squad,
		trafficLimit:     i.TrafficLimit,
		actualStrategy:   i.ActualStrategy,
		expectedStrategy: i.ExpectedStrategy,
		lastResetAt:      i.LastResetAt,
		checkedAt:        i.CheckedAt,
	})
}

func formatIssueLines(kind string, i issueFields) string {
	title := "⚠️ Drift: alert"
	switch kind {
	case check.IssueWrongStrategy:
		title = "⚠️ Drift: Wrong reset strategy"
	case check.IssueStaleReset:
		title = "⚠️ Drift: Stale traffic reset"
	}

	lines := []string{
		title,
		formatSubLine(i.username, i.userID, i.panelBase),
		"🏷 Tariff tag: " + escapeHTML(i.tag),
	}
	if i.status != "" {
		lines = append(lines, "📋 Status: "+escapeHTML(i.status))
	}
	if i.squad != "" {
		lines = append(lines, "📦 Squad: "+escapeHTML(i.squad))
	}
	if i.trafficLimit > 0 {
		lines = append(lines, "📊 Limit: "+escapeHTML(formatBytes(i.trafficLimit)))
	}
	lines = append(lines,
		"🧩 Strategy: "+escapeHTML(formatStrategy(i.actualStrategy)),
		"🧩 Expected: "+escapeHTML(formatStrategy(i.expectedStrategy)),
	)
	if kind == check.IssueStaleReset {
		if i.lastResetAt == nil {
			lines = append(lines, "🧩 Last reset: Never reset")
		} else {
			age := i.checkedAt.Sub(*i.lastResetAt)
			lines = append(lines, "🧩 Last reset: "+escapeHTML(humanAgo(age)+", at "+formatWhen(*i.lastResetAt)))
		}
	}
	return strings.Join(lines, "\n")
}

func formatWhen(t time.Time) string {
	return t.UTC().Format("2006-01-02 15:04:05")
}

func humanAgo(d time.Duration) string {
	if d < 0 {
		d = -d
	}
	h := int(d.Round(time.Hour).Hours())
	if h >= 1 {
		return fmt.Sprintf("%dh ago", h)
	}
	m := int(d.Round(time.Minute).Minutes())
	if m >= 1 {
		return fmt.Sprintf("%dm ago", m)
	}
	return fmt.Sprintf("%ds ago", int(d.Round(time.Second).Seconds()))
}

func (t *Telegram) pace(ctx context.Context) error {
	t.mu.Lock()
	wait := t.sendInterval - time.Since(t.lastSend)
	t.mu.Unlock()
	if wait <= 0 {
		return nil
	}
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (t *Telegram) markSent() {
	t.mu.Lock()
	t.lastSend = time.Now()
	t.mu.Unlock()
}

func parseRetryAfter(raw []byte) time.Duration {
	var apiErr telegramAPIError
	if err := json.Unmarshal(raw, &apiErr); err != nil {
		return 0
	}
	if apiErr.Parameters.RetryAfter <= 0 {
		return 0
	}
	return time.Duration(apiErr.Parameters.RetryAfter) * time.Second
}

func (t *Telegram) waitRetry(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		d = t.sendInterval
	}
	d += time.Second
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (t *Telegram) send(ctx context.Context, chatID string, threadID int, text string, keyboard *inlineKeyboard) error {
	reqBody := sendMessageReq{
		ChatID:      chatID,
		Text:        text,
		ParseMode:   "HTML",
		ReplyMarkup: keyboard,
	}
	if threadID > 0 {
		reqBody.MessageThreadID = threadID
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.token)

	for attempt := 0; attempt < maxSendRetries; attempt++ {
		if err := t.pace(ctx); err != nil {
			return err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := t.client.Do(req)
		if err != nil {
			return err
		}

		raw, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			t.markSent()
			return nil
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			if err := t.waitRetry(ctx, parseRetryAfter(raw)); err != nil {
				return err
			}
			continue
		}

		return fmt.Errorf("chat %s: %s %s", chatID, resp.Status, string(raw))
	}

	return fmt.Errorf("chat %s: rate limit, gave up after %d retries", chatID, maxSendRetries)
}
