package runner

import (
	"context"
	"strconv"
	"time"

	"github.com/sedyh/remna-traffic-drifter-bot/internal/notify"
	"github.com/sedyh/remna-traffic-drifter-bot/internal/state"
)

func (r *Runner) RunTelegramCallbacks(ctx context.Context) {
	offsetPath := r.cfg.TelegramOffsetPath
	offset, err := state.LoadOffset(offsetPath)
	if err != nil {
		r.log.Error("telegram offset load", "err", err)
	}

	if err := r.notify.DeleteWebhook(ctx); err != nil {
		r.log.Warn("telegram deleteWebhook", "err", err)
	}

	allowed := r.cfg.TelegramAllowedChatIDs()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		updates, err := r.notify.GetUpdates(ctx, offset, 30)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			r.log.Error("telegram getUpdates", "err", err)
			time.Sleep(3 * time.Second)
			continue
		}

		for _, upd := range updates {
			offset = upd.UpdateID + 1
			if upd.CallbackQuery == nil {
				continue
			}
			r.handleCallback(ctx, allowed, *upd.CallbackQuery)
		}

		if len(updates) > 0 {
			if err := state.SaveOffset(offsetPath, offset); err != nil {
				r.log.Error("telegram offset save", "err", err)
			}
		}
	}
}

func (r *Runner) handleCallback(ctx context.Context, allowed map[string]struct{}, cq notify.CallbackQuery) {
	chatID := strconv.FormatInt(cq.Message.Chat.ID, 10)
	if _, ok := allowed[chatID]; !ok {
		r.notify.AnswerCallback(ctx, cq.ID, "Chat not allowed")
		return
	}

	action, uuid, err := notify.DecodeCallback(cq.Data)
	if err != nil {
		r.notify.AnswerCallback(ctx, cq.ID, "Invalid button")
		return
	}

	var expected string
	switch action {
	case notify.CallbackFix, "s", "r":
		user, getErr := r.panel.GetUser(ctx, uuid)
		if getErr != nil {
			r.notify.AnswerCallback(ctx, cq.ID, "Panel error: "+truncateErr(getErr))
			return
		}
		expected = r.expectedStrategy(user)
		err = r.panel.FixUser(ctx, uuid, expected)
	default:
		r.notify.AnswerCallback(ctx, cq.ID, "Unknown action")
		return
	}

	done := expected + " + reset"

	if err != nil {
		r.log.Error("panel fix failed", "uuid", uuid, "action", action, "err", err)
		r.notify.AnswerCallback(ctx, cq.ID, "Panel error: "+truncateErr(err))
		return
	}

	user, err := r.panel.GetUser(ctx, uuid)
	if err != nil {
		r.notify.AnswerCallback(ctx, cq.ID, "✅ "+done+" (panel refresh failed)")
		r.log.Warn("panel get user after fix", "uuid", uuid, "err", err)
		return
	}

	r.store.ClearUser(user.Username)
	_ = r.store.Save()

	html := notify.FormatMessageAfterFix(cq.Message.Text, user, r.expectedStrategy(user), r.staleAfterFor(user), r.cfg.PanelURL)
	if err := r.notify.EditMessageAfterFix(ctx, chatID, cq.Message.MessageID, html); err != nil {
		r.log.Error("telegram edit message", "uuid", uuid, "err", err)
		r.notify.AnswerCallback(ctx, cq.ID, "✅ "+done+" (message edit failed)")
		return
	}

	r.notify.AnswerCallback(ctx, cq.ID, "✅ Fix: "+done)
	r.log.Info("panel fix applied", "uuid", uuid, "action", action, "user", user.Username)
}

func truncateErr(err error) string {
	msg := err.Error()
	if len(msg) > 120 {
		return msg[:120]
	}
	return msg
}
