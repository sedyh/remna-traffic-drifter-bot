package runner

import (
	"github.com/sedyh/remna-traffic-drifter-bot/internal/config"
	"github.com/sedyh/remna-traffic-drifter-bot/internal/notify"
)

func telegramChatsFromConfig(chats []config.TelegramChat) []notify.TelegramChat {
	out := make([]notify.TelegramChat, len(chats))
	for i, c := range chats {
		out[i] = notify.TelegramChat{ChatID: c.ChatID, ThreadID: c.ThreadID}
	}
	return out
}
