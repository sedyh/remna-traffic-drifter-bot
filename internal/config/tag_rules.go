package config

import (
	"fmt"
	"strings"

	"github.com/sedyh/remna-traffic-drifter-bot/internal/check"
)

func parseTagRules(raw string) (map[string]check.TagRule, error) {
	rules := make(map[string]check.TagRule)
	for _, entry := range strings.Split(raw, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		tag, rule, err := parseTagRuleEntry(entry)
		if err != nil {
			return nil, err
		}
		rules[tag] = rule
	}
	if len(rules) == 0 {
		return nil, fmt.Errorf("no tag rules")
	}
	return rules, nil
}

func parseTagRuleEntry(entry string) (string, check.TagRule, error) {
	parts := strings.Split(entry, ":")
	if len(parts) < 2 || len(parts) > 3 {
		return "", check.TagRule{}, fmt.Errorf("invalid entry %q, want TAG:STRATEGY or TAG:STRATEGY:STALE", entry)
	}

	tag := strings.ToUpper(strings.TrimSpace(parts[0]))
	if tag == "" {
		return "", check.TagRule{}, fmt.Errorf("empty tag in %q", entry)
	}

	strategy := strings.ToUpper(strings.TrimSpace(parts[1]))
	if err := validateStrategy(strategy); err != nil {
		return "", check.TagRule{}, fmt.Errorf("%q: %w", entry, err)
	}

	rule := check.TagRule{Strategy: strategy}
	if len(parts) == 3 {
		staleRaw := strings.TrimSpace(parts[2])
		if staleRaw == "" {
			return "", check.TagRule{}, fmt.Errorf("empty stale threshold in %q", entry)
		}
		stale, err := parseDuration(staleRaw)
		if err != nil {
			return "", check.TagRule{}, fmt.Errorf("%q: stale: %w", entry, err)
		}
		if stale <= 0 {
			return "", check.TagRule{}, fmt.Errorf("stale threshold must be positive in %q", entry)
		}
		rule.StaleAfter = stale
	}

	return tag, rule, nil
}

func parseUntaggedRule(raw string) (check.TagRule, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return check.TagRule{}, fmt.Errorf("empty untagged rule")
	}
	parts := strings.Split(raw, ":")
	if len(parts) < 1 || len(parts) > 2 {
		return check.TagRule{}, fmt.Errorf("invalid untagged rule %q, want STRATEGY or STRATEGY:STALE", raw)
	}

	strategy := strings.ToUpper(strings.TrimSpace(parts[0]))
	if err := validateStrategy(strategy); err != nil {
		return check.TagRule{}, fmt.Errorf("%q: %w", raw, err)
	}

	rule := check.TagRule{Strategy: strategy}
	if len(parts) == 2 {
		stale, err := parseDuration(strings.TrimSpace(parts[1]))
		if err != nil {
			return check.TagRule{}, fmt.Errorf("%q: stale: %w", raw, err)
		}
		if stale <= 0 {
			return check.TagRule{}, fmt.Errorf("stale threshold must be positive in %q", raw)
		}
		rule.StaleAfter = stale
	}

	return rule, nil
}

func validateStrategy(strategy string) error {
	switch strategy {
	case check.StrategyMonth, check.StrategyNoReset:
		return nil
	default:
		return fmt.Errorf("unknown strategy %q", strategy)
	}
}
