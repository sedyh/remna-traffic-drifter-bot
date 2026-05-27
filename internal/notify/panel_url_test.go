package notify

import (
	"strings"
	"testing"
)

func TestUserAdminURL(t *testing.T) {
	got := UserAdminURL("https://panel.example.com/", "550e8400-e29b-41d4-a716-446655440000")
	want := "https://panel.example.com/dashboard/management/users?user=550e8400-e29b-41d4-a716-446655440000"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestFormatSubLineLink(t *testing.T) {
	line := formatSubLine("VPN-1", "550e8400-e29b-41d4-a716-446655440000", "https://panel.example.com")
	want := `href="https://panel.example.com/dashboard/management/users?user=550e8400-e29b-41d4-a716-446655440000"`
	if !strings.Contains(line, want) || !strings.Contains(line, ">VPN-1</a>") {
		t.Fatalf("unexpected: %s", line)
	}
}
