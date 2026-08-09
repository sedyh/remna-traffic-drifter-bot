package check

import (
	"strings"
	"time"

	"github.com/sedyh/remna-traffic-drifter-bot/internal/panel"
)

const (
	IssueWrongStrategy = "wrong_strategy"
	IssueStaleReset    = "stale_reset"

	StrategyMonth   = "MONTH"
	StrategyNoReset = "NO_RESET"
)

type Issue struct {
	Kind             string
	UserID           int64
	Username         string
	Tag              string
	Status           string
	Squad            string
	TrafficLimit     int64
	ActualStrategy   string
	ExpectedStrategy string
	LastResetAt      *time.Time
	StaleAfter       time.Duration
	CheckedAt        time.Time
}

func Run(users []panel.User, opt Options) []Issue {
	now := opt.Now.UTC()
	if now.IsZero() {
		now = time.Now().UTC()
	}

	var issues []Issue
	for _, u := range users {
		if skipStatus(u.Status) {
			continue
		}

		if !shouldWatch(u, opt) {
			continue
		}

		if u.SkipUnlimited() {
			continue
		}

		tag := formatTag(u.Tag)
		strategy := strings.TrimSpace(u.TrafficLimitStrategy)
		expected := ExpectedStrategy(u, opt)
		if expected == "" {
			continue
		}
		staleAfter := StaleAfterFor(u, opt)
		meta := issueMeta(u)

		if strategy != expected {
			issues = append(issues, Issue{
				Kind:             IssueWrongStrategy,
				UserID:           u.ID,
				Username:         u.Username,
				Tag:              tag,
				Status:           meta.status,
				Squad:            meta.squad,
				TrafficLimit:     meta.limit,
				ActualStrategy:   strategy,
				ExpectedStrategy: expected,
				CheckedAt:        now,
			})
			continue
		}

		if strategy != StrategyMonth {
			continue
		}

		if !u.HasUsedTraffic() {
			continue
		}

		if u.LastTrafficResetAt == nil {
			if u.CreatedAt != nil && now.Sub(u.CreatedAt.UTC()) <= staleAfter {
				continue
			}
			issues = append(issues, Issue{
				Kind:             IssueStaleReset,
				UserID:           u.ID,
				Username:         u.Username,
				Tag:              tag,
				Status:           meta.status,
				Squad:            meta.squad,
				TrafficLimit:     meta.limit,
				ActualStrategy:   strategy,
				ExpectedStrategy: expected,
				StaleAfter:       staleAfter,
				CheckedAt:        now,
			})
			continue
		}

		lastReset := u.LastTrafficResetAt.UTC()
		if now.Sub(lastReset) > staleAfter {
			reset := lastReset
			issues = append(issues, Issue{
				Kind:             IssueStaleReset,
				UserID:           u.ID,
				Username:         u.Username,
				Tag:              tag,
				Status:           meta.status,
				Squad:            meta.squad,
				TrafficLimit:     meta.limit,
				ActualStrategy:   strategy,
				ExpectedStrategy: expected,
				LastResetAt:      &reset,
				StaleAfter:       staleAfter,
				CheckedAt:        now,
			})
		}
	}
	return issues
}

func skipStatus(status string) bool {
	return strings.EqualFold(strings.TrimSpace(status), "EXPIRED")
}

func shouldWatch(u panel.User, opt Options) bool {
	tag := normalizeTag(u.Tag)
	if tag != "" {
		_, ok := opt.TagRules[tag]
		return ok
	}
	return opt.WatchUntaggedLimited && u.TrafficLimitBytes > 0
}

type issueMetaFields struct {
	status string
	squad  string
	limit  int64
}

func issueMeta(u panel.User) issueMetaFields {
	return issueMetaFields{
		status: strings.TrimSpace(u.Status),
		squad:  u.SquadNames(),
		limit:  u.TrafficLimitBytes,
	}
}

func formatTag(tag string) string {
	tag = normalizeTag(tag)
	if tag == "" {
		return "No tariff"
	}
	return tag
}
