package notify

import (
	"testing"

	"github.com/sedyh/remna-traffic-drifter-bot/internal/check"
)

func TestIssueKeyboardSingleFix(t *testing.T) {
	kb := issueKeyboard(check.Issue{
		Kind:             check.IssueWrongStrategy,
		UUID:             "550e8400-e29b-41d4-a716-446655440000",
		ExpectedStrategy: "MONTH",
	})
	if kb == nil || len(kb.InlineKeyboard[0]) != 1 {
		t.Fatalf("unexpected keyboard: %+v", kb)
	}
	if kb.InlineKeyboard[0][0].Text != "Fix → MONTH" {
		t.Fatalf("unexpected label: %s", kb.InlineKeyboard[0][0].Text)
	}
}

func TestDecodeCallback(t *testing.T) {
	action, uuid, err := DecodeCallback("f:550e8400-e29b-41d4-a716-446655440000")
	if err != nil || action != CallbackFix || uuid == "" {
		t.Fatalf("unexpected: %s %s %v", action, uuid, err)
	}
}
