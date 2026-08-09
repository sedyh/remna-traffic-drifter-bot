package notify

import (
	"testing"

	"github.com/sedyh/remna-traffic-drifter-bot/internal/check"
)

func TestIssueKeyboardSingleFix(t *testing.T) {
	kb := issueKeyboard(check.Issue{
		Kind:             check.IssueWrongStrategy,
		UserID:           42,
		ExpectedStrategy: "MONTH",
	})
	if kb == nil || len(kb.InlineKeyboard[0]) != 1 {
		t.Fatalf("unexpected keyboard: %+v", kb)
	}
	if kb.InlineKeyboard[0][0].Text != "Fix → MONTH" {
		t.Fatalf("unexpected label: %s", kb.InlineKeyboard[0][0].Text)
	}
	if kb.InlineKeyboard[0][0].CallbackData != "f:42" {
		t.Fatalf("unexpected callback: %s", kb.InlineKeyboard[0][0].CallbackData)
	}
}

func TestDecodeCallback(t *testing.T) {
	action, userID, err := DecodeCallback("f:42")
	if err != nil || action != CallbackFix || userID != 42 {
		t.Fatalf("unexpected: %s %d %v", action, userID, err)
	}
}
