package notify

import (
	"fmt"
	"strconv"
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

func EncodeCallback(action string, userID int64) string {
	return action + ":" + strconv.FormatInt(userID, 10)
}

func DecodeCallback(data string) (action string, userID int64, err error) {
	parts := strings.SplitN(data, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", 0, fmt.Errorf("invalid callback data")
	}
	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || id <= 0 {
		return "", 0, fmt.Errorf("invalid callback user id")
	}
	return parts[0], id, nil
}

func issueKeyboard(i check.Issue) *inlineKeyboard {
	if i.UserID == 0 {
		return nil
	}
	switch i.Kind {
	case check.IssueWrongStrategy, check.IssueStaleReset:
		return &inlineKeyboard{InlineKeyboard: [][]inlineButton{{
			{
				Text:         "Fix → " + i.ExpectedStrategy,
				CallbackData: EncodeCallback(CallbackFix, i.UserID),
			},
		}}}
	default:
		return nil
	}
}
