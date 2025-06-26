package httputil

import (
	"net/http"
	"strings"
)

type Cookier interface {
	Cookies() []*http.Cookie
}

func parseCookieName(c string) string {
	parts := strings.Split(c, "=")
	if len(parts) < 2 {
		return ""
	}

	return parts[0]
}
