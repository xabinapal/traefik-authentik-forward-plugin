package authentik

import (
	"net/http"
	"strings"

	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/httputil"
)

const (
	HeaderPrefix = "X-Authentik-"
	CookiePrefix = "authentik_proxy_"
)

func GetHeaders(res *http.Response) http.Header {
	headers := http.Header{}
	for k, v := range res.Header {
		if strings.HasPrefix(k, HeaderPrefix) && k != HeaderPrefix {
			headers[k] = v
		}
	}

	return headers
}

func GetCookies(cookier httputil.Cookier) []*http.Cookie {
	cookies := make([]*http.Cookie, 0, 1)

	for _, c := range cookier.Cookies() {
		if strings.HasPrefix(c.Name, CookiePrefix) && c.Name != CookiePrefix {
			cookies = append(cookies, c)
		}
	}

	return cookies
}

func RequestMangle(req *http.Request) {
	// remove downstream authentik headers
	for k := range req.Header {
		if strings.HasPrefix(k, HeaderPrefix) {
			delete(req.Header, k)
		}
	}

	// remove downstream authentik cookies
	cookies := req.Cookies()
	req.Header.Del("Cookie")
	for _, c := range cookies {
		if !strings.HasPrefix(c.Name, CookiePrefix) {
			req.AddCookie(c)
		}
	}
}

func GetResponseMangler(cookies []*http.Cookie) func(rw http.ResponseWriter) {
	return func(rw http.ResponseWriter) {
		// remove upstream authentik headers
		for k := range rw.Header() {
			if strings.HasPrefix(k, HeaderPrefix) {
				delete(rw.Header(), k)
			}
		}

		// remove upstream authentik cookies
		upCookies := rw.Header().Values("Set-Cookie")
		rw.Header().Del("Set-Cookie")

		for _, c := range upCookies {
			name := httputil.ParseCookieName(c)
			if name == "" || strings.HasPrefix(name, CookiePrefix) {
				continue
			}

			rw.Header().Add("Set-Cookie", c)
		}

		// add forward authentik cookies
		for _, c := range cookies {
			rw.Header().Add("Set-Cookie", c.String())
		}
	}
}
