package httputil

import (
	"net/http"
	"strings"
)

type Cookier interface {
	Cookies() []*http.Cookie
}

func ParseCookieName(c string) string {
	eqIndex := strings.IndexByte(c, '=')
	if eqIndex == -1 {
		return ""
	}

	name := strings.TrimSpace(c[:eqIndex])
	return name
}
