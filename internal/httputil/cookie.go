package httputil

import (
	"net/http"
	"strings"
)

type Cookier interface {
	Cookies() []*http.Cookie
}

func ParseCookieName(c string) string {
	index := strings.IndexByte(c, '=')
	if index == -1 {
		// return empty string if cookie format is not name=value
		return ""
	}

	name := strings.TrimSpace(c[:index])
	return name
}
