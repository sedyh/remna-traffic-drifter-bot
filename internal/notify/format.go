package notify

import (
	"fmt"
	"strings"
	"time"

	"github.com/sedyh/remna-traffic-drifter-bot/internal/panel"
)


type issueFields struct {
	username         string
	uuid             string
	panelBase        string
	tag              string
	status           string
	squad            string
	trafficLimit     int64
	actualStrategy   string
	expectedStrategy string
	lastResetAt      *time.Time
	checkedAt        time.Time
}

func FormatMessageAfterFix(original string, u panel.User, expectedStrategy string, staleAfter time.Duration, panelBase string) string {
	now := time.Now().UTC()
	current := formatPanelSnapshot(u, expectedStrategy, staleAfter, now, panelBase)
	if strings.TrimSpace(original) == "" {
		return current
	}
	return strikeHTML(original) + "\n\n" + current
}

func formatPanelSnapshot(u panel.User, expectedStrategy string, staleAfter time.Duration, now time.Time, panelBase string) string {
	strategy := strings.TrimSpace(u.TrafficLimitStrategy)
	lines := []string{
		"✅ Panel now",
		formatSubLine(u.Username, u.UUID, panelBase),
		"🧩 Strategy: " + escapeHTML(formatStrategy(strategy)),
		"🧩 Expected: " + escapeHTML(formatStrategy(expectedStrategy)),
	}

	if u.UserTraffic != nil {
		lines = append(lines, "🧩 Traffic: "+escapeHTML(formatBytes(u.UserTraffic.UsedTrafficBytes)))
	} else {
		lines = append(lines, "🧩 Traffic: 0 B")
	}

	if u.LastTrafficResetAt == nil {
		lines = append(lines, "🧩 Last reset: Never reset")
	} else {
		reset := u.LastTrafficResetAt.UTC()
		age := now.Sub(reset)
		lines = append(lines, "🧩 Last reset: "+escapeHTML(humanAgo(age)+", at "+formatWhen(reset)))
	}

	if strategy == expectedStrategy {
		if u.LastTrafficResetAt == nil {
			if u.HasUsedTraffic() {
				lines = append(lines, "🧩 Drift: stale (Never reset)")
			}
		} else if now.Sub(u.LastTrafficResetAt.UTC()) > staleAfter {
			if u.HasUsedTraffic() {
				lines = append(lines, "🧩 Drift: stale reset")
			}
		} else {
			lines = append(lines, "🧩 Drift: none")
		}
	} else {
		lines = append(lines, "🧩 Drift: wrong strategy")
	}

	return strings.Join(lines, "\n")
}

func strikeHTML(text string) string {
	lines := strings.Split(text, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		out = append(out, "<s>"+escapeHTML(line)+"</s>")
	}
	return strings.Join(out, "\n")
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}
