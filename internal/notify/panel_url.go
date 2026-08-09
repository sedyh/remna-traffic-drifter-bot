package notify

import (
	"strconv"
	"strings"
)

func UserAdminURL(panelBase string, userID int64) string {
	base := strings.TrimSuffix(strings.TrimSpace(panelBase), "/")
	if base == "" || userID == 0 {
		return ""
	}
	return base + "/dashboard/management/users?user=" + strconv.FormatInt(userID, 10)
}

func formatSubLine(username string, userID int64, panelBase string) string {
	if href := UserAdminURL(panelBase, userID); href != "" {
		return `💳 Sub: <a href="` + escapeHTML(href) + `">` + escapeHTML(username) + `</a>`
	}
	return "💳 Sub: " + escapeHTML(username)
}
