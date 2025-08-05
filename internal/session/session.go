package session

import "net/http"

type Session struct {
	IsAuthenticated bool
	Headers         http.Header
	Cookies         []*http.Cookie
}
