package notify

import (
	"strings"
	"testing"
)

func TestUserAdminURL(t *testing.T) {
	got := UserAdminURL("https://panel.example.com/", 42)
	want := "https://panel.example.com/dashboard/management/users?user=42"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestFormatSubLineLink(t *testing.T) {
	line := formatSubLine("ICE-1", 42, "https://panel.example.com")
	want := `href="https://panel.example.com/dashboard/management/users?user=42"`
	if !strings.Contains(line, want) || !strings.Contains(line, ">ICE-1</a>") {
		t.Fatalf("unexpected: %s", line)
	}
}
