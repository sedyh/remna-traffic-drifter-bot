package config

import (
	"testing"
	"time"
)

func TestParseDurationDays(t *testing.T) {
	d, err := parseDuration("35d")
	if err != nil {
		t.Fatal(err)
	}
	if d != 35*24*time.Hour {
		t.Fatalf("want 35d, got %v", d)
	}
}

func TestParseDurationStd(t *testing.T) {
	d, err := parseDuration("15m")
	if err != nil || d != 15*time.Minute {
		t.Fatalf("want 15m, got %v err=%v", d, err)
	}
}

func TestParseTelegramChats(t *testing.T) {
	chats, err := parseTelegramChats("-1001:289, -1002")
	if err != nil {
		t.Fatal(err)
	}
	if len(chats) != 2 {
		t.Fatalf("want 2 chats, got %d", len(chats))
	}
	if chats[0].ChatID != "-1001" || chats[0].ThreadID != 289 {
		t.Fatalf("first chat: %+v", chats[0])
	}
	if chats[1].ChatID != "-1002" || chats[1].ThreadID != 0 {
		t.Fatalf("second chat: %+v", chats[1])
	}
}

func TestParseTelegramChatsInvalidTopic(t *testing.T) {
	_, err := parseTelegramChats("-1001:abc")
	if err == nil {
		t.Fatal("want error")
	}
}

func TestParseTagRules(t *testing.T) {
	rules, err := parseTagRules("TAG_TARIFF_MIN:MONTH:35d,TAG_TARRIF_PRO:MONTH,MAX:NO_RESET")
	if err != nil {
		t.Fatal(err)
	}
	if rules["TAG_TARIFF_MIN"].Strategy != "MONTH" || rules["TAG_TARIFF_MIN"].StaleAfter != 35*24*time.Hour {
		t.Fatalf("TAG_TARIFF_MIN: %+v", rules["TAG_TARIFF_MIN"])
	}
	if rules["TAG_TARRIF_PRO"].Strategy != "MONTH" || rules["TAG_TARRIF_PRO"].StaleAfter != 0 {
		t.Fatalf("TAG_TARRIF_PRO: %+v", rules["TAG_TARRIF_PRO"])
	}
	if rules["MAX"].Strategy != "NO_RESET" {
		t.Fatalf("MAX: %+v", rules["MAX"])
	}
}

func TestParseUntaggedRule(t *testing.T) {
	rule, err := parseUntaggedRule("MONTH:720h")
	if err != nil {
		t.Fatal(err)
	}
	if rule.Strategy != "MONTH" || rule.StaleAfter != 720*time.Hour {
		t.Fatalf("unexpected: %+v", rule)
	}
}
