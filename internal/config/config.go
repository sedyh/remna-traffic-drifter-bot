package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sedyh/remna-traffic-drifter-bot/internal/check"
)

type TelegramChat struct {
	ChatID   string
	ThreadID int
}

type Config struct {
	PanelURL             string
	PanelToken           string
	TelegramToken        string
	TelegramChats        []TelegramChat
	TelegramSendInterval time.Duration
	TelegramProxyURL        string
	PollInterval            time.Duration
	StaleAfter              time.Duration
	PageSize                int
	StatePath            string
	TelegramOffsetPath   string
	TagRules             map[string]check.TagRule
	WatchUntaggedLimited bool
	UntaggedRule         check.TagRule
}

func Load() (Config, error) {
	cfg := Config{
		PanelURL:             strings.TrimSuffix(strings.TrimSpace(os.Getenv("PANEL_URL")), "/"),
		PanelToken:           strings.TrimSpace(os.Getenv("PANEL_TOKEN")),
		TelegramToken:        strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN")),
		TelegramProxyURL:     strings.TrimSpace(os.Getenv("TELEGRAM_PROXY_URL")),
		StatePath:            envOr("STATE_PATH", "/data/state.json"),
		TelegramOffsetPath:   envOr("TELEGRAM_OFFSET_PATH", "/data/telegram_offset"),
		PageSize:             envIntOr("PAGE_SIZE", 500),
		WatchUntaggedLimited: envBoolOr("WATCH_UNTAGGED_LIMITED", true),
	}

	chatRaw := strings.TrimSpace(os.Getenv("TELEGRAM_CHAT_IDS"))
	if chatRaw == "" {
		return cfg, fmt.Errorf("TELEGRAM_CHAT_IDS is required")
	}
	chats, err := parseTelegramChats(chatRaw)
	if err != nil {
		return cfg, fmt.Errorf("TELEGRAM_CHAT_IDS: %w", err)
	}
	cfg.TelegramChats = chats

	poll, err := time.ParseDuration(envOr("POLL_INTERVAL", "15m"))
	if err != nil {
		return cfg, fmt.Errorf("POLL_INTERVAL: %w", err)
	}
	cfg.PollInterval = poll

	stale, err := parseDuration(envOr("STALE_RESET_AFTER", "35d"))
	if err != nil {
		return cfg, fmt.Errorf("STALE_RESET_AFTER: %w", err)
	}
	cfg.StaleAfter = stale

	sendInterval, err := time.ParseDuration(envOr("TELEGRAM_SEND_INTERVAL", "3s"))
	if err != nil {
		return cfg, fmt.Errorf("TELEGRAM_SEND_INTERVAL: %w", err)
	}
	cfg.TelegramSendInterval = sendInterval

	tagRulesRaw := strings.TrimSpace(os.Getenv("TAG_RULES"))
	if tagRulesRaw == "" {
		return cfg, fmt.Errorf("TAG_RULES is required")
	}
	tagRules, err := parseTagRules(tagRulesRaw)
	if err != nil {
		return cfg, fmt.Errorf("TAG_RULES: %w", err)
	}
	cfg.TagRules = tagRules

	if cfg.WatchUntaggedLimited {
		untaggedRaw := envOr("UNTAGGED_LIMITED_RULE", "MONTH")
		untaggedRule, err := parseUntaggedRule(untaggedRaw)
		if err != nil {
			return cfg, fmt.Errorf("UNTAGGED_LIMITED_RULE: %w", err)
		}
		cfg.UntaggedRule = untaggedRule
	}

	if cfg.PanelURL == "" {
		return cfg, fmt.Errorf("PANEL_URL is required")
	}
	if cfg.PanelToken == "" {
		return cfg, fmt.Errorf("PANEL_TOKEN is required")
	}
	if cfg.TelegramToken == "" {
		return cfg, fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}
	return cfg, nil
}

func parseTelegramChats(raw string) ([]TelegramChat, error) {
	var chats []TelegramChat
	for _, entry := range strings.Split(raw, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		chat, err := parseTelegramChatEntry(entry)
		if err != nil {
			return nil, err
		}
		chats = append(chats, chat)
	}
	if len(chats) == 0 {
		return nil, fmt.Errorf("no chat ids")
	}
	return chats, nil
}

func parseTelegramChatEntry(entry string) (TelegramChat, error) {
	chatID := entry
	threadID := 0

	if i := strings.LastIndex(entry, ":"); i > 0 {
		suffix := strings.TrimSpace(entry[i+1:])
		if suffix != "" {
			n, err := strconv.Atoi(suffix)
			if err != nil {
				return TelegramChat{}, fmt.Errorf("invalid topic id in %q", entry)
			}
			if n <= 0 {
				return TelegramChat{}, fmt.Errorf("topic id must be positive in %q", entry)
			}
			chatID = strings.TrimSpace(entry[:i])
			threadID = n
		}
	}

	chatID = strings.TrimSpace(chatID)
	if chatID == "" {
		return TelegramChat{}, fmt.Errorf("empty chat id in %q", entry)
	}

	return TelegramChat{ChatID: chatID, ThreadID: threadID}, nil
}

func (c Config) TelegramAllowedChatIDs() map[string]struct{} {
	allowed := make(map[string]struct{}, len(c.TelegramChats))
	for _, chat := range c.TelegramChats {
		allowed[chat.ChatID] = struct{}{}
	}
	return allowed
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func envIntOr(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func parseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty duration")
	}
	if strings.HasSuffix(s, "d") {
		n, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
		if err != nil || n < 0 {
			return 0, fmt.Errorf("invalid days: %q", s)
		}
		return time.Duration(n) * 24 * time.Hour, nil
	}
	return time.ParseDuration(s)
}

func envBoolOr(key string, fallback bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	switch strings.ToLower(v) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}
