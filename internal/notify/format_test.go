package notify

import (
	"strings"
	"testing"
	"time"

	"github.com/sedyh/remna-traffic-drifter-bot/internal/panel"
)

func TestFormatMessageAfterFix(t *testing.T) {
	original := "⚠️ Drift: Wrong reset strategy\n💳 Sub: VPN-1\n🧩 Strategy: NO_RESET"
	reset := time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC)
	u := panel.User{
		UUID:                 "550e8400-e29b-41d4-a716-446655440000",
		Username:             "VPN-1",
		TrafficLimitStrategy: "MONTH",
		LastTrafficResetAt:   &reset,
		UserTraffic:          &panel.UserTraffic{UsedTrafficBytes: 0},
	}
	html := FormatMessageAfterFix(original, u, "MONTH", 35*24*time.Hour, "https://panel.example.com")
	if !strings.Contains(html, "<s>") || !strings.Contains(html, "NO_RESET</s>") {
		t.Fatalf("want strikethrough, got: %s", html)
	}
	if !strings.Contains(html, "✅ Panel now") {
		t.Fatalf("want panel block, got: %s", html)
	}
	if !strings.Contains(html, "🧩 Strategy: MONTH") {
		t.Fatalf("want updated strategy, got: %s", html)
	}
	if !strings.Contains(html, "dashboard/management/users?user=550e8400") {
		t.Fatalf("want panel link, got: %s", html)
	}
}
