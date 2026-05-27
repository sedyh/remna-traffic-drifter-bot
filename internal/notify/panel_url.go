package notify

import "strings"

func UserAdminURL(panelBase, uuid string) string {
	base := strings.TrimSuffix(strings.TrimSpace(panelBase), "/")
	uuid = strings.TrimSpace(uuid)
	if base == "" || uuid == "" {
		return ""
	}
	return base + "/dashboard/management/users?user=" + uuid
}

func formatSubLine(username, uuid, panelBase string) string {
	if href := UserAdminURL(panelBase, uuid); href != "" {
		return `💳 Sub: <a href="` + escapeHTML(href) + `">` + escapeHTML(username) + `</a>`
	}
	return "💳 Sub: " + escapeHTML(username)
}
