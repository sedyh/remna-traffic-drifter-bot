package check

import (
	"testing"
	"time"

	"github.com/sedyh/remna-traffic-drifter-bot/internal/panel"
)

const testWatchedTag = "TAG_TARIFF_MIN"

func testOptions(now time.Time, mutate func(*Options)) Options {
	opt := Options{
		TagRules: map[string]TagRule{
			testWatchedTag: {Strategy: StrategyMonth, StaleAfter: 35 * 24 * time.Hour},
		},
		GlobalStaleAfter: 35 * 24 * time.Hour,
		Now:              now,
	}
	if mutate != nil {
		mutate(&opt)
	}
	return opt
}

func trafficUsed(bytes int64) *panel.UserTraffic {
	return &panel.UserTraffic{UsedTrafficBytes: bytes}
}

func TestExpectedStrategyFromTagRule(t *testing.T) {
	opt := testOptions(time.Now().UTC(), func(o *Options) {
		o.TagRules["MAX"] = TagRule{Strategy: StrategyNoReset}
	})
	if ExpectedStrategy(panel.User{Tag: testWatchedTag}, opt) != StrategyMonth {
		t.Fatal("TAG_TARIFF_MIN should be MONTH")
	}
	if ExpectedStrategy(panel.User{Tag: "MAX"}, opt) != StrategyNoReset {
		t.Fatal("MAX should be NO_RESET")
	}
}

func TestStaleAfterForUsesPerTagThreshold(t *testing.T) {
	opt := testOptions(time.Now().UTC(), func(o *Options) {
		o.TagRules[testWatchedTag] = TagRule{Strategy: StrategyMonth, StaleAfter: 10 * 24 * time.Hour}
		o.GlobalStaleAfter = 35 * 24 * time.Hour
	})
	u := panel.User{Tag: testWatchedTag}
	if StaleAfterFor(u, opt) != 10*24*time.Hour {
		t.Fatalf("want per-tag stale, got %v", StaleAfterFor(u, opt))
	}
}

func TestRunWrongStrategy(t *testing.T) {
	now := time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC)
	users := []panel.User{{
		UUID:                 "uuid-1",
		Username:             "VPN-1",
		Status:               "ACTIVE",
		Tag:                  testWatchedTag,
		TrafficLimitBytes:    100 * 1024 * 1024 * 1024,
		TrafficLimitStrategy: "NO_RESET",
		UserTraffic:          trafficUsed(1024),
	}}

	issues := Run(users, testOptions(now, nil))

	if len(issues) != 1 || issues[0].Kind != IssueWrongStrategy {
		t.Fatalf("unexpected issues: %+v", issues)
	}
	if issues[0].ExpectedStrategy != StrategyMonth {
		t.Fatalf("expected MONTH, got %s", issues[0].ExpectedStrategy)
	}
}

func TestRunStaleResetOldAccount(t *testing.T) {
	now := time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC)
	created := now.Add(-43 * 24 * time.Hour)
	users := []panel.User{{
		UUID:                 "uuid-2",
		Username:             "VPN-662044",
		Status:               "ACTIVE",
		Tag:                  testWatchedTag,
		TrafficLimitBytes:    100 * 1024 * 1024 * 1024,
		TrafficLimitStrategy: "MONTH",
		CreatedAt:            &created,
		LastTrafficResetAt:   nil,
		UserTraffic:          trafficUsed(1024),
	}}

	issues := Run(users, testOptions(now, nil))

	if len(issues) != 1 || issues[0].Kind != IssueStaleReset {
		t.Fatalf("unexpected issues: %+v", issues)
	}
}

func TestRunStaleResetSkipsNewAccount(t *testing.T) {
	now := time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC)
	created := now.Add(-8 * 24 * time.Hour)
	users := []panel.User{{
		Username:             "VPN-new",
		Tag:                  "",
		TrafficLimitBytes:    100 * 1024 * 1024 * 1024,
		TrafficLimitStrategy: "MONTH",
		CreatedAt:            &created,
		LastTrafficResetAt:   nil,
		UserTraffic:          trafficUsed(1024),
	}}

	issues := Run(users, testOptions(now, func(o *Options) {
		o.WatchUntaggedLimited = true
		o.UntaggedRule = TagRule{Strategy: StrategyMonth, StaleAfter: 35 * 24 * time.Hour}
	}))

	if len(issues) != 0 {
		t.Fatalf("want no issues for new account, got %+v", issues)
	}
}

func TestRunSkipsUnlimited(t *testing.T) {
	now := time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC)
	users := []panel.User{{
		Username:             "VPN-unlimited",
		Tag:                  testWatchedTag,
		TrafficLimitBytes:    0,
		TrafficLimitStrategy: "NO_RESET",
		UserTraffic:          trafficUsed(1 << 30),
	}}

	issues := Run(users, testOptions(now, nil))

	if len(issues) != 0 {
		t.Fatalf("want no issues for unlimited, got %+v", issues)
	}
}

func TestRunSkipsExpired(t *testing.T) {
	now := time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC)
	users := []panel.User{{
		Username:             "VPN-expired",
		Status:               "EXPIRED",
		Tag:                  "",
		TrafficLimitBytes:    100 * 1024 * 1024 * 1024,
		TrafficLimitStrategy: "NO_RESET",
		UserTraffic:          trafficUsed(1 << 30),
	}}

	issues := Run(users, testOptions(now, func(o *Options) {
		o.WatchUntaggedLimited = true
		o.UntaggedRule = TagRule{Strategy: StrategyMonth}
	}))

	if len(issues) != 0 {
		t.Fatalf("want no issues for EXPIRED, got %+v", issues)
	}
}

func TestRunSkipsTrialTag(t *testing.T) {
	now := time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC)
	created := now.Add(-60 * 24 * time.Hour)
	users := []panel.User{{
		Username:             "VPN-trial",
		Tag:                  "TRIAL",
		TrafficLimitBytes:    100 * 1024 * 1024 * 1024,
		TrafficLimitStrategy: "NO_RESET",
		CreatedAt:            &created,
		LastTrafficResetAt:   nil,
		UserTraffic:          trafficUsed(1 << 30),
	}}

	issues := Run(users, testOptions(now, nil))

	if len(issues) != 0 {
		t.Fatalf("want no issues for TRIAL, got %+v", issues)
	}
}
