package ui

import "strings"

func cookieHasCSRF(cookie string) bool {
	return strings.Contains(cookie, "com.xk72.webparts.csrf=")
}

func cookieHasClearance(cookie string) bool {
	return strings.Contains(cookie, "cf_clearance=")
}

func cloudflareHint(err error) (string, bool) {
	if err == nil {
		return "", false
	}
	msg := err.Error()
	idx := strings.Index(msg, "Cloudflare challenge detected")
	if idx == -1 {
		return "", false
	}
	line := strings.TrimSpace(msg[idx:])
	if newline := strings.IndexByte(line, '\n'); newline != -1 {
		line = strings.TrimSpace(line[:newline])
	}
	if line == "" {
		return "", false
	}
	return line, true
}
