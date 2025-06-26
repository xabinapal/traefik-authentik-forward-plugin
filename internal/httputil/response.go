package httputil

import (
	"net/http"
	"strings"
)

type ResponseCookieModifier struct {
	http.ResponseWriter

	CookiesPrefix string
	Cookies       []*http.Cookie
}

func (rcm *ResponseCookieModifier) WriteHeader(code int) {
	cookies := rcm.ResponseWriter.Header().Values("Set-Cookie")
	rcm.ResponseWriter.Header().Del("Set-Cookie")

	for _, c := range cookies {
		name := parseCookieName(c)
		if strings.HasPrefix(name, rcm.CookiesPrefix) {
			continue
		}

		rcm.ResponseWriter.Header().Add("Set-Cookie", c)
	}

	for _, cookie := range rcm.Cookies {
		rcm.ResponseWriter.Header().Add("Set-Cookie", cookie.String())
	}

	rcm.ResponseWriter.WriteHeader(code)
}
