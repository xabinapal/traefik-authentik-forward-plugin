package authentik

import (
	"net/http"
	"net/url"

	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/session"
)

type RequestMeta struct {
	URL     *url.URL
	Cookies []*http.Cookie
}

type ResponseMeta struct {
	URL     *url.URL
	Session *session.Session
}
