package check

import (
	"strings"
	"time"

	"github.com/sedyh/remna-traffic-drifter-bot/internal/panel"
)

type TagRule struct {
	Strategy   string
	StaleAfter time.Duration
}

type Options struct {
	TagRules             map[string]TagRule
	GlobalStaleAfter     time.Duration
	WatchUntaggedLimited bool
	UntaggedRule         TagRule
	Now                  time.Time
}

func ExpectedStrategy(u panel.User, opt Options) string {
	tag := normalizeTag(u.Tag)
	if tag != "" {
		if rule, ok := opt.TagRules[tag]; ok {
			return rule.Strategy
		}
		return ""
	}
	if opt.WatchUntaggedLimited && u.TrafficLimitBytes > 0 {
		return opt.UntaggedRule.Strategy
	}
	return ""
}

func StaleAfterFor(u panel.User, opt Options) time.Duration {
	tag := normalizeTag(u.Tag)
	if tag != "" {
		if rule, ok := opt.TagRules[tag]; ok && rule.StaleAfter > 0 {
			return rule.StaleAfter
		}
	} else if opt.WatchUntaggedLimited && u.TrafficLimitBytes > 0 && opt.UntaggedRule.StaleAfter > 0 {
		return opt.UntaggedRule.StaleAfter
	}
	return opt.GlobalStaleAfter
}

func normalizeTag(tag string) string {
	return strings.ToUpper(strings.TrimSpace(tag))
}
