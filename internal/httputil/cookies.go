package httputil

import (
	"net/http"
	"strings"
)

type Cookier interface {
	Cookies() []*http.Cookie
}

//nolint:mnd
func parseCookieName(c string) string {
	parts := strings.SplitN(c, "=", 2)
	if len(parts) != 2 {
		return ""
	}

	return parts[0]
}
