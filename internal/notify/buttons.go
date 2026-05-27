package notify

import (
	"fmt"
	"strings"

	"github.com/sedyh/remna-traffic-drifter-bot/internal/check"
)

const CallbackFix = "f"

type inlineButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
}

type inlineKeyboard struct {
	InlineKeyboard [][]inlineButton `json:"inline_keyboard"`
}

func EncodeCallback(action, uuid string) string {
	return action + ":" + uuid
}

func DecodeCallback(data string) (action, uuid string, err error) {
	parts := strings.SplitN(data, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid callback data")
	}
	return parts[0], parts[1], nil
}

func issueKeyboard(i check.Issue) *inlineKeyboard {
	if strings.TrimSpace(i.UUID) == "" {
		return nil
	}
	switch i.Kind {
	case check.IssueWrongStrategy, check.IssueStaleReset:
		return &inlineKeyboard{InlineKeyboard: [][]inlineButton{{
			{
				Text:         "Fix → " + i.ExpectedStrategy,
				CallbackData: EncodeCallback(CallbackFix, i.UUID),
			},
		}}}
	default:
		return nil
	}
}
