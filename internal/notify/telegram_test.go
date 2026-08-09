package notify

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/sedyh/remna-traffic-drifter-bot/internal/check"
)

func TestParseRetryAfter(t *testing.T) {
	raw := []byte(`{"ok":false,"error_code":429,"description":"Too Many Requests","parameters":{"retry_after":41}}`)
	if got := parseRetryAfter(raw); got != 41*time.Second {
		t.Fatalf("want 41s, got %v", got)
	}
}

func TestSendMessageReqThreadID(t *testing.T) {
	req := sendMessageReq{ChatID: "-1001", Text: "hi", MessageThreadID: 525}
	raw, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"message_thread_id":525`) {
		t.Fatalf("unexpected: %s", raw)
	}
}

func TestFormatIssueWrongStrategy(t *testing.T) {
	msg := formatIssueFromCheck(check.Issue{
		Kind:             check.IssueWrongStrategy,
		UserID:           42,
		Username:         "VPN-123059",
		Tag:              "TAG_TARIFF_MIN",
		ActualStrategy:   "NO_RESET",
		ExpectedStrategy: "MONTH",
	}, "https://panel.example.com")
	if !strings.Contains(msg, "Wrong reset strategy") {
		t.Fatalf("unexpected: %s", msg)
	}
	if !strings.Contains(msg, "🧩 Strategy: NO_RESET") {
		t.Fatalf("unexpected: %s", msg)
	}
	if !strings.Contains(msg, "🏷 Tariff tag: TAG_TARIFF_MIN") {
		t.Fatalf("unexpected: %s", msg)
	}
	if !strings.Contains(msg, `href="https://panel.example.com/dashboard/management/users?user=42"`) {
		t.Fatalf("want panel link, got: %s", msg)
	}
}

func TestFormatIssueStaleReset(t *testing.T) {
	now := time.Date(2026, 5, 26, 0, 5, 0, 0, time.UTC)
	reset := now.Add(-72 * time.Hour)
	msg := formatIssueFromCheck(check.Issue{
		Kind:             check.IssueStaleReset,
		UserID:           42,
		Username:         "VPN-123059",
		Tag:              "TAG_TARIFF_MIN",
		ActualStrategy:   "MONTH",
		ExpectedStrategy: "MONTH",
		LastResetAt:      &reset,
		StaleAfter:       35 * 24 * time.Hour,
		CheckedAt:        now,
	}, "https://panel.example.com")
	if !strings.Contains(msg, "🧩 Strategy: MONTH") {
		t.Fatalf("want strategy line, got: %s", msg)
	}
	if !strings.Contains(msg, "🧩 Expected: MONTH") {
		t.Fatalf("want expected strategy, got: %s", msg)
	}
	if !strings.Contains(msg, "🧩 Last reset: 72h ago, at 2026-05-23 00:05:00") {
		t.Fatalf("want last reset line, got: %s", msg)
	}
}
