package runner

import (
	"time"

	"github.com/sedyh/remna-traffic-drifter-bot/internal/check"
	"github.com/sedyh/remna-traffic-drifter-bot/internal/panel"
)

func (r *Runner) checkOptions(now time.Time) check.Options {
	return check.Options{
		TagRules:             r.cfg.TagRules,
		GlobalStaleAfter:     r.cfg.StaleAfter,
		WatchUntaggedLimited: r.cfg.WatchUntaggedLimited,
		UntaggedRule:         r.cfg.UntaggedRule,
		Now:                  now,
	}
}

func (r *Runner) expectedStrategy(u panel.User) string {
	return check.ExpectedStrategy(u, r.checkOptions(time.Now().UTC()))
}

func (r *Runner) staleAfterFor(u panel.User) time.Duration {
	return check.StaleAfterFor(u, r.checkOptions(time.Now().UTC()))
}
