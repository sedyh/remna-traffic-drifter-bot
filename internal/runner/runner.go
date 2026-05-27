package runner

import (
	"context"
	"log/slog"
	"time"

	"github.com/sedyh/remna-traffic-drifter-bot/internal/check"
	"github.com/sedyh/remna-traffic-drifter-bot/internal/config"
	"github.com/sedyh/remna-traffic-drifter-bot/internal/notify"
	"github.com/sedyh/remna-traffic-drifter-bot/internal/panel"
	"github.com/sedyh/remna-traffic-drifter-bot/internal/state"
)

const alertSendTimeout = 60 * time.Second

type Runner struct {
	cfg    config.Config
	panel  *panel.Client
	notify *notify.Telegram
	store  *state.Store
	log    *slog.Logger
}

func New(cfg config.Config, log *slog.Logger) (*Runner, error) {
	store, err := state.Open(cfg.StatePath)
	if err != nil {
		return nil, err
	}
	return &Runner{
		cfg:    cfg,
		panel:  panel.NewClient(cfg.PanelURL, cfg.PanelToken),
		notify: notify.NewTelegram(
			cfg.TelegramToken,
			telegramChatsFromConfig(cfg.TelegramChats),
			cfg.TelegramSendInterval,
			cfg.TelegramProxyURL,
			cfg.PanelURL,
		),
		store:  store,
		log:    log,
	}, nil
}

func (r *Runner) RunOnce(ctx context.Context) error {
	users, err := r.panel.ListAllUsers(ctx, r.cfg.PageSize)
	if err != nil {
		return err
	}

	issues := check.Run(users, r.checkOptions(time.Now().UTC()))

	active := make(map[string]struct{}, len(issues))
	for _, issue := range issues {
		active[state.Key(issue.Username, issue.Kind)] = struct{}{}
	}
	r.store.SyncActive(active)

	sent := 0
	for _, issue := range issues {
		if !r.store.ShouldNotify(issue.Username, issue.Kind) {
			continue
		}
		sendCtx, cancel := r.alertContext(ctx)
		err := r.notify.SendIssue(sendCtx, issue)
		cancel()
		if err != nil {
			r.log.Error("telegram send failed", "user", issue.Username, "kind", issue.Kind, "err", err)
			continue
		}
		r.store.MarkNotified(issue.Username, issue.Kind)
		sent++
		r.log.Info("alert sent", "user", issue.Username, "kind", issue.Kind)
	}

	if err := r.store.Save(); err != nil {
		return err
	}

	r.log.Info(
		"poll done",
		"users_total", len(users),
		"issues", len(issues),
		"alerts_sent", sent,
	)
	return nil
}

func (r *Runner) alertContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx.Err() == nil {
		return context.WithTimeout(ctx, alertSendTimeout)
	}
	return context.WithTimeout(context.Background(), alertSendTimeout)
}

func (r *Runner) Loop(ctx context.Context) {
	ticker := time.NewTicker(r.cfg.PollInterval)
	defer ticker.Stop()

	r.log.Info(
		"traffic-drifter started",
		"poll_interval", r.cfg.PollInterval,
		"stale_after", r.cfg.StaleAfter,
	)

	go r.RunTelegramCallbacks(ctx)

	if err := r.RunOnce(ctx); err != nil {
		r.log.Error("poll failed", "err", err)
	}

	for {
		select {
		case <-ctx.Done():
			r.log.Info("traffic-drifter stopped")
			return
		case <-ticker.C:
			if err := r.RunOnce(ctx); err != nil {
				r.log.Error("poll failed", "err", err)
			}
		}
	}
}
