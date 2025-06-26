package authentik

import (
	"net/http"
	"strings"

	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/httputil"
)

const (
	CookiePrefix = "authentik_proxy_"
)

func GetCookies(req httputil.Cookier) []*http.Cookie {
	cookies := []*http.Cookie{}
	for _, c := range req.Cookies() {
		if strings.HasPrefix(c.Name, CookiePrefix) && c.Name != CookiePrefix {
			cookies = append(cookies, c)
		}
	}

	return cookies
}
