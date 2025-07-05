package authentik

import (
	"net/http"
	"net/url"
)

type RequestMeta struct {
	URL     *url.URL
	Cookies []*http.Cookie
}

type ResponseMeta struct {
	URL             *url.URL
	IsAuthenticated bool
	Headers         http.Header
	Cookies         []*http.Cookie
}
